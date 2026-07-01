package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func TestApplyRandomDelayStaysWithinBounds(t *testing.T) {
	scheduledAt := time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)
	const maxDelay = 15

	for i := 0; i < 200; i++ {
		actual := applyRandomDelay(scheduledAt, maxDelay)
		diff := actual.Sub(scheduledAt)
		if diff < -15*time.Minute || diff > 15*time.Minute {
			t.Fatalf("random delay out of bounds: got %v", diff)
		}
	}
}

func TestApplyRandomDelayWithZeroDelayReturnsScheduledTime(t *testing.T) {
	scheduledAt := time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)

	actual := applyRandomDelay(scheduledAt, 0)
	if !actual.Equal(scheduledAt) {
		t.Fatalf("expected unchanged time, got %s want %s", actual, scheduledAt)
	}
}

func TestListPostsOrderExpressionKeepsCoalesceCall(t *testing.T) {
	sqldb, err := sql.Open("sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=private")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqldb.Close()
	})

	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() {
		_ = db.Close()
	})

	query := db.NewSelect().
		Model((*models.Post)(nil))
	query = applyListPostsOrder(query).Limit(50)

	require.Contains(t, query.String(), "ORDER BY COALESCE(scheduled_at, created_at) DESC")

	_, err = db.NewCreateTable().Model((*models.Post)(nil)).IfNotExists().Exec(context.Background())
	require.NoError(t, err)

	var posts []models.Post
	query = db.NewSelect().Model(&posts)
	err = applyListPostsOrder(query).Limit(50).Scan(context.Background())
	require.NoError(t, err)
}

func TestListPostsPaginatesVisiblePostsWithHeaders(t *testing.T) {
	t.Parallel()

	srv := newListPostsTestServer(t)
	srv.seedPosts(t)

	resp := srv.getJSON(t, "/api/v1/posts?limit=2&offset=1")

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	require.Equal(t, "4", resp.Header().Get("X-Total-Count"))
	require.Equal(t, "2", resp.Header().Get("X-Limit"))
	require.Equal(t, "1", resp.Header().Get("X-Offset"))
	require.Equal(t, "3", resp.Header().Get("X-Next-Offset"))
	require.Equal(t, "true", resp.Header().Get("X-Has-More"))

	var out []PostResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Len(t, out, 2)
	require.Equal(t, "post-3", out[0].ID)
	require.Equal(t, "post-2", out[1].ID)
}

func TestListPostsCountsFilteredWorkspaceScope(t *testing.T) {
	t.Parallel()

	srv := newListPostsTestServer(t)
	srv.seedPosts(t)

	resp := srv.getJSON(t, "/api/v1/posts?workspace_id=ws-1&status=draft&limit=1")

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	require.Equal(t, "2", resp.Header().Get("X-Total-Count"))
	require.Equal(t, "1", resp.Header().Get("X-Limit"))
	require.Equal(t, "0", resp.Header().Get("X-Offset"))
	require.Equal(t, "1", resp.Header().Get("X-Next-Offset"))
	require.Equal(t, "true", resp.Header().Get("X-Has-More"))

	var out []PostResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Len(t, out, 1)
	require.Equal(t, statusDraft, out[0].Status)
}

func TestListPostsRejectsNegativeOffset(t *testing.T) {
	t.Parallel()

	srv := newListPostsTestServer(t)

	resp := srv.getJSON(t, "/api/v1/posts?offset=-1")

	require.Equal(t, http.StatusBadRequest, resp.Code, resp.Body.String())
	require.Contains(t, resp.Body.String(), "offset must be greater than or equal to 0")
}

type listPostsTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newListPostsTestServer(t *testing.T) *listPostsTestServer {
	t.Helper()

	db := createHandlerTestDB(
		t,
		(*models.WorkspaceMember)(nil),
		(*models.Post)(nil),
		(*models.PostDestination)(nil),
		(*models.SocialAccount)(nil),
		(*models.PostMedia)(nil),
	)
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	NewPostHandler(db, testAuthenticator{}).ListPosts(api)
	return &listPostsTestServer{echo: e, db: db}
}

func (s *listPostsTestServer) seedPosts(t *testing.T) {
	t.Helper()

	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	posts := []models.Post{
		{ID: "post-1", WorkspaceID: "ws-1", CreatedByID: "user-1", Content: "one", Status: statusDraft, ScheduledAt: now.Add(time.Minute), CreatedAt: now.Add(time.Minute)},
		{ID: "post-2", WorkspaceID: "ws-1", CreatedByID: "user-1", Content: "two", Status: statusScheduled, ScheduledAt: now.Add(2 * time.Minute), CreatedAt: now.Add(2 * time.Minute)},
		{ID: "post-3", WorkspaceID: "ws-1", CreatedByID: "user-1", Content: "three", Status: statusDraft, ScheduledAt: now.Add(3 * time.Minute), CreatedAt: now.Add(3 * time.Minute)},
		{ID: "post-4", WorkspaceID: "ws-1", CreatedByID: "user-1", Content: "four", Status: statusScheduled, ScheduledAt: now.Add(4 * time.Minute), CreatedAt: now.Add(4 * time.Minute)},
		{ID: "post-foreign", WorkspaceID: "ws-2", CreatedByID: "other-user", Content: "foreign", Status: statusDraft, ScheduledAt: now.Add(5 * time.Minute), CreatedAt: now.Add(5 * time.Minute)},
	}
	_, err := s.db.NewInsert().Model(&posts).Exec(context.Background())
	require.NoError(t, err)
}

func (s *listPostsTestServer) getJSON(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}
