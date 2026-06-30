package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/billing"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type billingTestServer struct {
	echo *echo.Echo
	db   *bun.DB
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
