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

type postPublicationTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newPostPublicationTestServer(t *testing.T) *postPublicationTestServer {
	t.Helper()

	db := createHandlerTestDB(
		t,
		(*models.Workspace)(nil),
		(*models.WorkspaceMember)(nil),
		(*models.SocialAccount)(nil),
		(*models.Publication)(nil),
		(*models.Post)(nil),
		(*models.PostDestination)(nil),
		(*models.PostMedia)(nil),
		(*models.MediaAttachment)(nil),
		(*models.Job)(nil),
		(*models.ThreadDraft)(nil),
		(*models.UsageCounter)(nil),
	)
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.Workspace{ID: "ws-1", Name: "Launch"}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Workspace{ID: "ws-2", Name: "Other"}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.SocialAccount{
		ID:             "account-1",
		WorkspaceID:    "ws-1",
		Slug:           "main",
		Platform:       "x",
		AccountID:      "x-1",
		AccessTokenEnc: []byte("token"),
		IsActive:       true,
	}).Exec(ctx)
	require.NoError(t, err)
	publications := []models.Publication{
		{
			ID:              "pub-1",
			WorkspaceID:     "ws-1",
			CreatedByID:     "user-1",
			Title:           "Launch source",
			SourceContent:   "Canonical launch notes",
			Status:          models.PublicationStatusDraft,
			ReleasePlanJSON: "{}",
			CreatedAt:       time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:              "pub-2",
			WorkspaceID:     "ws-1",
			CreatedByID:     "user-1",
			Title:           "Follow-up source",
			SourceContent:   "Follow-up notes",
			Status:          models.PublicationStatusDraft,
			ReleasePlanJSON: "{}",
			CreatedAt:       time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC),
		},
		{
			ID:              "pub-other",
			WorkspaceID:     "ws-2",
			CreatedByID:     "other-user",
			Title:           "Other source",
			SourceContent:   "Other workspace notes",
			Status:          models.PublicationStatusDraft,
			ReleasePlanJSON: "{}",
			CreatedAt:       time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC),
		},
	}
	_, err = db.NewInsert().Model(&publications).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	handler := NewPostHandler(db, testAuthenticator{})
	handler.CreatePost(api)
	handler.CreateThread(api)
	handler.ListPosts(api)
	handler.GetPost(api)
	handler.UpdatePost(api)

	return &postPublicationTestServer{echo: e, db: db}
}

func (s *postPublicationTestServer) request(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload := &bytes.Buffer{}
	if body != nil {
		require.NoError(t, json.NewEncoder(payload).Encode(body))
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

func TestCreatePostLinksPublication(t *testing.T) {
	t.Parallel()

	srv := newPostPublicationTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/posts", map[string]any{
		"workspace_id":         "ws-1",
		"content":              "Draft from source",
		"publication_id":       "pub-1",
		"social_account_ids":   []string{"account-1"},
		"random_delay_minutes": 0,
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out PostResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "pub-1", out.PublicationID)

	var stored models.Post
	require.NoError(t, srv.db.NewSelect().Model(&stored).Where("id = ?", out.ID).Scan(context.Background()))
	require.Equal(t, "pub-1", stored.PublicationID)

	listResp := srv.request(t, http.MethodGet, "/api/v1/posts?workspace_id=ws-1", nil)
	require.Equal(t, http.StatusOK, listResp.Code)
	var listed []PostResponse
	require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &listed))
	require.Len(t, listed, 1)
	require.Equal(t, "pub-1", listed[0].PublicationID)

	getResp := srv.request(t, http.MethodGet, "/api/v1/posts/"+out.ID, nil)
	require.Equal(t, http.StatusOK, getResp.Code)
	var detail PostDetailResponse
	require.NoError(t, json.Unmarshal(getResp.Body.Bytes(), &detail))
	require.Equal(t, "pub-1", detail.PublicationID)
}

func TestCreatePostRejectsForeignPublication(t *testing.T) {
	t.Parallel()

	srv := newPostPublicationTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/posts", map[string]any{
		"workspace_id":       "ws-1",
		"content":            "Bad source",
		"publication_id":     "pub-other",
		"social_account_ids": []string{"account-1"},
	})

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "publication_id")
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("posts").Scan(context.Background(), &count))
	require.Equal(t, 0, count)
}

func TestCreateThreadLinksPublicationToEveryPost(t *testing.T) {
	t.Parallel()

	srv := newPostPublicationTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/posts/thread", map[string]any{
		"workspace_id":         "ws-1",
		"publication_id":       "pub-1",
		"social_account_ids":   []string{"account-1"},
		"random_delay_minutes": 0,
		"posts": []map[string]any{
			{"content": "Thread opener"},
			{"content": "Thread reply"},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out CreateThreadOutput
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out.Body))
	require.Len(t, out.Body.PostIDs, 2)

	var posts []models.Post
	require.NoError(t, srv.db.NewSelect().
		Model(&posts).
		Where("id IN (?)", bun.List(out.Body.PostIDs)).
		Order("thread_sequence ASC").
		Scan(context.Background()))
	require.Len(t, posts, 2)
	for _, post := range posts {
		require.Equal(t, "pub-1", post.PublicationID)
	}
}

func TestUpdatePostRelinksAndClearsPublication(t *testing.T) {
	t.Parallel()

	srv := newPostPublicationTestServer(t)
	_, err := srv.db.NewInsert().Model(&models.Post{
		ID:            "post-1",
		WorkspaceID:   "ws-1",
		CreatedByID:   "user-1",
		PublicationID: "pub-1",
		Content:       "Draft from source",
		Status:        statusDraft,
		CreatedAt:     time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC),
	}).Exec(context.Background())
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.Post{
		ID:             "post-2",
		WorkspaceID:    "ws-1",
		CreatedByID:    "user-1",
		PublicationID:  "pub-1",
		ParentPostID:   "post-1",
		Content:        "Thread child",
		Status:         statusDraft,
		ThreadSequence: 1,
		CreatedAt:      time.Date(2026, 6, 30, 15, 1, 0, 0, time.UTC),
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, http.MethodPatch, "/api/v1/posts/post-1", map[string]any{
		"publication_id": "pub-2",
	})
	require.Equal(t, http.StatusOK, resp.Code)
	var out PostDetailResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "pub-2", out.PublicationID)

	var stored models.Post
	require.NoError(t, srv.db.NewSelect().Model(&stored).Where("id = ?", "post-1").Scan(context.Background()))
	require.Equal(t, "pub-2", stored.PublicationID)
	var storedChild models.Post
	require.NoError(t, srv.db.NewSelect().Model(&storedChild).Where("id = ?", "post-2").Scan(context.Background()))
	require.Equal(t, "pub-2", storedChild.PublicationID)

	resp = srv.request(t, http.MethodPatch, "/api/v1/posts/post-1", map[string]any{
		"publication_id": "",
	})
	require.Equal(t, http.StatusOK, resp.Code)
	var cleared PostDetailResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &cleared))
	require.Empty(t, cleared.PublicationID)

	var storedCleared models.Post
	require.NoError(t, srv.db.NewSelect().Model(&storedCleared).Where("id = ?", "post-1").Scan(context.Background()))
	require.Empty(t, storedCleared.PublicationID)
	var storedChildCleared models.Post
	require.NoError(t, srv.db.NewSelect().Model(&storedChildCleared).Where("id = ?", "post-2").Scan(context.Background()))
	require.Empty(t, storedChildCleared.PublicationID)
}
