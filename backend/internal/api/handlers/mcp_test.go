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
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type mcpTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newMCPTestServer(t *testing.T) *mcpTestServer {
	t.Helper()

	db := createHandlerTestDB(t, (*models.Workspace)(nil), (*models.WorkspaceMember)(nil))
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

	e := echo.New()
	NewMCPHandler(db, testAuthenticator{}).RegisterRoutes(e)
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
	require.Len(t, tools, 1)
	require.Equal(t, "list_workspaces", tools[0].(map[string]any)["name"])
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
