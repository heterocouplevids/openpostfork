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
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/mediasigner"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/openpost/backend/internal/services/usage"
	"github.com/uptrace/bun"
)

const (
	ThumbnailSizeSM             = 150
	ThumbnailSizeMD             = 400
	MaxMediaUploadBytes   int64 = 50 * 1024 * 1024
	MediaUploadSessionTTL       = 15 * time.Minute
	defaultMediaMimeType        = "application/octet-stream"
	mediaProcessingStatus       = "processing"
	mediaReadyStatus            = "ready"
	mediaFailedStatus           = "failed"
)

type MediaHandler struct {
	db      *bun.DB
	storage mediastore.BlobStorage
	auth    *auth.Service
	authn   middleware.Authenticator
	signer  *mediasigner.Signer
	quota   entitlements.Service
	usage   *usage.Service
}

type mediaUploadBytesInput struct {
	WorkspaceID      string
	Filename         string
	DeclaredMimeType string
	Size             int64
	Content          []byte
	AltText          string
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
	return &MediaHandler{
		db:      db,
		storage: storage,
		auth:    authService,
		authn:   authenticator,
		signer:  signer,
		quota:   entitlements.NewSelfHostedService(),
		usage:   usage.NewService(db),
	}
}

func (h *MediaHandler) SetEntitlement(entitlement entitlements.Service) {
	if entitlement != nil {
		h.quota = entitlement
	}
}

func (h *MediaHandler) SetUsage(usageService *usage.Service) {
	if usageService != nil {
		h.usage = usageService
	}
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
	Size      int64  `json:"size" doc:"File size in bytes"`
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

type CreateMediaUploadSessionInput struct {
	Body struct {
		WorkspaceID string `json:"workspace_id" doc:"Workspace ID"`
		Filename    string `json:"filename" doc:"Original filename"`
		MimeType    string `json:"mime_type,omitempty" doc:"Declared MIME type"`
		Size        int64  `json:"size" doc:"Expected upload size in bytes"`
		AltText     string `json:"alt_text,omitempty" doc:"Alt text for accessibility"`
	}
}

type DirectMediaUploadTarget struct {
	Method    string            `json:"method" doc:"HTTP method to use for the direct upload"`
	URL       string            `json:"url" doc:"Presigned upload URL"`
	Headers   map[string]string `json:"headers" doc:"Headers that must be sent with the upload request"`
	ExpiresAt string            `json:"expires_at" doc:"Upload URL expiration time"`
	ObjectKey string            `json:"object_key" doc:"Storage object key reserved for the upload"`
}

type CreateMediaUploadSessionOutput struct {
	Body struct {
		MediaID     string                  `json:"media_id" doc:"Pending media ID"`
		Upload      DirectMediaUploadTarget `json:"upload" doc:"Direct upload request details"`
		CompleteURL string                  `json:"complete_url" doc:"API path to call after the direct upload succeeds"`
	}
}

type CompleteMediaUploadSessionInput struct {
	PathID string `path:"id" doc:"Pending media ID"`
	Body   struct {
		WorkspaceID string `json:"workspace_id" doc:"Workspace ID"`
	}
}

type MediaUploadResult struct {
	ID       string `json:"id" doc:"Media ID"`
	MimeType string `json:"mime_type" doc:"MIME type"`
	URL      string `json:"url" doc:"URL to access the media"`
	Size     int64  `json:"size" doc:"File size in bytes"`
	Deduped  bool   `json:"deduped" doc:"Whether an existing media attachment was reused"`
}

type CompleteMediaUploadSessionOutput struct {
	Body MediaUploadResult
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

		if err := h.ensureMediaWorkspaceAccess(ctx, userID, input.WorkspaceID); err != nil {
			return nil, err
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
		total, err := query.Count(ctx)
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

		if err := h.ensureMediaWorkspaceAccess(ctx, userID, media.WorkspaceID); err != nil {
			return nil, err
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

		if err := h.ensureMediaWorkspaceAccess(ctx, userID, media.WorkspaceID); err != nil {
			return nil, err
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

			if err := h.ensureMediaWorkspaceAccess(ctx, userID, media.WorkspaceID); err != nil {
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

		if err := h.ensureMediaWorkspaceAccess(ctx, userID, media.WorkspaceID); err != nil {
			return nil, err
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

		if err := h.ensureMediaWorkspaceAccess(ctx, userID, media.WorkspaceID); err != nil {
			return nil, err
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

	huma.Register(api, huma.Operation{
		OperationID: "create-media-upload-session",
		Method:      http.MethodPost,
		Path:        "/media/upload-session",
		Summary:     "Create a direct-to-storage media upload session",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{400, 403},
	}, func(ctx context.Context, input *CreateMediaUploadSessionInput) (*CreateMediaUploadSessionOutput, error) {
		userID := middleware.GetUserID(ctx)

		workspaceID := strings.TrimSpace(input.Body.WorkspaceID)
		if workspaceID == "" {
			return nil, huma.Error400BadRequest(errWorkspaceIDRequired)
		}
		if err := h.ensureMediaWorkspaceAccess(ctx, userID, workspaceID); err != nil {
			return nil, err
		}

		filename := cleanUploadFilename(input.Body.Filename)
		if filename == "" {
			return nil, huma.Error400BadRequest("filename is required")
		}
		if input.Body.Size <= 0 {
			return nil, huma.Error400BadRequest("size must be positive")
		}
		if input.Body.Size > MaxMediaUploadBytes {
			return nil, huma.Error400BadRequest("file size exceeds 50MB limit")
		}
		if err := h.checkUploadQuota(ctx, workspaceID, input.Body.Size); err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}

		directStorage, ok := h.storage.(mediastore.DirectUploadStorage)
		if !ok {
			return nil, huma.Error400BadRequest("direct media upload sessions require s3 storage")
		}

		mediaID := uuid.New().String()
		objectKey := mediaID + filepath.Ext(filename)
		mimeType := strings.TrimSpace(input.Body.MimeType)
		if mimeType == "" {
			mimeType = defaultMediaMimeType
		}
		session, err := directStorage.CreateDirectUploadSession(ctx, mediastore.DirectUploadInput{
			Key:         objectKey,
			ContentType: mimeType,
			Size:        input.Body.Size,
			ExpiresIn:   MediaUploadSessionTTL,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to create media upload session")
		}

		now := time.Now().UTC()
		media := &models.MediaAttachment{
			ID:               mediaID,
			WorkspaceID:      workspaceID,
			FilePath:         session.Key,
			StorageType:      h.storage.Driver(),
			MimeType:         mimeType,
			ProcessingStatus: mediaProcessingStatus,
			Size:             input.Body.Size,
			OriginalFilename: filename,
			FileHash:         "pending:" + mediaID,
			AltText:          input.Body.AltText,
			CreatedAt:        now,
		}
		if _, err := h.db.NewInsert().Model(media).Exec(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to reserve media upload")
		}

		return &CreateMediaUploadSessionOutput{Body: struct {
			MediaID     string                  `json:"media_id" doc:"Pending media ID"`
			Upload      DirectMediaUploadTarget `json:"upload" doc:"Direct upload request details"`
			CompleteURL string                  `json:"complete_url" doc:"API path to call after the direct upload succeeds"`
		}{
			MediaID: mediaID,
			Upload: DirectMediaUploadTarget{
				Method:    session.Method,
				URL:       session.URL,
				Headers:   session.Headers,
				ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
				ObjectKey: session.Key,
			},
			CompleteURL: "/api/v1/media/upload-session/" + mediaID + "/complete",
		}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "complete-media-upload-session",
		Method:      http.MethodPost,
		Path:        "/media/upload-session/{id}/complete",
		Summary:     "Complete a direct-to-storage media upload session",
		Tags:        []string{tagMedia},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.authn)},
		Errors:      []int{400, 403, 404},
	}, func(ctx context.Context, input *CompleteMediaUploadSessionInput) (*CompleteMediaUploadSessionOutput, error) {
		userID := middleware.GetUserID(ctx)
		workspaceID := strings.TrimSpace(input.Body.WorkspaceID)
		if workspaceID == "" {
			return nil, huma.Error400BadRequest(errWorkspaceIDRequired)
		}

		result, err := h.completeDirectMediaUpload(ctx, userID, workspaceID, input.PathID)
		if err != nil {
			return nil, err
		}
		return &CompleteMediaUploadSessionOutput{Body: result}, nil
	})
}

type mediaUsageSummary struct {
	Total    int
	Blocking int
}

func (h *MediaHandler) ensureMediaWorkspaceAccess(ctx context.Context, userID, workspaceID string) error {
	if !middleware.WorkspaceScopeAllows(ctx, workspaceID) {
		return huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	var memberCount int
	memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return huma.Error500InternalServerError(errValidateWorkspaceAccess)
	}
	if memberCount == 0 {
		return huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	return nil
}

func cleanUploadFilename(filename string) string {
	filename = strings.TrimSpace(strings.ReplaceAll(filename, "\\", "/"))
	if filename == "" {
		return ""
	}
	filename = filepath.Base(filename)
	if filename == "." || filename == "/" {
		return ""
	}
	return filename
}

func (h *MediaHandler) completeDirectMediaUpload(ctx context.Context, userID, workspaceID, mediaID string) (MediaUploadResult, error) {
	var result MediaUploadResult
	media, err := h.loadDirectMediaUpload(ctx, userID, workspaceID, mediaID)
	if err != nil {
		return result, err
	}
	if media.ProcessingStatus == mediaReadyStatus {
		return mediaUploadResultFromAttachment(media, false), nil
	}

	content, err := h.readDirectMediaUploadContent(ctx, media)
	if err != nil {
		return result, err
	}
	if err := h.checkUploadQuotaExcludingMedia(ctx, workspaceID, int64(len(content)), media.ID); err != nil {
		return result, huma.Error400BadRequest(err.Error())
	}

	hash := sha256.Sum256(content)
	fileHash := hex.EncodeToString(hash[:])
	if existing, found, err := h.findDuplicateMedia(ctx, workspaceID, fileHash, media.ID); err != nil {
		return result, err
	} else if found {
		_ = h.storage.Delete(filepath.Base(media.FilePath))
		_, _ = h.db.NewDelete().Model(&media).Where("id = ?", media.ID).Exec(ctx)
		return mediaUploadResultFromAttachment(existing, true), nil
	}

	media, err = h.finalizeDirectMediaUploadRecord(ctx, media, content, fileHash)
	if err != nil {
		return result, err
	}
	if _, err := h.usage.IncrementMonthly(ctx, workspaceID, entitlements.LimitMediaBytesUploadedMonthly, media.Size, time.Now().UTC()); err != nil {
		return result, huma.Error500InternalServerError("failed to record media upload usage")
	}

	return mediaUploadResultFromAttachment(media, false), nil
}

func (h *MediaHandler) loadDirectMediaUpload(ctx context.Context, userID, workspaceID, mediaID string) (models.MediaAttachment, error) {
	var media models.MediaAttachment
	if strings.TrimSpace(mediaID) == "" {
		return media, huma.Error400BadRequest("media id is required")
	}
	if err := h.db.NewSelect().Model(&media).Where("id = ?", mediaID).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return media, huma.Error404NotFound(errMediaNotFound)
		}
		return media, huma.Error500InternalServerError("failed to fetch media")
	}
	if media.WorkspaceID != workspaceID {
		return media, huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	if err := h.ensureMediaWorkspaceAccess(ctx, userID, workspaceID); err != nil {
		return media, err
	}
	if media.ProcessingStatus != mediaReadyStatus && media.ProcessingStatus != mediaProcessingStatus {
		return media, huma.Error400BadRequest("media upload session is not pending")
	}
	if media.StorageType != h.storage.Driver() {
		return media, huma.Error400BadRequest("media upload session belongs to a different storage driver")
	}
	return media, nil
}

func (h *MediaHandler) readDirectMediaUploadContent(ctx context.Context, media models.MediaAttachment) ([]byte, error) {
	file, err := h.storage.Open(filepath.Base(media.FilePath))
	if err != nil {
		h.markMediaUploadFailed(ctx, media.ID)
		return nil, huma.Error400BadRequest("uploaded media object was not found")
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, MaxMediaUploadBytes+1))
	if err != nil {
		h.markMediaUploadFailed(ctx, media.ID)
		return nil, huma.Error500InternalServerError("failed to read uploaded media")
	}
	if int64(len(content)) > MaxMediaUploadBytes {
		h.markMediaUploadFailed(ctx, media.ID)
		return nil, huma.Error400BadRequest("file size exceeds 50MB limit")
	}
	if len(content) == 0 {
		h.markMediaUploadFailed(ctx, media.ID)
		return nil, huma.Error400BadRequest("uploaded media object is empty")
	}
	if media.Size > 0 && media.Size != int64(len(content)) {
		h.markMediaUploadFailed(ctx, media.ID)
		return nil, huma.Error400BadRequest("uploaded media size does not match upload session")
	}
	return content, nil
}

func (h *MediaHandler) findDuplicateMedia(ctx context.Context, workspaceID, fileHash, mediaID string) (models.MediaAttachment, bool, error) {
	var existing models.MediaAttachment
	err := h.db.NewSelect().Model(&existing).
		Where("workspace_id = ? AND file_hash = ? AND id != ?", workspaceID, fileHash, mediaID).
		Scan(ctx)
	if err == nil {
		return existing, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return existing, false, nil
	}
	return existing, false, huma.Error500InternalServerError("failed to check duplicate media")
}

func (h *MediaHandler) finalizeDirectMediaUploadRecord(ctx context.Context, media models.MediaAttachment, content []byte, fileHash string) (models.MediaAttachment, error) {
	mimeType := detectedMediaMimeType(content, media.MimeType)
	width, height := 0, 0
	var thumbnails Thumbnails
	var err error
	if strings.HasPrefix(mimeType, "image/") {
		width, height, thumbnails, err = h.processImage(content, media.ID, mimeType)
		if err != nil {
			width, height = h.getImageDimensions(bytes.NewReader(content), mimeType)
		}
	}
	thumbsJSON := ""
	if encoded, err := json.Marshal(thumbnails); err == nil {
		thumbsJSON = string(encoded)
	}

	media.MimeType = mimeType
	media.ProcessingStatus = mediaReadyStatus
	media.Size = int64(len(content))
	media.FileHash = fileHash
	media.Width = width
	media.Height = height
	media.ThumbnailsJSON = thumbsJSON
	if _, err := h.db.NewUpdate().
		Model(&media).
		Column("mime_type", "processing_status", "size", "file_hash", "width", "height", "thumbnails").
		Where("id = ?", media.ID).
		Exec(ctx); err != nil {
		h.markMediaUploadFailed(ctx, media.ID)
		return media, huma.Error500InternalServerError("failed to finalize media record")
	}
	return media, nil
}

func detectedMediaMimeType(content []byte, fallback string) string {
	mimeType := http.DetectContentType(content)
	if !strings.HasPrefix(mimeType, defaultMediaMimeType) {
		return mimeType
	}
	if fallback != "" {
		return fallback
	}
	return defaultMediaMimeType
}

func (h *MediaHandler) markMediaUploadFailed(ctx context.Context, mediaID string) {
	_, _ = h.db.NewUpdate().
		Model((*models.MediaAttachment)(nil)).
		Set("processing_status = ?", mediaFailedStatus).
		Where("id = ?", mediaID).
		Exec(ctx)
}

func mediaUploadResultFromAttachment(media models.MediaAttachment, deduped bool) MediaUploadResult {
	return MediaUploadResult{
		ID:       media.ID,
		MimeType: media.MimeType,
		URL:      "/media/" + media.ID,
		Size:     media.Size,
		Deduped:  deduped,
	}
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

	if ok, err := h.userCanAccessWorkspace(c.Request().Context(), workspaceID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: "failed to validate workspace access"})
	} else if !ok {
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
			Size:     m.Size,
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

	if ok, err := h.userCanAccessWorkspace(c.Request().Context(), workspaceID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
	} else if !ok {
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

	if ok, err := h.userCanAccessWorkspace(c.Request().Context(), workspaceID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: errValidateWorkspaceAccess})
	} else if !ok {
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
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read file")
	}

	return h.processUploadBytes(context.Background(), mediaUploadBytesInput{
		WorkspaceID:      workspaceID,
		Filename:         fileHeader.Filename,
		DeclaredMimeType: fileHeader.Header.Get("Content-Type"),
		Size:             fileHeader.Size,
		Content:          content,
		AltText:          altText,
	})
}

func (h *MediaHandler) processUploadBytes(ctx context.Context, input mediaUploadBytesInput) (map[string]interface{}, error) {
	if input.Size > MaxMediaUploadBytes {
		return nil, errors.New("file size exceeds 50MB limit")
	}
	if input.Size < 0 {
		return nil, errors.New("file size is invalid")
	}
	if int64(len(input.Content)) != input.Size {
		input.Size = int64(len(input.Content))
	}

	var err error
	hash := sha256.Sum256(input.Content)
	fileHash := hex.EncodeToString(hash[:])

	mimeType := http.DetectContentType(input.Content)
	if strings.HasPrefix(mimeType, defaultMediaMimeType) {
		mimeType = input.DeclaredMimeType
		if mimeType == "" {
			mimeType = defaultMediaMimeType
		}
	}

	var existing models.MediaAttachment
	err = h.db.NewSelect().Model(&existing).
		Where("workspace_id = ? AND file_hash = ?", input.WorkspaceID, fileHash).
		Scan(ctx)
	if err == nil {
		return map[string]interface{}{
			"id":        existing.ID,
			"mime_type": existing.MimeType,
			"url":       "/media/" + existing.ID,
			"size":      existing.Size,
			"deduped":   true,
		}, nil
	}
	if err := h.checkUploadQuota(ctx, input.WorkspaceID, input.Size); err != nil {
		return nil, err
	}

	mediaID := uuid.New().String()
	ext := filepath.Ext(input.Filename)
	filename := mediaID + ext

	savedPath, err := h.storage.Save(filename, bytes.NewReader(input.Content))
	if err != nil {
		return nil, errors.New("failed to save media")
	}

	media := &models.MediaAttachment{
		ID:               mediaID,
		WorkspaceID:      input.WorkspaceID,
		FilePath:         savedPath,
		StorageType:      h.storage.Driver(),
		MimeType:         mimeType,
		ProcessingStatus: mediaReadyStatus,
		Size:             input.Size,
		OriginalFilename: input.Filename,
		FileHash:         fileHash,
		AltText:          input.AltText,
	}

	width, height := 0, 0
	var thumbnails Thumbnails

	if strings.HasPrefix(mimeType, "image/") {
		width, height, thumbnails, err = h.processImage(input.Content, mediaID, mimeType)
		if err != nil {
			width, height = h.getImageDimensions(bytes.NewReader(input.Content), mimeType)
		}
		media.Width = width
		media.Height = height
		if thumbsJSON, err := json.Marshal(thumbnails); err == nil {
			media.ThumbnailsJSON = string(thumbsJSON)
		}
	}

	if _, err := h.db.NewInsert().Model(media).Exec(ctx); err != nil {
		return nil, errors.New("failed to save media record")
	}
	if _, err := h.usage.IncrementMonthly(ctx, input.WorkspaceID, entitlements.LimitMediaBytesUploadedMonthly, input.Size, time.Now().UTC()); err != nil {
		return nil, errors.New("failed to record media upload usage")
	}

	return map[string]interface{}{
		"id":        mediaID,
		"mime_type": mimeType,
		"url":       "/media/" + mediaID,
		"size":      input.Size,
		"deduped":   false,
	}, nil
}

func (h *MediaHandler) checkUploadQuota(ctx context.Context, workspaceID string, size int64) error {
	return h.checkUploadQuotaExcludingMedia(ctx, workspaceID, size, "")
}

func (h *MediaHandler) checkUploadQuotaExcludingMedia(ctx context.Context, workspaceID string, size int64, excludeMediaID string) error {
	uploaded, err := h.usage.CurrentMonthly(ctx, workspaceID, entitlements.LimitMediaBytesUploadedMonthly, time.Now().UTC())
	if err != nil {
		return errors.New("failed to load upload usage")
	}
	if err := h.checkQuota(ctx, workspaceID, entitlements.LimitMediaBytesUploadedMonthly, uploaded, size); err != nil {
		return err
	}

	var stored int64
	storedQuery := h.db.NewSelect().
		Model((*models.MediaAttachment)(nil)).
		ColumnExpr("COALESCE(SUM(size), 0)").
		Where("workspace_id = ?", workspaceID)
	if excludeMediaID != "" {
		storedQuery = storedQuery.Where("id != ?", excludeMediaID)
	}
	if err := storedQuery.Scan(ctx, &stored); err != nil {
		return errors.New("failed to load stored media usage")
	}
	return h.checkQuota(ctx, workspaceID, entitlements.LimitMediaBytesStored, stored, size)
}

func (h *MediaHandler) checkQuota(ctx context.Context, workspaceID string, limit entitlements.LimitKey, current, amount int64) error {
	decision, err := h.quota.Check(ctx, entitlements.Request{
		WorkspaceID: workspaceID,
		Limit:       limit,
		Current:     current,
		Amount:      amount,
	})
	if err != nil {
		return errors.New("failed to check quota")
	}
	if !decision.Allowed {
		if decision.Reason != "" {
			return errors.New(decision.Reason)
		}
		return errors.New("quota exceeded")
	}
	return nil
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
			if authHeader != "" && h.authn != nil {
				return middleware.BearerMiddleware(h.authn)(next)(c)
			}
			if authHeader != "" && h.auth != nil {
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
		if userID := h.userIDFromQueryToken(c.Request().Context(), token); userID != "" {
			allowed, err := h.userCanAccessWorkspace(c.Request().Context(), media.WorkspaceID, userID)
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

func (h *MediaHandler) userIDFromQueryToken(ctx context.Context, token string) string {
	if h.authn != nil {
		principal, err := h.authn.AuthenticateBearer(ctx, token)
		if err == nil && principal != nil {
			return principal.UserID
		}
		return ""
	}
	if h.auth != nil {
		claims, err := h.auth.ValidateToken(token)
		if err == nil && claims != nil {
			return claims.UserID
		}
	}
	return ""
}

func (h *MediaHandler) userCanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error) {
	if !middleware.WorkspaceScopeAllows(ctx, workspaceID) {
		return false, nil
	}
	memberCount, err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return false, err
	}

	return memberCount > 0, nil
}
