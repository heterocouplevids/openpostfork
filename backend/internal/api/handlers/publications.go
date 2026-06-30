package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/uptrace/bun"
)

const publicationPathByID = "/publications/{id}"

type PublicationHandler struct {
	db   *bun.DB
	auth middleware.Authenticator
}

func NewPublicationHandler(db *bun.DB, authenticator middleware.Authenticator) *PublicationHandler {
	return &PublicationHandler{db: db, auth: authenticator}
}

type CreatePublicationInput struct {
	Body struct {
		WorkspaceID   string   `json:"workspace_id" doc:"Target workspace ID"`
		Title         string   `json:"title" minLength:"1" doc:"Short internal title"`
		SourceContent string   `json:"source_content" minLength:"1" doc:"Canonical idea, brief, announcement, notes, or source material"`
		SourceURL     string   `json:"source_url,omitempty" doc:"Optional source URL related to the publication"`
		Goal          string   `json:"goal,omitempty" doc:"Optional goal such as announce, explain, launch, ask for feedback, or promote article"`
		Audience      string   `json:"audience,omitempty" doc:"Optional intended audience"`
		MediaIDs      []string `json:"media_ids,omitempty" doc:"Workspace media IDs attached to this publication source"`
	}
}

type CreatePublicationOutput struct {
	Body *PublicationResponse
}

type ListPublicationsInput struct {
	WorkspaceID string `query:"workspace_id" required:"true" doc:"Filter by workspace ID"`
	Status      string `query:"status" doc:"Filter by publication status"`
	Limit       int    `query:"limit" doc:"Limit number of results (default 50, max 200)"`
	Offset      int    `query:"offset" doc:"Offset for pagination"`
}

type ListPublicationsOutput struct {
	Body []PublicationResponse
}

type GetPublicationInput struct {
	PathID string `path:"id" doc:"Publication ID"`
}

type GetPublicationOutput struct {
	Body *PublicationResponse
}

type UpdatePublicationInput struct {
	PathID string `path:"id" doc:"Publication ID"`
	Body   UpdatePublicationBody
}

type UpdatePublicationBody struct {
	Title         *string   `json:"title,omitempty" doc:"Short internal title"`
	SourceContent *string   `json:"source_content,omitempty" doc:"Canonical source material"`
	SourceURL     *string   `json:"source_url,omitempty" doc:"Optional source URL"`
	Goal          *string   `json:"goal,omitempty" doc:"Optional publication goal"`
	Audience      *string   `json:"audience,omitempty" doc:"Optional intended audience"`
	Status        *string   `json:"status,omitempty" doc:"Publication status"`
	MediaIDs      *[]string `json:"media_ids,omitempty" doc:"Replacement workspace media IDs"`
}

type UpdatePublicationOutput struct {
	Body *PublicationResponse
}

type PublicationResponse struct {
	ID              string   `json:"id" doc:"Publication ID"`
	WorkspaceID     string   `json:"workspace_id" doc:"Workspace ID"`
	CreatedByID     string   `json:"created_by" doc:"Creator user ID"`
	Title           string   `json:"title" doc:"Short internal title"`
	SourceContent   string   `json:"source_content" doc:"Canonical source material"`
	SourceURL       string   `json:"source_url,omitempty" doc:"Optional source URL"`
	Goal            string   `json:"goal,omitempty" doc:"Publication goal"`
	Audience        string   `json:"audience,omitempty" doc:"Intended audience"`
	Status          string   `json:"status" doc:"Publication status"`
	ReleasePlanJSON string   `json:"release_plan_json" doc:"Encoded release plan JSON"`
	MediaIDs        []string `json:"media_ids,omitempty" doc:"Attached workspace media IDs"`
	CreatedAt       string   `json:"created_at" doc:"Creation time"`
	UpdatedAt       string   `json:"updated_at" doc:"Last update time"`
}

func (h *PublicationHandler) checkWorkspaceAccess(ctx context.Context, workspaceID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return huma.Error400BadRequest(errWorkspaceIDRequired)
	}
	count, err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return huma.Error500InternalServerError(errValidateWorkspaceAccess)
	}
	if count == 0 {
		return huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	return nil
}

func (h *PublicationHandler) validateMediaBelongsToWorkspace(ctx context.Context, workspaceID string, mediaIDs []string) ([]string, error) {
	mediaIDs, err := normalizePublicationIDs(mediaIDs, "media_ids")
	if err != nil {
		return nil, err
	}
	if len(mediaIDs) == 0 {
		return mediaIDs, nil
	}
	count, err := h.db.NewSelect().
		Model((*models.MediaAttachment)(nil)).
		Where("workspace_id = ?", workspaceID).
		Where("id IN (?)", bun.List(mediaIDs)).
		Count(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to validate media attachments")
	}
	if count != len(mediaIDs) {
		return nil, huma.Error400BadRequest("one or more media attachments are invalid or outside this workspace")
	}
	return mediaIDs, nil
}

func normalizePublicationIDs(ids []string, field string) ([]string, error) {
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, huma.Error400BadRequest(field + " cannot contain empty values")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique, nil
}

func validatePublicationSourceURL(sourceURL string) error {
	if strings.TrimSpace(sourceURL) == "" {
		return nil
	}
	if _, err := url.ParseRequestURI(sourceURL); err != nil {
		return huma.Error400BadRequest("source_url must be a valid URI")
	}
	return nil
}

func validatePublicationStatus(status string) error {
	if status == "" {
		return nil
	}
	switch status {
	case models.PublicationStatusDraft, models.PublicationStatusReady, models.PublicationStatusScheduled, models.PublicationStatusPublished, models.PublicationStatusFailed:
		return nil
	default:
		return huma.Error400BadRequest("status must be one of draft, ready, scheduled, published, or failed")
	}
}

func (h *PublicationHandler) loadPublicationMediaIDs(ctx context.Context, publicationIDs []string) (map[string][]string, error) {
	mediaIDsByPublication := make(map[string][]string, len(publicationIDs))
	if len(publicationIDs) == 0 {
		return mediaIDsByPublication, nil
	}
	var rows []models.PublicationAsset
	err := h.db.NewSelect().
		Model(&rows).
		Where("publication_id IN (?)", bun.List(publicationIDs)).
		Order("publication_id ASC").
		Order("display_order ASC").
		Scan(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, huma.Error500InternalServerError("failed to load publication media")
	}
	for _, row := range rows {
		mediaIDsByPublication[row.PublicationID] = append(mediaIDsByPublication[row.PublicationID], row.MediaID)
	}
	return mediaIDsByPublication, nil
}

func (h *PublicationHandler) loadPublicationResponse(ctx context.Context, publicationID string) (*PublicationResponse, error) {
	var publication models.Publication
	if err := h.db.NewSelect().Model(&publication).Where("id = ?", publicationID).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, huma.Error404NotFound("publication not found")
		}
		return nil, huma.Error500InternalServerError("failed to load publication")
	}
	mediaIDsByPublication, err := h.loadPublicationMediaIDs(ctx, []string{publication.ID})
	if err != nil {
		return nil, err
	}
	return publicationResponse(publication, mediaIDsByPublication[publication.ID]), nil
}

func replacePublicationAssetsTx(ctx context.Context, tx bun.Tx, publicationID string, mediaIDs []string) error {
	if _, err := tx.NewDelete().
		Model((*models.PublicationAsset)(nil)).
		Where("publication_id = ?", publicationID).
		Exec(ctx); err != nil {
		return err
	}
	if len(mediaIDs) == 0 {
		return nil
	}
	assets := make([]models.PublicationAsset, 0, len(mediaIDs))
	now := time.Now().UTC()
	for i, mediaID := range mediaIDs {
		assets = append(assets, models.PublicationAsset{
			PublicationID: publicationID,
			MediaID:       mediaID,
			DisplayOrder:  i,
			CreatedAt:     now,
		})
	}
	_, err := tx.NewInsert().Model(&assets).Exec(ctx)
	return err
}

func publicationResponse(publication models.Publication, mediaIDs []string) *PublicationResponse {
	return &PublicationResponse{
		ID:              publication.ID,
		WorkspaceID:     publication.WorkspaceID,
		CreatedByID:     publication.CreatedByID,
		Title:           publication.Title,
		SourceContent:   publication.SourceContent,
		SourceURL:       publication.SourceURL,
		Goal:            publication.Goal,
		Audience:        publication.Audience,
		Status:          publication.Status,
		ReleasePlanJSON: publication.ReleasePlanJSON,
		MediaIDs:        mediaIDs,
		CreatedAt:       publication.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       publication.UpdatedAt.Format(time.RFC3339),
	}
}

func applyPublicationUpdateBody(publication *models.Publication, body UpdatePublicationBody) ([]string, error) {
	columns := []string{"updated_at"}
	if body.Title != nil {
		title := strings.TrimSpace(*body.Title)
		if title == "" {
			return nil, huma.Error400BadRequest("title cannot be empty")
		}
		publication.Title = title
		columns = append(columns, "title")
	}
	if body.SourceContent != nil {
		sourceContent := strings.TrimSpace(*body.SourceContent)
		if sourceContent == "" {
			return nil, huma.Error400BadRequest("source_content cannot be empty")
		}
		publication.SourceContent = sourceContent
		columns = append(columns, "source_content")
	}
	if body.SourceURL != nil {
		sourceURL := strings.TrimSpace(*body.SourceURL)
		if err := validatePublicationSourceURL(sourceURL); err != nil {
			return nil, err
		}
		publication.SourceURL = sourceURL
		columns = append(columns, "source_url")
	}
	if body.Goal != nil {
		publication.Goal = strings.TrimSpace(*body.Goal)
		columns = append(columns, "goal")
	}
	if body.Audience != nil {
		publication.Audience = strings.TrimSpace(*body.Audience)
		columns = append(columns, "audience")
	}
	if body.Status != nil {
		status := strings.TrimSpace(*body.Status)
		if err := validatePublicationStatus(status); err != nil {
			return nil, err
		}
		publication.Status = status
		columns = append(columns, "status")
	}
	return columns, nil
}

func (h *PublicationHandler) CreatePublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-publication",
		Method:      http.MethodPost,
		Path:        "/publications",
		Summary:     "Create a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *CreatePublicationInput) (*CreatePublicationOutput, error) {
		userID := middleware.GetUserID(ctx)
		workspaceID := strings.TrimSpace(input.Body.WorkspaceID)
		if err := h.checkWorkspaceAccess(ctx, workspaceID, userID); err != nil {
			return nil, err
		}
		title := strings.TrimSpace(input.Body.Title)
		sourceContent := strings.TrimSpace(input.Body.SourceContent)
		if title == "" {
			return nil, huma.Error400BadRequest("title is required")
		}
		if sourceContent == "" {
			return nil, huma.Error400BadRequest("source_content is required")
		}
		sourceURL := strings.TrimSpace(input.Body.SourceURL)
		if err := validatePublicationSourceURL(sourceURL); err != nil {
			return nil, err
		}
		mediaIDs, err := h.validateMediaBelongsToWorkspace(ctx, workspaceID, input.Body.MediaIDs)
		if err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		publication := &models.Publication{
			ID:              uuid.NewString(),
			WorkspaceID:     workspaceID,
			CreatedByID:     userID,
			Title:           title,
			SourceContent:   sourceContent,
			SourceURL:       sourceURL,
			Goal:            strings.TrimSpace(input.Body.Goal),
			Audience:        strings.TrimSpace(input.Body.Audience),
			Status:          models.PublicationStatusDraft,
			ReleasePlanJSON: "{}",
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			if _, err := tx.NewInsert().Model(publication).Exec(txCtx); err != nil {
				return err
			}
			return replacePublicationAssetsTx(txCtx, tx, publication.ID, mediaIDs)
		}); err != nil {
			return nil, huma.Error500InternalServerError("failed to create publication")
		}

		body, err := h.loadPublicationResponse(ctx, publication.ID)
		if err != nil {
			return nil, err
		}
		return &CreatePublicationOutput{Body: body}, nil
	})
}

func (h *PublicationHandler) ListPublications(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-publications",
		Method:      http.MethodGet,
		Path:        "/publications",
		Summary:     "List publications",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *ListPublicationsInput) (*ListPublicationsOutput, error) {
		userID := middleware.GetUserID(ctx)
		workspaceID := strings.TrimSpace(input.WorkspaceID)
		if err := h.checkWorkspaceAccess(ctx, workspaceID, userID); err != nil {
			return nil, err
		}
		status := strings.TrimSpace(input.Status)
		if err := validatePublicationStatus(status); err != nil {
			return nil, err
		}
		limit := input.Limit
		if limit == 0 {
			limit = 50
		}
		if limit < 1 || limit > 200 {
			return nil, huma.Error400BadRequest("limit must be between 1 and 200")
		}
		if input.Offset < 0 {
			return nil, huma.Error400BadRequest("offset must be greater than or equal to 0")
		}

		var rows []models.Publication
		query := h.db.NewSelect().
			Model(&rows).
			Where("workspace_id = ?", workspaceID).
			Order("created_at DESC").
			Limit(limit).
			Offset(input.Offset)
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if err := query.Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, huma.Error500InternalServerError("failed to list publications")
		}
		ids := make([]string, 0, len(rows))
		for _, publication := range rows {
			ids = append(ids, publication.ID)
		}
		mediaIDsByPublication, err := h.loadPublicationMediaIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		out := make([]PublicationResponse, 0, len(rows))
		for _, publication := range rows {
			out = append(out, *publicationResponse(publication, mediaIDsByPublication[publication.ID]))
		}
		return &ListPublicationsOutput{Body: out}, nil
	})
}

func (h *PublicationHandler) GetPublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-publication",
		Method:      http.MethodGet,
		Path:        publicationPathByID,
		Summary:     "Get a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *GetPublicationInput) (*GetPublicationOutput, error) {
		userID := middleware.GetUserID(ctx)
		body, err := h.loadPublicationResponse(ctx, input.PathID)
		if err != nil {
			return nil, err
		}
		if err := h.checkWorkspaceAccess(ctx, body.WorkspaceID, userID); err != nil {
			return nil, err
		}
		return &GetPublicationOutput{Body: body}, nil
	})
}

func (h *PublicationHandler) UpdatePublication(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "update-publication",
		Method:      http.MethodPatch,
		Path:        publicationPathByID,
		Summary:     "Update a publication",
		Tags:        []string{tagPublications},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403, 404},
	}, func(ctx context.Context, input *UpdatePublicationInput) (*UpdatePublicationOutput, error) {
		userID := middleware.GetUserID(ctx)
		var publication models.Publication
		if err := h.db.NewSelect().Model(&publication).Where("id = ?", input.PathID).Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("publication not found")
			}
			return nil, huma.Error500InternalServerError("failed to load publication")
		}
		if err := h.checkWorkspaceAccess(ctx, publication.WorkspaceID, userID); err != nil {
			return nil, err
		}

		columns, err := applyPublicationUpdateBody(&publication, input.Body)
		if err != nil {
			return nil, err
		}

		var mediaIDs []string
		if input.Body.MediaIDs != nil {
			mediaIDs, err = h.validateMediaBelongsToWorkspace(ctx, publication.WorkspaceID, *input.Body.MediaIDs)
			if err != nil {
				return nil, err
			}
		}
		publication.UpdatedAt = time.Now().UTC()
		if err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			if _, err := tx.NewUpdate().
				Model(&publication).
				Column(columns...).
				WherePK().
				Exec(txCtx); err != nil {
				return err
			}
			if input.Body.MediaIDs != nil {
				return replacePublicationAssetsTx(txCtx, tx, publication.ID, mediaIDs)
			}
			return nil
		}); err != nil {
			return nil, huma.Error500InternalServerError("failed to update publication")
		}

		body, err := h.loadPublicationResponse(ctx, publication.ID)
		if err != nil {
			return nil, err
		}
		return &UpdatePublicationOutput{Body: body}, nil
	})
}
