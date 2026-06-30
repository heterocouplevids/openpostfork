package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openpost/backend/internal/models"
	"github.com/uptrace/bun"
)

const ProviderPolar = "polar"

type Service struct {
	db            *bun.DB
	webhookSecret string
	now           func() time.Time
}

func NewService(db *bun.DB, webhookSecret string) *Service {
	return &Service{
		db:            db,
		webhookSecret: strings.TrimSpace(webhookSecret),
		now:           func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetNowForTest(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

type WebhookHeaders struct {
	ID        string
	Timestamp string
	Signature string
}

type WebhookResult struct {
	EventID     string
	EventType   string
	WorkspaceID string
	Duplicate   bool
}

func (s *Service) ProcessPolarWebhook(ctx context.Context, body []byte, headers WebhookHeaders) (WebhookResult, error) {
	if err := s.verifyStandardWebhook(body, headers); err != nil {
		return WebhookResult{}, err
	}

	var event polarEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return WebhookResult{}, fmt.Errorf("invalid webhook payload: %w", err)
	}
	if strings.TrimSpace(event.ID) == "" {
		event.ID = headers.ID
	}
	if event.ID == "" || event.Type == "" {
		return WebhookResult{}, fmt.Errorf("webhook event id and type are required")
	}

	result := WebhookResult{EventID: event.ID, EventType: event.Type}
	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		inserted, err := insertWebhookEvent(txCtx, tx, event.ID, event.Type)
		if err != nil {
			return err
		}
		if !inserted {
			result.Duplicate = true
			return nil
		}

		if !strings.HasPrefix(event.Type, "subscription.") {
			return nil
		}

		subscription, err := subscriptionFromPolarEvent(event)
		if err != nil {
			return err
		}
		result.WorkspaceID = subscription.WorkspaceID
		return upsertSubscription(txCtx, tx, subscription)
	})
	return result, err
}

func insertWebhookEvent(ctx context.Context, tx bun.Tx, eventID, eventType string) (bool, error) {
	event := &models.BillingWebhookEvent{
		EventID:     eventID,
		Provider:    ProviderPolar,
		EventType:   eventType,
		ProcessedAt: time.Now().UTC(),
	}
	res, err := tx.NewInsert().
		Model(event).
		On("CONFLICT (event_id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("recording webhook event: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func upsertSubscription(ctx context.Context, tx bun.Tx, subscription *models.BillingSubscription) error {
	_, err := tx.NewInsert().
		Model(subscription).
		On("CONFLICT (workspace_id) DO UPDATE").
		Set("provider = EXCLUDED.provider").
		Set("provider_customer_id = EXCLUDED.provider_customer_id").
		Set("provider_subscription_id = EXCLUDED.provider_subscription_id").
		Set("provider_product_id = EXCLUDED.provider_product_id").
		Set("provider_price_id = EXCLUDED.provider_price_id").
		Set("status = EXCLUDED.status").
		Set("plan_id = EXCLUDED.plan_id").
		Set("entitlement_snapshot = EXCLUDED.entitlement_snapshot").
		Set("current_period_end = EXCLUDED.current_period_end").
		Set("cancel_at_period_end = EXCLUDED.cancel_at_period_end").
		Set("raw_payload = EXCLUDED.raw_payload").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("upserting billing subscription: %w", err)
	}
	return nil
}

func (s *Service) verifyStandardWebhook(body []byte, headers WebhookHeaders) error {
	if s.webhookSecret == "" {
		return fmt.Errorf("polar webhook secret is not configured")
	}
	if headers.ID == "" || headers.Timestamp == "" || headers.Signature == "" {
		return fmt.Errorf("missing webhook signature headers")
	}

	timestamp, err := strconv.ParseInt(headers.Timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook timestamp")
	}
	signedAt := time.Unix(timestamp, 0).UTC()
	if delta := s.now().Sub(signedAt); delta > 5*time.Minute || delta < -5*time.Minute {
		return fmt.Errorf("webhook timestamp outside tolerance")
	}

	secret := decodeWebhookSecret(s.webhookSecret)
	signed := headers.ID + "." + headers.Timestamp + "." + string(body)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signed))
	expected := mac.Sum(nil)

	for _, candidate := range strings.Fields(headers.Signature) {
		if strings.HasPrefix(candidate, "v1,") {
			candidate = strings.TrimPrefix(candidate, "v1,")
		}
		got, err := base64.StdEncoding.DecodeString(candidate)
		if err != nil {
			continue
		}
		if hmac.Equal(got, expected) {
			return nil
		}
	}
	return fmt.Errorf("invalid webhook signature")
}

func decodeWebhookSecret(secret string) []byte {
	secret = strings.TrimSpace(strings.TrimPrefix(secret, "whsec_"))
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil {
		return decoded
	}
	return []byte(secret)
}

type polarEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type polarSubscriptionData struct {
	ID                string         `json:"id"`
	Status            string         `json:"status"`
	CustomerID        string         `json:"customer_id"`
	ProductID         string         `json:"product_id"`
	PriceID           string         `json:"price_id"`
	CurrentPeriodEnd  string         `json:"current_period_end"`
	CancelAtPeriodEnd bool           `json:"cancel_at_period_end"`
	Metadata          map[string]any `json:"metadata"`
	Customer          struct {
		ID         string         `json:"id"`
		ExternalID string         `json:"external_id"`
		Metadata   map[string]any `json:"metadata"`
	} `json:"customer"`
	Product struct {
		ID string `json:"id"`
	} `json:"product"`
	Price struct {
		ID string `json:"id"`
	} `json:"price"`
}

func subscriptionFromPolarEvent(event polarEvent) (*models.BillingSubscription, error) {
	var data polarSubscriptionData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil, fmt.Errorf("invalid subscription payload: %w", err)
	}

	workspaceID := firstMetadataString(data.Metadata, "workspace_id")
	if workspaceID == "" {
		workspaceID = firstMetadataString(data.Customer.Metadata, "workspace_id")
	}
	if workspaceID == "" {
		workspaceID = data.Customer.ExternalID
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("subscription webhook missing workspace_id metadata")
	}

	customerID := data.CustomerID
	if customerID == "" {
		customerID = data.Customer.ID
	}
	productID := data.ProductID
	if productID == "" {
		productID = data.Product.ID
	}
	priceID := data.PriceID
	if priceID == "" {
		priceID = data.Price.ID
	}
	planID := firstMetadataString(data.Metadata, "plan_id")
	if planID == "" {
		planID = firstMetadataString(data.Metadata, "plan")
	}

	entitlementSnapshot := entitlementSnapshotJSON(data, planID)
	currentPeriodEnd := parseOptionalTime(data.CurrentPeriodEnd)
	rawPayload, _ := json.Marshal(json.RawMessage(event.Data))
	now := time.Now().UTC()
	return &models.BillingSubscription{
		WorkspaceID:            workspaceID,
		Provider:               ProviderPolar,
		ProviderCustomerID:     customerID,
		ProviderSubscriptionID: data.ID,
		ProviderProductID:      productID,
		ProviderPriceID:        priceID,
		Status:                 data.Status,
		PlanID:                 planID,
		EntitlementSnapshot:    entitlementSnapshot,
		CurrentPeriodEnd:       currentPeriodEnd,
		CancelAtPeriodEnd:      data.CancelAtPeriodEnd,
		RawPayload:             string(rawPayload),
		CreatedAt:              now,
		UpdatedAt:              now,
	}, nil
}

func entitlementSnapshotJSON(data polarSubscriptionData, planID string) string {
	snapshot := map[string]any{
		"provider":   ProviderPolar,
		"plan_id":    planID,
		"status":     data.Status,
		"product_id": data.ProductID,
		"price_id":   data.PriceID,
	}
	if limits, ok := data.Metadata["limits"]; ok {
		snapshot["limits"] = limits
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func firstMetadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func parseOptionalTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05-07:00"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
