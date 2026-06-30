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
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type workspaceTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newWorkspaceTestServer(t *testing.T, entitlement entitlements.Service) *workspaceTestServer {
	t.Helper()

	db := createHandlerTestDB(t, (*models.Workspace)(nil), (*models.WorkspaceMember)(nil))
	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	handler := NewWorkspaceHandler(db, testAuthenticator{}, entitlement)
	handler.CreateWorkspace(api)

	return &workspaceTestServer{echo: e, db: db}
}

func (s *workspaceTestServer) createWorkspace(t *testing.T, name string) *httptest.ResponseRecorder {
	t.Helper()

	var payload bytes.Buffer
	require.NoError(t, json.NewEncoder(&payload).Encode(map[string]string{"name": name}))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/workspaces", &payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func TestCreateWorkspaceAllowsSelfHostedDefault(t *testing.T) {
	t.Parallel()

	srv := newWorkspaceTestServer(t, nil)
	resp := srv.createWorkspace(t, "Launch")

	require.Equal(t, http.StatusOK, resp.Code)
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("workspaces").Scan(context.Background(), &count))
	require.Equal(t, 1, count)
}

func TestCreateWorkspaceCloudBootstrapAllowsFirstWorkspaceOnly(t *testing.T) {
	t.Parallel()

	srv := newWorkspaceTestServer(t, entitlements.NewCloudBootstrapService())

	first := srv.createWorkspace(t, "Launch")
	require.Equal(t, http.StatusOK, first.Code)

	second := srv.createWorkspace(t, "Second")
	require.Equal(t, http.StatusPaymentRequired, second.Code)
	require.Contains(t, second.Body.String(), "workspaces limit exceeded")

	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("workspaces").Scan(context.Background(), &count))
	require.Equal(t, 1, count)
}

func TestCreateWorkspaceRejectsWhenEntitlementLimitExceeded(t *testing.T) {
	t.Parallel()

	entitlement := entitlements.NewStaticService(entitlements.PlanSnapshot{
		PlanID: "starter",
		Limits: map[entitlements.LimitKey]int64{
			entitlements.LimitWorkspaces: 1,
		},
	})
	srv := newWorkspaceTestServer(t, entitlement)
	ctx := context.Background()
	_, err := srv.db.NewInsert().Model(&models.Workspace{
		ID:        "existing-ws",
		Name:      "Existing",
		CreatedAt: time.Now().UTC(),
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = srv.db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "existing-ws",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)

	resp := srv.createWorkspace(t, "Blocked")

	require.Equal(t, http.StatusPaymentRequired, resp.Code)
	require.Contains(t, resp.Body.String(), "workspaces limit exceeded")
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("workspaces").Scan(ctx, &count))
	require.Equal(t, 1, count)
}
