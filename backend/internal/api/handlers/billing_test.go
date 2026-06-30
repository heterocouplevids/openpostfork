package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/billing"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type billingTestServer struct {
	echo   *echo.Echo
	db     *bun.DB
	client *billingHTTPClient
}

type billingHTTPClient struct {
	t        *testing.T
	requests []billingHTTPRequest
	response string
	status   int
}

type billingHTTPRequest struct {
	Path string
	Body map[string]any
}

func (c *billingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.t.Helper()

	var body map[string]any
	require.NoError(c.t, json.NewDecoder(req.Body).Decode(&body))
	c.requests = append(c.requests, billingHTTPRequest{Path: req.URL.Path, Body: body})
	status := c.status
	if status == 0 {
		status = http.StatusCreated
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(c.response)),
		Header:     make(http.Header),
	}, nil
}

func newBillingHandlerTestServer(t *testing.T, secret string, now time.Time) *billingTestServer {
	t.Helper()

	db := createHandlerTestDB(
		t,
		(*models.Workspace)(nil),
		(*models.BillingSubscription)(nil),
		(*models.BillingWebhookEvent)(nil),
	)
	_, err := db.NewInsert().Model(&models.Workspace{ID: "ws-1", Name: "Launch"}).Exec(context.Background())
	require.NoError(t, err)

	e := echo.New()
	service := billing.NewService(db, secret)
	service.SetNowForTest(func() time.Time { return now })
	NewBillingHandler(service).RegisterRoutes(e)
	return &billingTestServer{echo: e, db: db}
}

func newBillingAPITestServer(t *testing.T) *billingTestServer {
	t.Helper()

	db := createHandlerTestDB(
		t,
		(*models.User)(nil),
		(*models.Workspace)(nil),
		(*models.WorkspaceMember)(nil),
	)
	ctx := context.Background()
	_, err := db.NewInsert().Model(&models.User{
		ID:           "user-1",
		Email:        "user@example.com",
		PasswordHash: "hash",
	}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.Workspace{ID: "ws-1", Name: "Launch"}).Exec(ctx)
	require.NoError(t, err)
	_, err = db.NewInsert().Model(&models.WorkspaceMember{
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        models.WorkspaceRoleAdmin,
	}).Exec(ctx)
	require.NoError(t, err)

	client := &billingHTTPClient{t: t}
	service := billing.NewService(db, "", billing.PolarConfig{
		AccessToken: "polar-token",
		APIBaseURL:  "https://api.polar.test",
		SuccessURL:  "https://app.openpost.test/settings/billing?checkout_id={CHECKOUT_ID}",
		ReturnURL:   "https://app.openpost.test/settings/billing",
		Plans: map[string]billing.PlanConfig{
			"creator": {
				ProductID: "product-creator",
				Limits: map[entitlements.LimitKey]int64{
					entitlements.LimitScheduledPostsMonthly: 500,
					entitlements.LimitSocialAccounts:        6,
				},
			},
		},
	})
	service.SetHTTPClientForTest(client)

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	NewBillingHandler(service, db, testAuthenticator{}).RegisterAPIRoutes(api)
	return &billingTestServer{echo: e, db: db, client: client}
}

func (s *billingTestServer) postWebhook(t *testing.T, body []byte, headers billing.WebhookHeaders) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/billing/polar/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("webhook-id", headers.ID)
	req.Header.Set("webhook-timestamp", headers.Timestamp)
	req.Header.Set("webhook-signature", headers.Signature)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func (s *billingTestServer) postJSON(t *testing.T, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload bytes.Buffer
	require.NoError(t, json.NewEncoder(&payload).Encode(body))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func signedPolarWebhookHeaders(t *testing.T, secret string, now time.Time, eventID string, body []byte) billing.WebhookHeaders {
	t.Helper()

	timestamp := fmt.Sprintf("%d", now.Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(eventID + "." + timestamp + "." + string(body)))
	return billing.WebhookHeaders{
		ID:        eventID,
		Timestamp: timestamp,
		Signature: "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil)),
	}
}

func TestPolarWebhookRouteStoresSubscription(t *testing.T) {
	t.Parallel()

	secret := "route-secret"
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	srv := newBillingHandlerTestServer(t, secret, now)
	body := []byte(`{"id":"evt-route","type":"subscription.active","data":{"id":"sub-route","customer_id":"cus-route","status":"active","metadata":{"workspace_id":"ws-1","plan_id":"creator"}}}`)

	resp := srv.postWebhook(t, body, signedPolarWebhookHeaders(t, secret, now, "evt-route", body))

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.True(t, out["ok"].(bool))
	require.Equal(t, "evt-route", out["event_id"])

	var sub models.BillingSubscription
	require.NoError(t, srv.db.NewSelect().Model(&sub).Where("workspace_id = ?", "ws-1").Scan(context.Background()))
	require.Equal(t, "sub-route", sub.ProviderSubscriptionID)
	require.Equal(t, "creator", sub.PlanID)
}

func TestPolarWebhookRouteRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	secret := "route-secret"
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	srv := newBillingHandlerTestServer(t, secret, now)
	body := []byte(`{"id":"evt-route","type":"subscription.active","data":{}}`)

	resp := srv.postWebhook(t, body, billing.WebhookHeaders{
		ID:        "evt-route",
		Timestamp: signedPolarWebhookHeaders(t, secret, now, "evt-route", body).Timestamp,
		Signature: "v1,invalid",
	})

	require.Equal(t, http.StatusUnauthorized, resp.Code, resp.Body.String())
	var count int
	require.NoError(t, srv.db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("billing_subscriptions").Scan(context.Background(), &count))
	require.Equal(t, 0, count)
}

func TestCreateBillingCheckoutRoute(t *testing.T) {
	t.Parallel()

	srv := newBillingAPITestServer(t)
	srv.client.response = `{"id":"checkout-1","url":"https://checkout.polar.test/session"}`

	resp := srv.postJSON(t, "/api/v1/billing/checkout", map[string]any{
		"workspace_id": "ws-1",
		"plan_id":      "creator",
	})

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "checkout-1", out["id"])
	require.Equal(t, "https://checkout.polar.test/session", out["url"])
	require.Len(t, srv.client.requests, 1)
	req := srv.client.requests[0]
	require.Equal(t, "/v1/checkouts/", req.Path)
	require.Equal(t, "user@example.com", req.Body["customer_email"])
	require.Equal(t, "ws-1", req.Body["external_customer_id"])
	metadata := req.Body["metadata"].(map[string]any)
	require.Equal(t, "creator", metadata["plan_id"])
	require.Equal(t, "ws-1", metadata["workspace_id"])
}

func TestCreateBillingPortalRoute(t *testing.T) {
	t.Parallel()

	srv := newBillingAPITestServer(t)
	srv.client.response = `{"id":"session-1","customer_portal_url":"https://polar.test/portal"}`

	resp := srv.postJSON(t, "/api/v1/billing/portal", map[string]any{
		"workspace_id": "ws-1",
	})

	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	require.Equal(t, "session-1", out["id"])
	require.Equal(t, "https://polar.test/portal", out["url"])
	require.Len(t, srv.client.requests, 1)
	req := srv.client.requests[0]
	require.Equal(t, "/v1/customer-sessions/", req.Path)
	require.Equal(t, "ws-1", req.Body["external_customer_id"])
}
