package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func newBillingTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString()))
	require.NoError(t, err)
	sqldb.SetMaxOpenConns(1)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	for _, model := range []interface{}{
		(*models.Workspace)(nil),
		(*models.BillingSubscription)(nil),
		(*models.BillingWebhookEvent)(nil),
	} {
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(context.Background())
		require.NoError(t, err)
	}
	_, err = db.NewInsert().Model(&models.Workspace{ID: "ws-1", Name: "Launch"}).Exec(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db
}

func signedWebhookHeaders(t *testing.T, secret string, now time.Time, eventID string, body []byte) WebhookHeaders {
	t.Helper()

	timestamp := fmt.Sprintf("%d", now.Unix())
	mac := hmac.New(sha256.New, decodeWebhookSecret(secret))
	_, _ = mac.Write([]byte(eventID + "." + timestamp + "." + string(body)))
	return WebhookHeaders{
		ID:        eventID,
		Timestamp: timestamp,
		Signature: "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil)),
	}
}

func TestProcessPolarWebhookUpsertsSubscription(t *testing.T) {
	t.Parallel()

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("secret"))
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	db := newBillingTestDB(t)
	service := NewService(db, secret)
	service.now = func() time.Time { return now }
	body := []byte(`{
		"id": "evt-1",
		"type": "subscription.active",
		"data": {
			"id": "sub-1",
			"customer_id": "cus-1",
			"product_id": "prod-creator",
			"price_id": "price-monthly",
			"status": "active",
			"current_period_end": "2026-07-30T12:00:00Z",
			"metadata": {
				"workspace_id": "ws-1",
				"plan_id": "creator",
				"limits": {"scheduled_posts_monthly": 500}
			}
		}
	}`)

	result, err := service.ProcessPolarWebhook(context.Background(), body, signedWebhookHeaders(t, secret, now, "evt-1", body))

	require.NoError(t, err)
	require.False(t, result.Duplicate)
	require.Equal(t, "ws-1", result.WorkspaceID)

	var sub models.BillingSubscription
	require.NoError(t, db.NewSelect().Model(&sub).Where("workspace_id = ?", "ws-1").Scan(context.Background()))
	require.Equal(t, "polar", sub.Provider)
	require.Equal(t, "cus-1", sub.ProviderCustomerID)
	require.Equal(t, "sub-1", sub.ProviderSubscriptionID)
	require.Equal(t, "prod-creator", sub.ProviderProductID)
	require.Equal(t, "price-monthly", sub.ProviderPriceID)
	require.Equal(t, "active", sub.Status)
	require.Equal(t, "creator", sub.PlanID)
	require.Contains(t, sub.EntitlementSnapshot, "scheduled_posts_monthly")
	require.Equal(t, time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), sub.CurrentPeriodEnd)
}

func TestProcessPolarWebhookDeduplicatesEvents(t *testing.T) {
	t.Parallel()

	secret := "plain-secret"
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	db := newBillingTestDB(t)
	service := NewService(db, secret)
	service.now = func() time.Time { return now }
	body := []byte(`{
		"id": "evt-1",
		"type": "subscription.active",
		"data": {
			"id": "sub-1",
			"customer_id": "cus-1",
			"status": "active",
			"metadata": {"workspace_id": "ws-1", "plan_id": "starter"}
		}
	}`)
	headers := signedWebhookHeaders(t, secret, now, "evt-1", body)

	first, err := service.ProcessPolarWebhook(context.Background(), body, headers)
	require.NoError(t, err)
	second, err := service.ProcessPolarWebhook(context.Background(), body, headers)
	require.NoError(t, err)

	require.False(t, first.Duplicate)
	require.True(t, second.Duplicate)
	var eventCount int
	require.NoError(t, db.NewSelect().ColumnExpr("COUNT(*)").TableExpr("billing_webhook_events").Scan(context.Background(), &eventCount))
	require.Equal(t, 1, eventCount)
}

func TestProcessPolarWebhookRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	service := NewService(newBillingTestDB(t), "secret")
	service.now = func() time.Time { return now }
	body := []byte(`{"id":"evt-1","type":"subscription.active","data":{}}`)

	_, err := service.ProcessPolarWebhook(context.Background(), body, WebhookHeaders{
		ID:        "evt-1",
		Timestamp: fmt.Sprintf("%d", now.Unix()),
		Signature: "v1," + base64.StdEncoding.EncodeToString([]byte("wrong")),
	})

	require.ErrorContains(t, err, "invalid webhook signature")
}

func TestProcessPolarWebhookRejectsStaleTimestamp(t *testing.T) {
	t.Parallel()

	secret := "secret"
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	service := NewService(newBillingTestDB(t), secret)
	service.now = func() time.Time { return now }
	body := []byte(`{"id":"evt-1","type":"subscription.active","data":{}}`)

	_, err := service.ProcessPolarWebhook(context.Background(), body, signedWebhookHeaders(t, secret, now.Add(-10*time.Minute), "evt-1", body))

	require.ErrorContains(t, err, "outside tolerance")
}
