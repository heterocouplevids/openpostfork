package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type mcpTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newMCPTestServer(t *testing.T) *mcpTestServer {
	return newMCPTestServerWithEntitlement(t, nil)
}

func newMCPTestServerWithEntitlement(t *testing.T, entitlement entitlements.Service) *mcpTestServer {
	t.Helper()

	db := createHandlerTestDB(
		t,
		(*models.Workspace)(nil),
		(*models.WorkspaceMember)(nil),
		(*models.SocialAccount)(nil),
		(*models.Post)(nil),
		(*models.PostDestination)(nil),
		(*models.Job)(nil),
		(*models.UsageCounter)(nil),
	)
	ctx := context.Background()
	workspaces := []models.Workspace{
		{ID: "ws-1", Name: "Launch", CreatedAt: time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)},
		{ID: "ws-2", Name: "Personal", CreatedAt: time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)},
	}
	_, err := db.NewInsert().Model(&workspaces).Exec(ctx)
	require.NoError(t, err)
	members := []models.WorkspaceMember{
		{WorkspaceID: "ws-1", UserID: "user-1", Role: models.WorkspaceRoleAdmin},
		{WorkspaceID: "ws-2", UserID: "user-1", Role: models.WorkspaceRoleEditor},
	}
	_, err = db.NewInsert().Model(&members).Exec(ctx)
	require.NoError(t, err)
	accounts := []models.SocialAccount{
		{
			ID:              "account-1",
			WorkspaceID:     "ws-1",
			Platform:        "x",
			AccountID:       "x-1",
			AccountUsername: "openpost",
			Slug:            "x-openpost",
			AccessTokenEnc:  []byte("token"),
			IsActive:        true,
			CreatedAt:       time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:             "account-inactive",
			WorkspaceID:    "ws-1",
			Platform:       "mastodon",
			AccountID:      "masto-1",
			Slug:           "mastodon-old",
			AccessTokenEnc: []byte("token"),
			IsActive:       true,
			CreatedAt:      time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC),
		},
		{
			ID:             "account-other-workspace",
			WorkspaceID:    "ws-2",
			Platform:       "bluesky",
			AccountID:      "did:plc:abc",
			Slug:           "bsky-personal",
			AccessTokenEnc: []byte("token"),
			IsActive:       true,
			CreatedAt:      time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC),
		},
	}
	_, err = db.NewInsert().Model(&accounts).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewUpdate().
		Model((*models.SocialAccount)(nil)).
		Set("is_active = ?", false).
		Where("id = ?", "account-inactive").
		Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	NewMCPHandler(db, testAuthenticator{}, entitlement).RegisterRoutes(e)
	return &mcpTestServer{echo: e, db: db}
}

func (s *mcpTestServer) request(t *testing.T, token string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload bytes.Buffer
	require.NoError(t, json.NewEncoder(&payload).Encode(body))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", &payload)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func TestMCPRejectsMissingAuthorization(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestMCPToolsList(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	tools := result["tools"].([]any)
	require.Len(t, tools, 6)
	require.Equal(t, "list_workspaces", tools[0].(map[string]any)["name"])
	require.Equal(t, "list_accounts", tools[1].(map[string]any)["name"])
	require.Equal(t, "create_draft", tools[2].(map[string]any)["name"])
	require.Equal(t, "schedule_post", tools[3].(map[string]any)["name"])
	require.Equal(t, "get_post_status", tools[4].(map[string]any)["name"])
	require.Equal(t, "cancel_post", tools[5].(map[string]any)["name"])
}

func TestMCPCallListWorkspaces(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-1",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_workspaces",
			"arguments": map[string]any{},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	content := result["content"].([]any)
	require.Contains(t, content[0].(map[string]any)["text"], "Launch")
	structured := result["structuredContent"].(map[string]any)
	workspaces := structured["workspaces"].([]any)
	require.Len(t, workspaces, 2)
	require.Equal(t, "ws-1", workspaces[0].(map[string]any)["id"])
	require.Equal(t, "admin", workspaces[0].(map[string]any)["role"])
}

func TestMCPCallListAccounts(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-accounts",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_accounts",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	content := result["content"].([]any)
	require.Contains(t, content[0].(map[string]any)["text"], "x:x-openpost")
	structured := result["structuredContent"].(map[string]any)
	accounts := structured["accounts"].([]any)
	require.Len(t, accounts, 1)
	require.Equal(t, "account-1", accounts[0].(map[string]any)["id"])
	require.Equal(t, "x", accounts[0].(map[string]any)["platform"])
}

func TestMCPCallCreateDraft(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-draft",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_draft",
			"arguments": map[string]any{
				"workspace_id":       "ws-1",
				"content":            "Draft from an agent",
				"social_account_ids": []string{"account-1"},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	post := structured["post"].(map[string]any)
	require.Equal(t, "draft", post["status"])
	require.Equal(t, "ws-1", post["workspace_id"])
	postID := post["id"].(string)
	require.NotEmpty(t, postID)

	var stored models.Post
	require.NoError(t, srv.db.NewSelect().Model(&stored).Where("id = ?", postID).Scan(context.Background()))
	require.Equal(t, "Draft from an agent", stored.Content)
	require.Equal(t, "user-1", stored.CreatedByID)
	var destinationCount int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("post_destinations").Where("post_id = ?", postID).Scan(context.Background(), &destinationCount))
	require.Equal(t, 1, destinationCount)
}

func TestMCPCallCreateDraftRejectsOutsideAccount(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-draft",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_draft",
			"arguments": map[string]any{
				"workspace_id":       "ws-1",
				"content":            "Draft from an agent",
				"social_account_ids": []string{"account-other-workspace"},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	rpcErr := out["error"].(map[string]any)
	require.Contains(t, rpcErr["message"], "outside this workspace")
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("posts").Scan(context.Background(), &count))
	require.Equal(t, 0, count)
}

func TestMCPCallSchedulePostCreatesPublishJob(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	scheduledAt := "2026-07-01T12:00:00Z"
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-schedule",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "schedule_post",
			"arguments": map[string]any{
				"workspace_id":       "ws-1",
				"content":            "Ship agentic scheduling",
				"scheduled_at":       scheduledAt,
				"social_account_ids": []string{"account-1"},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	post := structured["post"].(map[string]any)
	require.Equal(t, "scheduled", post["status"])
	require.Equal(t, scheduledAt, post["scheduled_at"])
	require.Equal(t, "Ship agentic scheduling", post["content"])
	destinations := post["destinations"].([]any)
	require.Len(t, destinations, 1)
	require.Equal(t, "account-1", destinations[0].(map[string]any)["social_account_id"])
	postID := post["id"].(string)

	var storedPost models.Post
	require.NoError(t, srv.db.NewSelect().Model(&storedPost).Where("id = ?", postID).Scan(context.Background()))
	require.Equal(t, statusScheduled, storedPost.Status)
	require.Equal(t, "user-1", storedPost.CreatedByID)
	require.Equal(t, time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC), storedPost.ScheduledAt)
	require.Equal(t, storedPost.ScheduledAt, storedPost.ActualRunAt)

	var job models.Job
	require.NoError(t, srv.db.NewSelect().Model(&job).Where("type = ?", jobTypePublishPost).Scan(context.Background()))
	require.Equal(t, "pending", job.Status)
	require.Equal(t, storedPost.ScheduledAt, job.RunAt)
	var payload map[string]string
	require.NoError(t, json.Unmarshal([]byte(job.Payload), &payload))
	require.Equal(t, postID, payload[postIDKey])
}

func TestMCPCallGetPostStatusReturnsDestinations(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	scheduledAt := time.Date(2026, 7, 2, 9, 30, 0, 0, time.UTC)
	post := models.Post{
		ID:          "post-status",
		WorkspaceID: "ws-1",
		CreatedByID: "user-1",
		Content:     "Check the launch queue",
		Status:      statusScheduled,
		ScheduledAt: scheduledAt,
		ActualRunAt: scheduledAt,
		CreatedAt:   time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC),
	}
	_, err := srv.db.NewInsert().Model(&post).Exec(context.Background())
	require.NoError(t, err)
	destination := models.PostDestination{
		ID:              "destination-status",
		PostID:          post.ID,
		SocialAccountID: "account-1",
		Status:          postStatusPending,
	}
	_, err = srv.db.NewInsert().Model(&destination).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-status",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_post_status",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"post_id":      post.ID,
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	gotPost := structured["post"].(map[string]any)
	require.Equal(t, post.ID, gotPost["id"])
	require.Equal(t, "scheduled", gotPost["status"])
	require.Equal(t, scheduledAt.Format(time.RFC3339), gotPost["actual_run_at"])
	destinations := gotPost["destinations"].([]any)
	require.Len(t, destinations, 1)
	require.Equal(t, "x", destinations[0].(map[string]any)["platform"])
	require.Equal(t, "x-openpost", destinations[0].(map[string]any)["slug"])
}

func TestMCPCallCancelPostRemovesQueuedJobAndReturnsDraft(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	postID := "post-cancel"
	scheduledAt := time.Date(2026, 7, 3, 8, 0, 0, 0, time.UTC)
	_, err := srv.db.NewInsert().Model(&models.Post{
		ID:          postID,
		WorkspaceID: "ws-1",
		CreatedByID: "user-1",
		Content:     "Cancel me",
		Status:      statusScheduled,
		ScheduledAt: scheduledAt,
		ActualRunAt: scheduledAt,
		CreatedAt:   time.Date(2026, 6, 30, 16, 0, 0, 0, time.UTC),
	}).Exec(context.Background())
	require.NoError(t, err)
	payload, err := json.Marshal(map[string]string{postIDKey: postID})
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.Job{
		ID:      "job-cancel",
		Type:    jobTypePublishPost,
		Payload: string(payload),
		Status:  "pending",
		RunAt:   scheduledAt,
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-cancel",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "cancel_post",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"post_id":      postID,
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	post := structured["post"].(map[string]any)
	require.Equal(t, "draft", post["status"])
	require.NotContains(t, post, "scheduled_at")
	require.NotContains(t, post, "actual_run_at")

	var storedPost models.Post
	require.NoError(t, srv.db.NewSelect().Model(&storedPost).Where("id = ?", postID).Scan(context.Background()))
	require.Equal(t, statusDraft, storedPost.Status)
	require.True(t, storedPost.ScheduledAt.IsZero())
	require.True(t, storedPost.ActualRunAt.IsZero())
	var jobCount int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("jobs").Scan(context.Background(), &jobCount))
	require.Equal(t, 0, jobCount)
}

func TestMCPCallSchedulePostHonorsQuota(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServerWithEntitlement(t, entitlements.NewStaticService(entitlements.PlanSnapshot{
		Limits: map[entitlements.LimitKey]int64{
			entitlements.LimitScheduledPostsMonthly: 0,
		},
	}))
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-schedule-quota",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "schedule_post",
			"arguments": map[string]any{
				"workspace_id":       "ws-1",
				"content":            "This should hit the limit",
				"scheduled_at":       "2026-07-01T12:00:00Z",
				"social_account_ids": []string{"account-1"},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	rpcErr := out["error"].(map[string]any)
	require.Contains(t, rpcErr["message"], "scheduled_posts_monthly")
	var postCount int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("posts").Scan(context.Background(), &postCount))
	require.Equal(t, 0, postCount)
}
