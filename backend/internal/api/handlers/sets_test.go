package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestCreateSetReturnsInitialAccounts(t *testing.T) {
	t.Parallel()

	srv := newSetsTestServer(t)
	resp := srv.request(t, http.MethodPost, "/api/v1/sets", map[string]any{
		"workspace_id": "ws-1",
		"name":         "Launch",
		"is_default":   true,
		"account_ids":  []string{"acc-1", "acc-2"},
	})
	require.Equal(t, http.StatusOK, resp.Code)

	var out SetResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "Launch", out.Name)
	require.True(t, out.IsDefault)
	require.Len(t, out.Accounts, 2)
	require.Equal(t, "acc-2", out.Accounts[0].SocialAccountID)
	require.Equal(t, "bluesky", out.Accounts[0].Platform)
	require.Equal(t, "acc-1", out.Accounts[1].SocialAccountID)
	require.Equal(t, "x", out.Accounts[1].Platform)
}

type setsTestServer struct {
	echo *echo.Echo
	db   *bun.DB
}

func newSetsTestServer(t *testing.T) *setsTestServer {
	t.Helper()

	db := createHandlerTestDB(t,
		(*models.WorkspaceMember)(nil),
		(*models.SocialAccount)(nil),
		(*models.SocialMediaSet)(nil),
		(*models.SocialMediaSetAccount)(nil),
	)
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        "admin",
	}).Exec(ctx)
	require.NoError(t, err)
	accounts := []models.SocialAccount{
		{ID: "acc-1", WorkspaceID: "ws-1", Platform: "x", AccountID: "1", AccountUsername: "rodrigo", AccessTokenEnc: []byte("token"), IsActive: true},
		{ID: "acc-2", WorkspaceID: "ws-1", Platform: "bluesky", AccountID: "2", AccountUsername: "rodrigo.bsky.social", AccessTokenEnc: []byte("token"), IsActive: true},
	}
	_, err = db.NewInsert().Model(&accounts).Exec(ctx)
	require.NoError(t, err)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	handler := NewSetHandler(db, testAuthenticator{})
	handler.CreateSet(api)

	return &setsTestServer{echo: e, db: db}
}

func (s *setsTestServer) request(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload bytes.Buffer
	require.NoError(t, json.NewEncoder(&payload).Encode(body))
	req := httptest.NewRequestWithContext(t.Context(), method, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}
