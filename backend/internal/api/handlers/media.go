package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/auth"
	"github.com/openpost/backend/internal/services/mediasigner"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/uptrace/bun"
)

const (
	ThumbnailSizeSM = 150
	ThumbnailSizeMD = 400
)

type MediaHandler struct {
	db      *bun.DB
	storage mediastore.BlobStorage
	auth    *auth.Service
	authn   middleware.Authenticator
	signer  *mediasigner.Signer
}

func NewMediaHandler(
	db *bun.DB,
	storage mediastore.BlobStorage,
	authService *auth.Service,
	authenticator middleware.Authenticator,
	signer *mediasigner.Signer,
) *MediaHandler {
	if authenticator == nil && authService != nil {
		authenticator = middleware.NewJWTAuthenticator(authService)
	}
	return &MediaHandler{db: db, storage: storage, auth: authService, authn: authenticator, signer: signer}
}

type Thumbnails struct {
	SM string `json:"sm,omitempty"`
	MD string `json:"md,omitempty"`
}

type MediaUsageItem struct {
	PostID    string `json:"post_id" doc:"Post ID"`
	Content   string `json:"content" doc:"Post content (truncated)"`
	Status    string `json:"status" doc:"Post status"`
	Scheduled string `json:"scheduled_at" doc:"Scheduled time"`
}

type MediaListItem struct {
	ID               string `json:"id" doc:"Media ID"`
	WorkspaceID      string `json:"workspace_id" doc:"Workspace ID"`
	MimeType         string `json:"mime_type" doc:"MIME type"`
	Size             int64  `json:"size" doc:"File size in bytes"`
	OriginalFilename string `json:"original_filename" doc:"Original filename"`
	Width            int    `json:"width" doc:"Image width"`
	Height           int    `json:"height" doc:"Image height"`
	AltText          string `json:"alt_text" doc:"Alt text"`
	IsFavorite       bool   `json:"is_favorite" doc:"Whether media is favorited"`
	CreatedAt        string `json:"created_at" doc:"Creation time"`
	URL              string `json:"url" doc:"URL to access the media"`
	ThumbnailURL     string `json:"thumbnail_url" doc:"Thumbnail URL for grid view"`
	UsageCount       int    `json:"usage_count" doc:"Number of posts using this media"`
	CanDelete        bool   `json:"can_delete" doc:"Whether media can be deleted"`
	ProcessingStatus string `json:"processing_status" doc:"Processing status"`
}

type ListMediaInput struct {
	WorkspaceID string `query:"workspace_id" required:"true" doc:"Filter by workspace ID"`
	Filter      string `query:"filter" doc:"Filter: all, used, unused, favorites"`
	Sort        string `query:"sort" doc:"Sort: newest, oldest, size"`
	Limit       int    `query:"limit" doc:"Limit (default 50, max 200)"`
	Offset      int    `query:"offset" doc:"Offset for pagination"`
}

type ListMediaOutput struct {
	Body struct {
		Media []MediaListItem `json:"media" doc:"Media attachments"`
		Total int             `json:"total" doc:"Total count matching filter"`
	}
}

type GetMediaUsageInput struct {
	PathID string `path:"id" doc:"Media ID"`
}

type GetMediaUsageOutput struct {
	Body struct {
		Usage []MediaUsageItem `json:"usage" doc:"Posts using this media"`
		Count int              `json:"count" doc:"Number of posts using this media"`
	}
}

type MediaMetadataItem struct {
	ID        string `json:"id" doc:"Media ID"`
	MimeType  string `json:"mime_type" doc:"MIME type"`
	AltText   string `json:"alt_text" doc:"Alt text"`
	Width     int    `json:"width" doc:"Image width"`
	Height    int    `json:"height" doc:"Image height"`
	URL       string `json:"url" doc:"URL to access the media"`
	Thumbnail string `json:"thumbnail_url" doc:"Thumbnail URL"`
}

type MediaMetadataInput struct {
	WorkspaceID string   `query:"workspace_id" required:"true" doc:"Workspace ID"`
	MediaIDs    []string `query:"media_ids" doc:"Comma-separated list of media IDs"`
}

type MediaMetadataOutput struct {
	Body struct {
		Media []MediaMetadataItem `json:"media" doc:"Media metadata list"`
	}
}

type DeleteMediaInput struct {
	PathID string `path:"id" doc:"Media ID"`
}

type DeleteMediaOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

type BatchDeleteMediaInput struct {
	Body struct {
		MediaIDs []string `json:"media_ids" doc:"Array of media IDs to delete"`
	}
}

type BatchDeleteMediaOutput struct {
	Body struct {
		Deleted   int      `json:"deleted" doc:"Number of media deleted"`
		FailedIDs []string `json:"failed_ids" doc:"IDs that could not be deleted (in use)"`
	}
}

type UpdateMediaFavoriteInput struct {
	PathID string `path:"id" doc:"Media ID"`
}

type UpdateMediaFavoriteOutput struct {
	Body struct {
		IsFavorite bool `json:"is_favorite" doc:"Updated favorite status"`
	}
}

type UpdateMediaInput struct {
	PathID string `path:"id" doc:"Media ID"`
	Body   struct {
		AltText string `json:"alt_text" doc:"Alt text for accessibility"`
	}
}

type UpdateMediaOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

//nolint:gocyclo
func (h *MediaHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-media",
		Method:      http.MethodGet,
		Path:        "/media",
		Summary:     "List media attachments for a workspace",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *ListMediaInput) (*ListMediaOutput, error) {
		userID := middleware.GetUserID(ctx)

		if input.WorkspaceID == "" {
			return nil, huma.Error400BadRequest(errWorkspaceIDRequired)
		}

		var memberCount int
		memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", input.WorkspaceID, userID).
			Count(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError(errValidateWorkspaceAccess)
		}
		if memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		limit := input.Limit
		if limit <= 0 || limit > 200 {
			limit = 50
		}

		query := h.db.NewSelect().Model(&models.MediaAttachment{}).
			Where("workspace_id = ?", input.WorkspaceID)

		switch input.Filter {
		case "favorites":
			query = query.Where("is_favorite = ?", true)
		case "used":
			query = query.Where("id IN (SELECT media_id FROM post_media)")
		case "unused":
			query = query.Where("id NOT IN (SELECT media_id FROM post_media)")
		}

		var total int
		total, err = query.Count(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to count media")
		}

		switch input.Sort {
		case "oldest":
			query = query.Order("created_at ASC")
		case mediaSortSize:
			query = query.Order("size DESC")
		default:
			query = query.Order("created_at DESC")
		}

		var media []models.MediaAttachment
		err = query.Limit(limit).Offset(input.Offset).Scan(ctx, &media)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch media")
		}

		result := make([]MediaListItem, len(media))
		for i, m := range media {
			usage, usageErr := h.mediaUsageSummary(ctx, m.WorkspaceID, m.ID)
			if usageErr != nil {
				return nil, huma.Error500InternalServerError("failed to check media usage")
			}

			var thumbs Thumbnails
			if m.ThumbnailsJSON != "" {
				if err := json.Unmarshal([]byte(m.ThumbnailsJSON), &thumbs); err != nil {
					thumbs = Thumbnails{}
				}
			}

			result[i] = MediaListItem{
				ID:               m.ID,
				WorkspaceID:      m.WorkspaceID,
				MimeType:         m.MimeType,
				Size:             m.Size,
				OriginalFilename: m.OriginalFilename,
				Width:            m.Width,
				Height:           m.Height,
				AltText:          m.AltText,
				IsFavorite:       m.IsFavorite,
				CreatedAt:        m.CreatedAt.Format(time.RFC3339),
				URL:              "/media/" + m.ID,
				ThumbnailURL:     "/media/" + m.ID + "/thumb",
				UsageCount:       usage.Total,
				CanDelete:        usage.Blocking == 0,
				ProcessingStatus: m.ProcessingStatus,
			}
			if thumbs.SM != "" {
				result[i].ThumbnailURL = "/media/" + m.ID + "/thumb/sm"
			}
		}

		return &ListMediaOutput{Body: struct {
			Media []MediaListItem `json:"media" doc:"Media attachments"`
			Total int             `json:"total" doc:"Total count matching filter"`
		}{Media: result, Total: total}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-media-usage",
		Method:      http.MethodGet,
		Path:        "/media/{id}/usage",
		Summary:     "Get posts that use a media attachment",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *GetMediaUsageInput) (*GetMediaUsageOutput, error) {
		userID := middleware.GetUserID(ctx)

		var media models.MediaAttachment
		err := h.db.NewSelect().Model(&media).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound(errMediaNotFound)
			}
			return nil, huma.Error500InternalServerError("failed to fetch media")
		}

		var memberCount int
		memberCount, err = h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", media.WorkspaceID, userID).
			Count(ctx)
		if err != nil || memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		posts, err := h.postsUsingMedia(ctx, media.WorkspaceID, input.PathID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch usage")
		}

		usage := make([]MediaUsageItem, 0, len(posts))
		for _, post := range posts {
			content := post.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			scheduled := ""
			if !post.ScheduledAt.IsZero() {
				scheduled = post.ScheduledAt.Format(time.RFC3339)
			}
			usage = append(usage, MediaUsageItem{
				PostID:    post.ID,
				Content:   content,
				Status:    post.Status,
				Scheduled: scheduled,
			})
		}

		return &GetMediaUsageOutput{Body: struct {
			Usage []MediaUsageItem `json:"usage" doc:"Posts using this media"`
			Count int              `json:"count" doc:"Number of posts using this media"`
		}{Usage: usage, Count: len(usage)}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-media",
		Method:      http.MethodDelete,
		Path:        "/media/{id}",
		Summary:     "Delete a media attachment (only if not used in any post)",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *DeleteMediaInput) (*DeleteMediaOutput, error) {
		userID := middleware.GetUserID(ctx)

		var media models.MediaAttachment
		err := h.db.NewSelect().Model(&media).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound(errMediaNotFound)
			}
			return nil, huma.Error500InternalServerError("failed to fetch media")
		}

		var memberCount int
		memberCount, err = h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", media.WorkspaceID, userID).
			Count(ctx)
		if err != nil || memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		usage, err := h.mediaUsageSummary(ctx, media.WorkspaceID, input.PathID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to check usage")
		}
		if usage.Blocking > 0 {
			return nil, huma.Error400BadRequest("cannot delete media attached to posts that have not been published yet")
		}

		if err := h.deleteMediaFiles(&media); err != nil {
			return nil, huma.Error500InternalServerError("failed to delete media files")
		}

		if err := h.removeMediaReferences(ctx, media.WorkspaceID, input.PathID); err != nil {
			return nil, huma.Error500InternalServerError("failed to remove media references")
		}

		_, err = h.db.NewDelete().Model(&media).Where("id = ?", input.PathID).Exec(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to delete media record")
		}

		return &DeleteMediaOutput{Body: struct {
			Message string `json:"message" doc:"Success message"`
		}{Message: "media deleted successfully"}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "batch-delete-media",
		Method:      http.MethodPost,
		Path:        "/media/batch-delete",
		Summary:     "Delete multiple media attachments at once",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *BatchDeleteMediaInput) (*BatchDeleteMediaOutput, error) {
		userID := middleware.GetUserID(ctx)

		if len(input.Body.MediaIDs) == 0 {
			return nil, huma.Error400BadRequest("media_ids is required")
		}

		if len(input.Body.MediaIDs) > 100 {
			return nil, huma.Error400BadRequest("max 100 media IDs at once")
		}

		deleted := 0
		failedIDs := []string{}

		for _, mediaID := range input.Body.MediaIDs {
			var media models.MediaAttachment
			err := h.db.NewSelect().Model(&media).Where("id = ?", mediaID).Scan(ctx)
			if err != nil {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			var memberCount int
			memberCount, err = h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
				Where("workspace_id = ? AND user_id = ?", media.WorkspaceID, userID).
				Count(ctx)
			if err != nil || memberCount == 0 {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			usage, err := h.mediaUsageSummary(ctx, media.WorkspaceID, mediaID)
			if err != nil || usage.Blocking > 0 {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			err = h.deleteMediaFiles(&media)
			if err != nil {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			if err := h.removeMediaReferences(ctx, media.WorkspaceID, mediaID); err != nil {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			_, err = h.db.NewDelete().Model(&media).Where("id = ?", mediaID).Exec(ctx)
			if err != nil {
				failedIDs = append(failedIDs, mediaID)
				continue
			}

			deleted++
		}

		return &BatchDeleteMediaOutput{Body: struct {
			Deleted   int      `json:"deleted" doc:"Number of media deleted"`
			FailedIDs []string `json:"failed_ids" doc:"IDs that could not be deleted (in use)"`
		}{Deleted: deleted, FailedIDs: failedIDs}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-media-favorite",
		Method:      http.MethodPatch,
		Path:        "/media/{id}/favorite",
		Summary:     "Toggle favorite status of a media attachment",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *UpdateMediaFavoriteInput) (*UpdateMediaFavoriteOutput, error) {
		userID := middleware.GetUserID(ctx)

		var media models.MediaAttachment
		err := h.db.NewSelect().Model(&media).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound(errMediaNotFound)
			}
			return nil, huma.Error500InternalServerError("failed to fetch media")
		}

		var memberCount int
		memberCount, err = h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", media.WorkspaceID, userID).
			Count(ctx)
		if err != nil || memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		media.IsFavorite = !media.IsFavorite
		_, err = h.db.NewUpdate().Model(&media).Column("is_favorite").Where("id = ?", input.PathID).Exec(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to update favorite status")
		}

		return &UpdateMediaFavoriteOutput{Body: struct {
			IsFavorite bool `json:"is_favorite" doc:"Updated favorite status"`
		}{IsFavorite: media.IsFavorite}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-media",
		Method:      http.MethodPatch,
		Path:        "/media/{id}",
		Summary:     "Update media metadata (alt text)",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *UpdateMediaInput) (*UpdateMediaOutput, error) {
		userID := middleware.GetUserID(ctx)

		var media models.MediaAttachment
		err := h.db.NewSelect().Model(&media).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound(errMediaNotFound)
			}
			return nil, huma.Error500InternalServerError("failed to fetch media")
		}

		var memberCount int
		memberCount, err = h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", media.WorkspaceID, userID).
			Count(ctx)
		if err != nil || memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		media.AltText = input.Body.AltText
		_, err = h.db.NewUpdate().Model(&media).Column("alt_text").Where("id = ?", input.PathID).Exec(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to update media")
		}

		return &UpdateMediaOutput{Body: struct {
			Message string `json:"message" doc:"Success message"`
		}{Message: "media updated successfully"}}, nil
	})
}

type mediaUsageSummary struct {
	Total    int
	Blocking int
}

func (h *MediaHandler) mediaUsageSummary(ctx context.Context, workspaceID, mediaID string) (mediaUsageSummary, error) {
	var summary mediaUsageSummary

	posts, err := h.postsUsingMedia(ctx, workspaceID, mediaID)
	if err != nil {
		return summary, err
	}

	summary.Total = len(posts)
	for _, post := range posts {
		if post.Status != models.PostStatusPublished {
			summary.Blocking++
		}
	}

	return summary, nil
}

func (h *MediaHandler) postsUsingMedia(ctx context.Context, workspaceID, mediaID string) ([]models.Post, error) {
	postRows := []models.Post{}
	if err := h.db.NewSelect().
		TableExpr("post_media AS pm").
		ColumnExpr("p.*").
		Join("JOIN posts AS p ON p.id = pm.post_id").
		Where("p.workspace_id = ?", workspaceID).
		Where("pm.media_id = ?", mediaID).
		Scan(ctx, &postRows); err != nil {
		return nil, err
	}

	var variants []models.PostVariant
	if err := h.db.NewSelect().
		Model(&variants).
		Where("media_ids != ''").
		Scan(ctx); err != nil {
		return nil, err
	}

	postsByID := make(map[string]models.Post, len(postRows)+len(variants))
	for _, post := range postRows {
		postsByID[post.ID] = post
	}
	for _, variant := range variants {
		if !variantContainsMedia(variant.MediaIDs, mediaID) {
			continue
		}
		var post models.Post
		if err := h.db.NewSelect().Model(&post).Where("id = ?", variant.PostID).Scan(ctx); err != nil {
			continue
		}
		if post.WorkspaceID != workspaceID {
			continue
		}
		postsByID[post.ID] = post
	}

	posts := make([]models.Post, 0, len(postsByID))
	for _, post := range postsByID {
		posts = append(posts, post)
	}
	return posts, nil
}

func (h *MediaHandler) removeMediaReferences(ctx context.Context, workspaceID, mediaID string) error {
	if _, err := h.db.NewDelete().
		Model((*models.PostMedia)(nil)).
		Where("media_id = ?", mediaID).
		Exec(ctx); err != nil {
		return err
	}

	var variants []models.PostVariant
	if err := h.db.NewSelect().
		Model(&variants).
		Where("media_ids != ''").
		Scan(ctx); err != nil {
		return err
	}

	for _, variant := range variants {
		var post models.Post
		if err := h.db.NewSelect().Model(&post).Where("id = ?", variant.PostID).Scan(ctx); err != nil {
			continue
		}
		if post.WorkspaceID != workspaceID {
			continue
		}

		var ids []string
		if err := json.Unmarshal([]byte(variant.MediaIDs), &ids); err != nil {
			continue
		}

		changed := false
		filtered := ids[:0]
		for _, id := range ids {
			if id == mediaID {
				changed = true
				continue
			}
			filtered = append(filtered, id)
		}
		if !changed {
			continue
		}

		encoded, err := json.Marshal(filtered)
		if err != nil {
			return err
		}
		if _, err := h.db.NewUpdate().
			Model(&variant).
			Column("media_ids").
			Set("media_ids = ?", string(encoded)).
			Where("id = ?", variant.ID).
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func variantContainsMedia(mediaIDsJSON, mediaID string) bool {
	var ids []string
	if err := json.Unmarshal([]byte(mediaIDsJSON), &ids); err != nil {
		return false
	}
	for _, id := range ids {
		if id == mediaID {
			return true
		}
	}
	return false
}

func (h *MediaHandler) mediaMetadata(c echo.Context) error {
	userID := c.Get(string(middleware.UserIDKey)).(string)

	workspaceID := c.QueryParam("workspace_id")
	if workspaceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: errWorkspaceIDRequired})
	}

	mediaIDsRaw := c.QueryParam("media_ids")
	if mediaIDsRaw == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{"media": []MediaMetadataItem{}})
	}

	mediaIDs := strings.Split(mediaIDsRaw, ",")
	for i := range mediaIDs {
		mediaIDs[i] = strings.TrimSpace(mediaIDs[i])
		if mediaIDs[i] == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "media_ids must not contain empty values"})
		}
	}

	var memberCount int
	memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: "failed to validate workspace access"})
	}
	if memberCount == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{fieldError: errWorkspaceAccessDenied})
	}

	var media []models.MediaAttachment
	if err := h.db.NewSelect().Model(&media).
		Where("workspace_id = ? AND id IN (?)", workspaceID, bun.List(mediaIDs)).
		Scan(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: "failed to fetch media"})
	}

	result := make([]MediaMetadataItem, 0, len(media))
	for _, m := range media {
		item := MediaMetadataItem{
			ID:       m.ID,
			MimeType: m.MimeType,
			AltText:  m.AltText,
			Width:    m.Width,
			Height:   m.Height,
			URL:      "/media/" + m.ID,
		}
		if thumbsJSON := m.ThumbnailsJSON; thumbsJSON != "" {
			var thumbs Thumbnails
			if json.Unmarshal([]byte(thumbsJSON), &thumbs) == nil && thumbs.SM != "" {
				item.Thumbnail = "/media/" + m.ID + "/thumb/sm"
			}
		}
		result = append(result, item)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"media": result})
}

func (h *MediaHandler) deleteMediaFiles(media *models.MediaAttachment) error {
	if err := h.storage.Delete(filepath.Base(media.FilePath)); err != nil {
		return err
	}

	var thumbs Thumbnails
	if media.ThumbnailsJSON != "" {
		_ = json.Unmarshal([]byte(media.ThumbnailsJSON), &thumbs)
	}

	if thumbs.SM != "" {
		h.storage.Delete(thumbs.SM) //nolint:errcheck
	}
	if thumbs.MD != "" {
		h.storage.Delete(thumbs.MD) //nolint:errcheck
	}

	return nil
}

func (h *MediaHandler) RegisterLegacyRoutes(e *echo.Echo) {
	// Legacy upload routes support both web (JWT) and CLI (op_cli_...)
	// credentials via the unified Authenticator. AuthMiddleware cannot
	// be used here because these are raw Echo handlers, not Huma ops.
	uploadAuth := middleware.BearerMiddleware(h.authn)
	e.POST("/api/v1/media/upload", h.uploadMedia, uploadAuth)
	e.POST("/api/v1/media/batch-upload", h.batchUploadMedia, uploadAuth)
	e.GET("/api/v1/media/metadata", h.mediaMetadata, uploadAuth)
	e.GET("/media/:id", h.serveMedia, h.optionalMediaAuth())
	e.HEAD("/media/:id", h.serveMedia, h.optionalMediaAuth())
	e.GET("/media/:id/thumb/:size", h.serveThumbnailSize, h.optionalMediaAuth())
	e.HEAD("/media/:id/thumb/:size", h.serveThumbnailSize, h.optionalMediaAuth())
}

func (h *MediaHandler) uploadMedia(c echo.Context) error {
	userID := c.Get(string(middleware.UserIDKey)).(string)

	workspaceID := c.FormValue("workspace_id")
	if workspaceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: errWorkspaceIDRequired})
	}

	var memberCount int
	memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
	}
	if memberCount == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{fieldError: errWorkspaceAccessDenied})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "file is required"})
	}

	result, err := h.processUpload(workspaceID, fileHeader, c.FormValue("alt_text"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *MediaHandler) batchUploadMedia(c echo.Context) error {
	userID := c.Get(string(middleware.UserIDKey)).(string)

	workspaceID := c.FormValue("workspace_id")
	if workspaceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: errWorkspaceIDRequired})
	}

	var memberCount int
	memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
	}
	if memberCount == 0 {
		return c.JSON(http.StatusForbidden, map[string]string{fieldError: errWorkspaceAccessDenied})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "failed to parse multipart form"})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "no files provided"})
	}

	if len(files) > 10 {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "max 10 files at once"})
	}

	results := []map[string]interface{}{}
	uploadErrors := []string{}

	for _, fileHeader := range files {
		result, err := h.processUpload(workspaceID, fileHeader, "")
		if err != nil {
			uploadErrors = append(uploadErrors, fileHeader.Filename+": "+err.Error())
			continue
		}
		results = append(results, result)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"uploaded": results,
		"errors":   uploadErrors,
	})
}

func (h *MediaHandler) processUpload(workspaceID string, fileHeader *multipart.FileHeader, altText string) (map[string]interface{}, error) {
	if fileHeader.Size > 50*1024*1024 {
		return nil, errors.New("file size exceeds 50MB limit")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read file")
	}

	hash := sha256.Sum256(content)
	fileHash := hex.EncodeToString(hash[:])

	mimeType := http.DetectContentType(content)
	if strings.HasPrefix(mimeType, "application/octet-stream") {
		mimeType = fileHeader.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	var existing models.MediaAttachment
	err = h.db.NewSelect().Model(&existing).
		Where("workspace_id = ? AND file_hash = ?", workspaceID, fileHash).
		Scan(context.Background())
	if err == nil {
		return map[string]interface{}{
			"id":        existing.ID,
			"mime_type": existing.MimeType,
			"url":       "/media/" + existing.ID,
			"size":      existing.Size,
			"deduped":   true,
		}, nil
	}

	mediaID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	filename := mediaID + ext

	savedPath, err := h.storage.Save(filename, bytes.NewReader(content))
	if err != nil {
		return nil, errors.New("failed to save media")
	}

	media := &models.MediaAttachment{
		ID:               mediaID,
		WorkspaceID:      workspaceID,
		FilePath:         savedPath,
		StorageType:      h.storage.Driver(),
		MimeType:         mimeType,
		ProcessingStatus: "ready",
		Size:             fileHeader.Size,
		OriginalFilename: fileHeader.Filename,
		FileHash:         fileHash,
		AltText:          altText,
	}

	width, height := 0, 0
	var thumbnails Thumbnails

	if strings.HasPrefix(mimeType, "image/") {
		width, height, thumbnails, err = h.processImage(content, mediaID, mimeType)
		if err != nil {
			width, height = h.getImageDimensions(bytes.NewReader(content), mimeType)
		}
		media.Width = width
		media.Height = height
		if thumbsJSON, err := json.Marshal(thumbnails); err == nil {
			media.ThumbnailsJSON = string(thumbsJSON)
		}
	}

	if _, err := h.db.NewInsert().Model(media).Exec(context.Background()); err != nil {
		return nil, errors.New("failed to save media record")
	}

	return map[string]interface{}{
		"id":        mediaID,
		"mime_type": mimeType,
		"url":       "/media/" + mediaID,
		"size":      fileHeader.Size,
		"deduped":   false,
	}, nil
}

func (h *MediaHandler) processImage(content []byte, mediaID, mimeType string) (int, int, Thumbnails, error) {
	reader := bytes.NewReader(content)

	var img image.Image
	var err error

	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/tiff":
		img, err = imaging.Decode(reader)
	default:
		return 0, 0, Thumbnails{}, errors.New("unsupported image format")
	}

	if err != nil {
		return 0, 0, Thumbnails{}, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	thumbnails := Thumbnails{}

	smThumb := imaging.Thumbnail(img, ThumbnailSizeSM, ThumbnailSizeSM, imaging.Lanczos)
	smFilename := "sm_" + mediaID + ".jpg"
	if err := h.saveThumbnail(smFilename, smThumb, imaging.JPEG); err == nil {
		thumbnails.SM = smFilename
	}

	mdThumb := imaging.Thumbnail(img, ThumbnailSizeMD, ThumbnailSizeMD, imaging.Lanczos)
	mdFilename := "md_" + mediaID + ".jpg"
	if err := h.saveThumbnail(mdFilename, mdThumb, imaging.JPEG); err == nil {
		thumbnails.MD = mdFilename
	}

	return width, height, thumbnails, nil
}

func (h *MediaHandler) saveThumbnail(filename string, img image.Image, format imaging.Format) error {
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, format); err != nil {
		return err
	}
	_, err := h.storage.Save(filename, &buf)
	return err
}

func (h *MediaHandler) getImageDimensions(reader io.Reader, _ string) (int, int) {
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func (h *MediaHandler) serveMedia(c echo.Context) error {
	mediaID := c.Param("id")

	// Strip file extension if present (e.g., "abc123.jpg" -> "abc123")
	// Media IDs in the database are UUIDs without extensions, but Threads
	// requires URLs with extensions for content-type detection.
	if dotIdx := strings.LastIndex(mediaID, "."); dotIdx > 0 {
		mediaID = mediaID[:dotIdx]
	}

	media := new(models.MediaAttachment)
	if err := h.db.NewSelect().Model(media).Where("id = ?", mediaID).Scan(c.Request().Context()); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: errMediaNotFound})
	}
	if err := h.authorizeMediaAccess(c, media); err != nil {
		return err
	}

	file, err := h.storage.Open(filepath.Base(media.FilePath))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: "media file not found"})
	}
	defer file.Close()

	c.Response().Header().Set("Content-Type", media.MimeType)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")

	if f, ok := file.(*os.File); ok {
		if stat, err := f.Stat(); err == nil {
			http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), f)
			return nil
		}
	}

	return c.Stream(http.StatusOK, media.MimeType, file)
}

func (h *MediaHandler) serveThumbnailSize(c echo.Context) error {
	mediaID := c.Param("id")

	// Strip file extension if present (e.g., "abc123.jpg" -> "abc123")
	if dotIdx := strings.LastIndex(mediaID, "."); dotIdx > 0 {
		mediaID = mediaID[:dotIdx]
	}

	size := c.Param("size")
	if size == "" {
		size = "md"
	}

	media := new(models.MediaAttachment)
	if err := h.db.NewSelect().Model(media).Where("id = ?", mediaID).Scan(c.Request().Context()); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: errMediaNotFound})
	}
	if err := h.authorizeMediaAccess(c, media); err != nil {
		return err
	}

	var thumbs Thumbnails
	if media.ThumbnailsJSON != "" {
		_ = json.Unmarshal([]byte(media.ThumbnailsJSON), &thumbs)
	}

	var thumbFilename string
	switch size {
	case "sm":
		thumbFilename = thumbs.SM
	case "md":
		thumbFilename = thumbs.MD
	default:
		thumbFilename = thumbs.MD
	}

	if thumbFilename == "" {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: "thumbnail not found"})
	}

	file, err := h.storage.Open(thumbFilename)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: "thumbnail file not found"})
	}
	defer file.Close()

	if f, ok := file.(*os.File); ok {
		if stat, err := f.Stat(); err == nil {
			c.Response().Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		}
	}

	c.Response().Header().Set("Content-Type", "image/jpeg")
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")

	return c.Stream(http.StatusOK, "image/jpeg", file)
}

func (h *MediaHandler) optionalMediaAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				return middleware.JWTMiddleware(h.auth)(next)(c)
			}
			return next(c)
		}
	}
}

func (h *MediaHandler) authorizeMediaAccess(c echo.Context, media *models.MediaAttachment) error {
	if media == nil {
		return c.JSON(http.StatusNotFound, map[string]string{fieldError: errMediaNotFound})
	}

	if userID, _ := c.Get(string(middleware.UserIDKey)).(string); userID != "" {
		allowed, err := h.userCanAccessWorkspace(c.Request().Context(), media.WorkspaceID, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]string{fieldError: errWorkspaceAccessDenied})
		}
		return nil
	}

	if token := c.QueryParam("token"); token != "" {
		claims, err := h.auth.ValidateToken(token)
		if err == nil && claims != nil && claims.UserID != "" {
			allowed, err := h.userCanAccessWorkspace(c.Request().Context(), media.WorkspaceID, claims.UserID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
			}
			if allowed {
				return nil
			}
			return c.JSON(http.StatusForbidden, map[string]string{fieldError: errWorkspaceAccessDenied})
		}
	}

	expiresAtUnix, _ := strconv.ParseInt(c.QueryParam("exp"), 10, 64)
	signature := c.QueryParam("sig")
	if signature == "" || h.signer == nil || !h.signer.Verify(media.ID, signature, expiresAtUnix) {
		return c.JSON(http.StatusUnauthorized, map[string]string{fieldError: "authentication required"})
	}

	return nil
}

func (h *MediaHandler) userCanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error) {
	memberCount, err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return false, err
	}

	return memberCount > 0, nil
}
