package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type publicationTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newPublicationTestServer(t *testing.T) *publicationTestServer {
	t.Helper()

	db := createHandlerTestDB(t,
		(*models.WorkspaceMember)(nil),
		(*models.MediaAttachment)(nil),
		(*models.Publication)(nil),
		(*models.PublicationAsset)(nil),
	)
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-2",
		UserID:      "other-user",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)
	media := []models.MediaAttachment{
		{
			ID:               "media-1",
			WorkspaceID:      "ws-1",
			FilePath:         "media-1.png",
			StorageType:      "local",
			MimeType:         "image/png",
			ProcessingStatus: "ready",
			OriginalFilename: "media-1.png",
			FileHash:         "media-1-hash",
		},
		{
			ID:               "media-2",
			WorkspaceID:      "ws-1",
			FilePath:         "media-2.png",
			StorageType:      "local",
			MimeType:         "image/png",
			ProcessingStatus: "ready",
			OriginalFilename: "media-2.png",
			FileHash:         "media-2-hash",
		},
		{
			ID:               "media-other",
			WorkspaceID:      "ws-2",
			FilePath:         "media-other.png",
			StorageType:      "local",
			MimeType:         "image/png",
			ProcessingStatus: "ready",
			OriginalFilename: "media-other.png",
			FileHash:         "media-other-hash",
		},
	}
	_, err = db.NewInsert().Model(&media).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	handler := NewPublicationHandler(db, testAuthenticator{})
	handler.CreatePublication(api)
	handler.ListPublications(api)
	handler.GetPublication(api)
	handler.UpdatePublication(api)

	return &publicationTestServer{echo: e, db: db}
}

func (s *publicationTestServer) request(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Buffer
	if body != nil {
		payload = &bytes.Buffer{}
		require.NoError(t, json.NewEncoder(payload).Encode(body))
	} else {
		payload = &bytes.Buffer{}
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, payload)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func seedPublication(t *testing.T, db *bun.DB, publication models.Publication, mediaIDs ...string) {
	t.Helper()
	if publication.CreatedAt.IsZero() {
		publication.CreatedAt = time.Date(2026, 6, 30, 18, 0, 0, 0, time.UTC)
	}
	if publication.UpdatedAt.IsZero() {
		publication.UpdatedAt = publication.CreatedAt
	}
	if publication.ReleasePlanJSON == "" {
		publication.ReleasePlanJSON = "{}"
	}
	_, err := db.NewInsert().Model(&publication).Exec(context.Background())
	require.NoError(t, err)

	assets := make([]models.PublicationAsset, 0, len(mediaIDs))
	for i, mediaID := range mediaIDs {
		assets = append(assets, models.PublicationAsset{
			PublicationID: publication.ID,
			MediaID:       mediaID,
			DisplayOrder:  i,
			CreatedAt:     publication.CreatedAt,
		})
	}
	if len(assets) > 0 {
		_, err = db.NewInsert().Model(&assets).Exec(context.Background())
		require.NoError(t, err)
	}
}

func TestCreatePublicationReturnsSourceAndMedia(t *testing.T) {
	t.Parallel()

	srv := newPublicationTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/publications", map[string]any{
		"workspace_id":   "ws-1",
		"title":          "MCP launch",
		"source_content": "OpenPost now supports source publications.",
		"source_url":     "https://openpost.social/blog/mcp-launch",
		"goal":           "announce",
		"audience":       "builders",
		"media_ids":      []string{"media-1"},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out PublicationResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.NotEmpty(t, out.ID)
	require.Equal(t, "ws-1", out.WorkspaceID)
	require.Equal(t, "MCP launch", out.Title)
	require.Equal(t, models.PublicationStatusDraft, out.Status)
	require.Equal(t, []string{"media-1"}, out.MediaIDs)

	var assetCount int
	require.NoError(t, srv.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("publication_assets").
		Where("publication_id = ?", out.ID).
		Scan(context.Background(), &assetCount))
	require.Equal(t, 1, assetCount)
}

func TestCreatePublicationRejectsForeignMedia(t *testing.T) {
	t.Parallel()

	srv := newPublicationTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/publications", map[string]any{
		"workspace_id":   "ws-1",
		"title":          "Bad media",
		"source_content": "This should fail.",
		"media_ids":      []string{"media-other"},
	})

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "media attachments are invalid")
}

func TestListPublicationsFiltersWorkspaceAndStatus(t *testing.T) {
	t.Parallel()

	srv := newPublicationTestServer(t)
	seedPublication(t, srv.db, models.Publication{
		ID:            "pub-draft",
		WorkspaceID:   "ws-1",
		CreatedByID:   "user-1",
		Title:         "Draft source",
		SourceContent: "Draft source content.",
		Status:        models.PublicationStatusDraft,
		CreatedAt:     time.Date(2026, 6, 30, 18, 0, 0, 0, time.UTC),
	}, "media-1")
	seedPublication(t, srv.db, models.Publication{
		ID:            "pub-ready",
		WorkspaceID:   "ws-1",
		CreatedByID:   "user-1",
		Title:         "Ready source",
		SourceContent: "Ready source content.",
		Status:        models.PublicationStatusReady,
		CreatedAt:     time.Date(2026, 6, 30, 19, 0, 0, 0, time.UTC),
	})
	seedPublication(t, srv.db, models.Publication{
		ID:            "pub-other",
		WorkspaceID:   "ws-2",
		CreatedByID:   "other-user",
		Title:         "Other source",
		SourceContent: "Other workspace content.",
		Status:        models.PublicationStatusDraft,
		CreatedAt:     time.Date(2026, 6, 30, 20, 0, 0, 0, time.UTC),
	})

	resp := srv.request(t, http.MethodGet, "/api/v1/publications?workspace_id=ws-1&status=draft", nil)

	require.Equal(t, http.StatusOK, resp.Code)
	var out []PublicationResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Len(t, out, 1)
	require.Equal(t, "pub-draft", out[0].ID)
	require.Equal(t, []string{"media-1"}, out[0].MediaIDs)
}

func TestUpdatePublicationReplacesMediaAndStatus(t *testing.T) {
	t.Parallel()

	srv := newPublicationTestServer(t)
	seedPublication(t, srv.db, models.Publication{
		ID:            "pub-update",
		WorkspaceID:   "ws-1",
		CreatedByID:   "user-1",
		Title:         "Old title",
		SourceContent: "Old source.",
		Status:        models.PublicationStatusDraft,
	}, "media-1")

	resp := srv.request(t, http.MethodPatch, "/api/v1/publications/pub-update", map[string]any{
		"title":     "Ready title",
		"status":    "ready",
		"media_ids": []string{"media-2"},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out PublicationResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "Ready title", out.Title)
	require.Equal(t, models.PublicationStatusReady, out.Status)
	require.Equal(t, []string{"media-2"}, out.MediaIDs)

	var mediaIDs []string
	require.NoError(t, srv.db.NewSelect().
		Model((*models.PublicationAsset)(nil)).
		Column("media_id").
		Where("publication_id = ?", "pub-update").
		Order("display_order ASC").
		Scan(context.Background(), &mediaIDs))
	require.Equal(t, []string{"media-2"}, mediaIDs)
}
