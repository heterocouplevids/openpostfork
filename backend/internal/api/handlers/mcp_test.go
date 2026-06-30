package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type mcpTestServer struct {
	echo    *echo.Echo
	db      *bun.DB
	handler *MCPHandler
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
		(*models.PostVariant)(nil),
		(*models.Job)(nil),
		(*models.UsageCounter)(nil),
		(*models.PostingSchedule)(nil),
		(*models.MediaAttachment)(nil),
		(*models.MCPToolCall)(nil),
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
	handler := NewMCPHandler(db, testAuthenticator{}, entitlement)
	handler.SetMediaStorage(mediastore.NewLocalStorage(t.TempDir(), "/media"))
	handler.SetPublicURL("https://app.openpost.test")
	handler.RegisterRoutes(e)
	return &mcpTestServer{echo: e, db: db, handler: handler}
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type mcpScopeAuthenticator map[string]middleware.Principal

func (a mcpScopeAuthenticator) AuthenticateBearer(_ context.Context, token string) (*middleware.Principal, error) {
	principal, ok := a[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return &principal, nil
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
	require.Contains(t, resp.Header().Get("WWW-Authenticate"), `resource_metadata="https://app.openpost.test/.well-known/oauth-protected-resource"`)
	require.Contains(t, resp.Header().Get("WWW-Authenticate"), `scope="mcp:full"`)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	meta := out["_meta"].(map[string]any)
	require.Equal(t, resp.Header().Get("WWW-Authenticate"), meta["mcp/www_authenticate"])
}

func TestMCPProtectedResourceMetadata(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()
	srv.echo.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Equal(t, "https://app.openpost.test/mcp", out["resource"])
	require.Equal(t, []any{"https://app.openpost.test"}, out["authorization_servers"])
	require.Equal(t, []any{"mcp:full"}, out["scopes_supported"])
	require.Equal(t, []any{"header"}, out["bearer_methods_supported"])
}

func TestMCPAuthenticatesSupportedTokenScopes(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	srv.handler.auth = mcpScopeAuthenticator{
		"web-token":   {UserID: "user-1", Email: "user@example.com"},
		"mcp-token":   {UserID: "user-1", Email: "user@example.com", Scope: "mcp:full"},
		"cli-token":   {UserID: "user-1", Email: "user@example.com", Scope: "cli:full"},
		"media-token": {UserID: "user-1", Email: "user@example.com", Scope: "media:read"},
	}

	for _, token := range []string{"web-token", "mcp-token", "cli-token"} {
		resp := srv.request(t, token, map[string]any{
			"jsonrpc": "2.0",
			"id":      token,
			"method":  "tools/list",
		})
		require.Equal(t, http.StatusOK, resp.Code, token)
	}

	resp := srv.request(t, "media-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "bad-scope",
		"method":  "tools/list",
	})
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Contains(t, resp.Header().Get("WWW-Authenticate"), `scope="mcp:full"`)
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
	require.Len(t, tools, 10)
	require.Equal(t, "list_workspaces", tools[0].(map[string]any)["name"])
	require.Equal(t, "list_accounts", tools[1].(map[string]any)["name"])
	require.Equal(t, "create_draft", tools[2].(map[string]any)["name"])
	require.Equal(t, "set_post_renditions", tools[3].(map[string]any)["name"])
	require.Equal(t, "schedule_post", tools[4].(map[string]any)["name"])
	require.Equal(t, "get_post_status", tools[5].(map[string]any)["name"])
	require.Equal(t, "list_scheduled_posts", tools[6].(map[string]any)["name"])
	require.Equal(t, "cancel_post", tools[7].(map[string]any)["name"])
	require.Equal(t, "suggest_next_slot", tools[8].(map[string]any)["name"])
	require.Equal(t, "upload_media_from_url", tools[9].(map[string]any)["name"])

	for _, tool := range tools {
		descriptor := tool.(map[string]any)
		securitySchemes := descriptor["securitySchemes"].([]any)
		require.Len(t, securitySchemes, 1)
		scheme := securitySchemes[0].(map[string]any)
		require.Equal(t, "oauth2", scheme["type"])
		require.Equal(t, []any{"mcp:full"}, scheme["scopes"])
		meta := descriptor["_meta"].(map[string]any)
		require.Equal(t, descriptor["securitySchemes"], meta["securitySchemes"])
	}
	annotations := tools[0].(map[string]any)["annotations"].(map[string]any)
	require.Equal(t, true, annotations["readOnlyHint"])
	require.Equal(t, false, annotations["destructiveHint"])
	require.Equal(t, false, annotations["openWorldHint"])
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

func TestMCPCallLogsSuccessfulToolCall(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-log-success",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_accounts",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var call models.MCPToolCall
	require.NoError(t, srv.db.NewSelect().Model(&call).Where("tool_name = ?", "list_accounts").Scan(context.Background()))
	require.Equal(t, "user-1", call.UserID)
	require.Equal(t, "ws-1", call.WorkspaceID)
	require.Equal(t, "success", call.Status)
	require.Empty(t, call.ErrorMessage)
	require.False(t, call.CreatedAt.IsZero())
}

func TestMCPCallLogsFailedToolCall(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-log-error",
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
	var call models.MCPToolCall
	require.NoError(t, srv.db.NewSelect().Model(&call).Where("tool_name = ?", "create_draft").Scan(context.Background()))
	require.Equal(t, "user-1", call.UserID)
	require.Equal(t, "ws-1", call.WorkspaceID)
	require.Equal(t, "error", call.Status)
	require.Contains(t, call.ErrorMessage, "outside this workspace")
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

func TestMCPCallSetPostRenditions(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	post := models.Post{
		ID:          "post-renditions",
		WorkspaceID: "ws-1",
		CreatedByID: "user-1",
		Content:     "One launch thought",
		Status:      statusDraft,
		CreatedAt:   time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC),
	}
	_, err := srv.db.NewInsert().Model(&post).Exec(context.Background())
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.PostDestination{
		ID:              "destination-rendition",
		PostID:          post.ID,
		SocialAccountID: "account-1",
		Status:          postStatusPending,
	}).Exec(context.Background())
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.MediaAttachment{
		ID:               "media-rendition",
		WorkspaceID:      "ws-1",
		FilePath:         "media-rendition.png",
		MimeType:         "image/png",
		ProcessingStatus: "ready",
		Size:             1234,
		OriginalFilename: "launch.png",
		FileHash:         "media-rendition-hash",
		CreatedAt:        time.Date(2026, 6, 30, 15, 5, 0, 0, time.UTC),
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-renditions",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "set_post_renditions",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"post_id":      post.ID,
				"renditions": []map[string]any{{
					"social_account_id": "account-1",
					"content":           "X-native launch copy with a sharper hook",
					"media_ids":         []string{"media-rendition"},
				}},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	content := result["content"].([]any)
	require.Contains(t, content[0].(map[string]any)["text"], "Updated 1 post renditions")
	structured := result["structuredContent"].(map[string]any)
	require.Equal(t, post.ID, structured["post_id"])
	renditions := structured["renditions"].([]any)
	require.Len(t, renditions, 1)
	rendition := renditions[0].(map[string]any)
	require.Equal(t, "account-1", rendition["social_account_id"])
	require.Equal(t, "x", rendition["platform"])
	require.Equal(t, "x-openpost", rendition["slug"])
	require.Equal(t, "X-native launch copy with a sharper hook", rendition["content"])
	require.Equal(t, []any{"media-rendition"}, rendition["media_ids"])
	require.Equal(t, true, rendition["is_unsynced"])

	var stored models.PostVariant
	require.NoError(t, srv.db.NewSelect().Model(&stored).Where("post_id = ?", post.ID).Scan(context.Background()))
	require.Equal(t, "account-1", stored.SocialAccountID)
	require.Equal(t, "X-native launch copy with a sharper hook", stored.Content)
	require.Equal(t, `["media-rendition"]`, stored.MediaIDs)
	require.True(t, stored.IsUnsynced)
}

func TestMCPCallSetPostRenditionsRejectsNonDestinationAccount(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	post := models.Post{
		ID:          "post-renditions-no-destination",
		WorkspaceID: "ws-1",
		CreatedByID: "user-1",
		Content:     "One launch thought",
		Status:      statusDraft,
		CreatedAt:   time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC),
	}
	_, err := srv.db.NewInsert().Model(&post).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-renditions-invalid",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "set_post_renditions",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"post_id":      post.ID,
				"renditions": []map[string]any{{
					"social_account_id": "account-1",
					"content":           "This should not be saved",
				}},
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	rpcErr := out["error"].(map[string]any)
	require.Contains(t, rpcErr["message"], "not destinations")
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("post_variants").Scan(context.Background(), &count))
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

func TestMCPCallListScheduledPostsReturnsQueue(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	posts := []models.Post{
		{
			ID:          "post-list-early",
			WorkspaceID: "ws-1",
			CreatedByID: "user-1",
			Content:     "First queued post",
			Status:      statusScheduled,
			ScheduledAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
			ActualRunAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 6, 30, 17, 0, 0, 0, time.UTC),
		},
		{
			ID:          "post-list-late",
			WorkspaceID: "ws-1",
			CreatedByID: "user-1",
			Content:     "Second queued post",
			Status:      statusScheduled,
			ScheduledAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC),
			ActualRunAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 6, 30, 17, 5, 0, 0, time.UTC),
		},
		{
			ID:          "post-list-draft",
			WorkspaceID: "ws-1",
			CreatedByID: "user-1",
			Content:     "Draft should not be listed",
			Status:      statusDraft,
			CreatedAt:   time.Date(2026, 6, 30, 17, 10, 0, 0, time.UTC),
		},
		{
			ID:          "post-list-other-workspace",
			WorkspaceID: "ws-2",
			CreatedByID: "user-1",
			Content:     "Other workspace queued post",
			Status:      statusScheduled,
			ScheduledAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 6, 30, 17, 15, 0, 0, time.UTC),
		},
	}
	_, err := srv.db.NewInsert().Model(&posts).Exec(context.Background())
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.PostDestination{
		ID:              "destination-list",
		PostID:          "post-list-early",
		SocialAccountID: "account-1",
		Status:          postStatusPending,
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-list-scheduled",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_scheduled_posts",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"from":         "2026-07-01T00:00:00Z",
				"to":           "2026-07-03T00:00:00Z",
				"limit":        10,
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	content := result["content"].([]any)
	require.Contains(t, content[0].(map[string]any)["text"], "Found 2 scheduled posts")
	structured := result["structuredContent"].(map[string]any)
	gotPosts := structured["posts"].([]any)
	require.Len(t, gotPosts, 2)
	first := gotPosts[0].(map[string]any)
	require.Equal(t, "post-list-early", first["id"])
	require.Equal(t, "First queued post", first["content"])
	require.Equal(t, "2026-07-01T09:00:00Z", first["scheduled_at"])
	destinations := first["destinations"].([]any)
	require.Len(t, destinations, 1)
	require.Equal(t, "x", destinations[0].(map[string]any)["platform"])
	require.Equal(t, "post-list-late", gotPosts[1].(map[string]any)["id"])
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

func TestMCPCallSuggestNextSlotReturnsFirstFreeSchedule(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	_, err := srv.db.NewInsert().Model(&[]models.PostingSchedule{
		{
			ID:          "slot-9",
			WorkspaceID: "ws-1",
			UTCHour:     9,
			UTCMinute:   0,
			DayOfWeek:   int(time.Monday),
			Label:       "Morning",
			IsActive:    true,
			CreatedAt:   time.Date(2026, 6, 30, 17, 0, 0, 0, time.UTC),
		},
		{
			ID:          "slot-17",
			WorkspaceID: "ws-1",
			UTCHour:     17,
			UTCMinute:   0,
			DayOfWeek:   int(time.Monday),
			Label:       "Evening",
			IsActive:    true,
			CreatedAt:   time.Date(2026, 6, 30, 17, 5, 0, 0, time.UTC),
		},
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-slot",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "suggest_next_slot",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"after":        "2026-07-06T08:00:00Z",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	suggestion := structured["suggestion"].(map[string]any)
	require.Equal(t, "Next available slot found.", suggestion["message"])
	require.Equal(t, "2026-07-06T09:00:00Z", suggestion["slot_time"])
	require.Equal(t, "2026-07-06T09:00:00Z", suggestion["slot_time_utc"])
	slot := suggestion["slot"].(map[string]any)
	require.Equal(t, "slot-9", slot["id"])
	require.Equal(t, "Morning", slot["label"])
}

func TestMCPCallSuggestNextSlotSkipsOccupiedSlot(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	_, err := srv.db.NewInsert().Model(&[]models.PostingSchedule{
		{
			ID:          "slot-9",
			WorkspaceID: "ws-1",
			UTCHour:     9,
			UTCMinute:   0,
			DayOfWeek:   int(time.Monday),
			Label:       "Morning",
			IsActive:    true,
			CreatedAt:   time.Date(2026, 6, 30, 17, 0, 0, 0, time.UTC),
		},
		{
			ID:          "slot-17",
			WorkspaceID: "ws-1",
			UTCHour:     17,
			UTCMinute:   0,
			DayOfWeek:   int(time.Monday),
			Label:       "Evening",
			IsActive:    true,
			CreatedAt:   time.Date(2026, 6, 30, 17, 5, 0, 0, time.UTC),
		},
	}).Exec(context.Background())
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.Post{
		ID:          "post-occupied-slot",
		WorkspaceID: "ws-1",
		CreatedByID: "user-1",
		Content:     "Already using the morning slot",
		Status:      statusScheduled,
		ScheduledAt: time.Date(2026, 7, 6, 9, 0, 0, 0, time.UTC),
		CreatedAt:   time.Date(2026, 6, 30, 18, 0, 0, 0, time.UTC),
	}).Exec(context.Background())
	require.NoError(t, err)

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-slot-occupied",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "suggest_next_slot",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"after":        "2026-07-06T08:00:00Z",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	suggestion := structured["suggestion"].(map[string]any)
	require.Equal(t, "2026-07-06T17:00:00Z", suggestion["slot_time"])
	slot := suggestion["slot"].(map[string]any)
	require.Equal(t, "slot-17", slot["id"])
	require.Equal(t, "Evening", slot["label"])
}

func TestMCPCallUploadMediaFromURLStoresMedia(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	srv.handler.SetMediaURLValidator(func(_ context.Context, remote *url.URL) error {
		require.Equal(t, "https", remote.Scheme)
		require.Equal(t, "cdn.example", remote.Hostname())
		return nil
	})
	srv.handler.SetMediaURLHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://cdn.example/launch.txt", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(bytes.NewBufferString("launch media")),
			Request:    req,
		}, nil
	})})

	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-upload-url",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "upload_media_from_url",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"url":          "https://cdn.example/launch.txt",
				"alt_text":     "Launch text asset",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	result := out["result"].(map[string]any)
	structured := result["structuredContent"].(map[string]any)
	media := structured["media"].(map[string]any)
	require.NotEmpty(t, media["id"])
	require.Equal(t, "text/plain; charset=utf-8", media["mime_type"])
	require.Equal(t, "/media/"+media["id"].(string), media["url"])
	require.Equal(t, "launch.txt", media["filename"])
	require.Equal(t, "Launch text asset", media["alt_text"])
	require.Equal(t, "https://cdn.example/launch.txt", media["source_url"])
	require.Equal(t, false, media["deduped"])

	var stored models.MediaAttachment
	require.NoError(t, srv.db.NewSelect().Model(&stored).Where("id = ?", media["id"]).Scan(context.Background()))
	require.Equal(t, "ws-1", stored.WorkspaceID)
	require.Equal(t, "launch.txt", stored.OriginalFilename)
	require.Equal(t, "Launch text asset", stored.AltText)
	require.Equal(t, int64(len("launch media")), stored.Size)
}

func TestMCPCallUploadMediaFromURLRejectsLocalhost(t *testing.T) {
	t.Parallel()

	srv := newMCPTestServer(t)
	resp := srv.request(t, "web-token", map[string]any{
		"jsonrpc": "2.0",
		"id":      "call-upload-localhost",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "upload_media_from_url",
			"arguments": map[string]any{
				"workspace_id": "ws-1",
				"url":          "http://127.0.0.1/private.png",
			},
		},
	})

	require.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	rpcErr := out["error"].(map[string]any)
	require.Contains(t, rpcErr["message"], "private or local address")
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("media_attachments").Scan(context.Background(), &count))
	require.Equal(t, 0, count)
}
