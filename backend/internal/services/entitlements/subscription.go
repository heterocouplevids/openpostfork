package entitlements

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openpost/backend/internal/models"
	"github.com/uptrace/bun"
)

type SubscriptionService struct {
	db       *bun.DB
	fallback Service
}

func NewSubscriptionService(db *bun.DB, fallback Service) *SubscriptionService {
	return &SubscriptionService{db: db, fallback: fallback}
}

type subscriptionSnapshot struct {
	Limits map[LimitKey]int64 `json:"limits"`
}

func (s *SubscriptionService) Check(ctx context.Context, req Request) (Decision, error) {
	if strings.TrimSpace(req.WorkspaceID) == "" {
		if req.Limit == LimitWorkspaces && strings.TrimSpace(req.UserID) != "" {
			decision, found, err := s.checkUserWorkspaceLimit(ctx, req)
			if err != nil {
				return Decision{}, err
			}
			if found {
				return decision, nil
			}
		}
		if s.fallback != nil {
			return s.fallback.Check(ctx, req)
		}
		return Decision{Allowed: false, Current: req.Current, Amount: req.Amount}, fmt.Errorf("workspace id is required")
	}
	if req.Amount <= 0 {
		return Decision{Allowed: false, Current: req.Current, Amount: req.Amount}, fmt.Errorf("entitlement check amount must be positive")
	}

	var sub models.BillingSubscription
	err := s.db.NewSelect().
		Model(&sub).
		Where("workspace_id = ?", req.WorkspaceID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return Decision{
			Allowed: false,
			Current: req.Current,
			Amount:  req.Amount,
			Reason:  "active subscription required",
		}, nil
	}
	if err != nil {
		return Decision{}, fmt.Errorf("loading subscription: %w", err)
	}
	if !subscriptionStatusAllowsUsage(sub.Status) {
		return Decision{
			Allowed: false,
			Current: req.Current,
			Amount:  req.Amount,
			Reason:  "active subscription required",
		}, nil
	}

	snapshot, err := parseSubscriptionSnapshot(sub.EntitlementSnapshot)
	if err != nil {
		return Decision{}, err
	}
	static := NewStaticService(PlanSnapshot{
		PlanID: sub.PlanID,
		Limits: snapshot.Limits,
	})
	return static.Check(ctx, req)
}

func (s *SubscriptionService) checkUserWorkspaceLimit(ctx context.Context, req Request) (Decision, bool, error) {
	var subscriptions []models.BillingSubscription
	err := s.db.NewSelect().
		Model(&subscriptions).
		ModelTableExpr("billing_subscriptions AS bs").
		ColumnExpr("bs.*").
		Join("JOIN workspace_members AS wm ON wm.workspace_id = bs.workspace_id").
		Where("wm.user_id = ?", req.UserID).
		Scan(ctx)
	if err != nil {
		return Decision{}, false, fmt.Errorf("loading user subscriptions: %w", err)
	}

	var maxLimit int64
	foundLimit := false
	for _, sub := range subscriptions {
		if !subscriptionStatusAllowsUsage(sub.Status) {
			continue
		}
		snapshot, err := parseSubscriptionSnapshot(sub.EntitlementSnapshot)
		if err != nil {
			return Decision{}, false, err
		}
		limit, ok := snapshot.Limits[LimitWorkspaces]
		if !ok {
			continue
		}
		if !foundLimit || limit > maxLimit {
			maxLimit = limit
			foundLimit = true
		}
	}
	if !foundLimit {
		return Decision{}, false, nil
	}

	static := NewStaticService(PlanSnapshot{
		PlanID: "user-subscription",
		Limits: map[LimitKey]int64{
			LimitWorkspaces: maxLimit,
		},
	})
	decision, err := static.Check(ctx, req)
	return decision, true, err
}

func subscriptionStatusAllowsUsage(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "trialing":
		return true
	default:
		return false
	}
}

func parseSubscriptionSnapshot(raw string) (subscriptionSnapshot, error) {
	if strings.TrimSpace(raw) == "" {
		return subscriptionSnapshot{Limits: map[LimitKey]int64{}}, nil
	}
	var decoded struct {
		Limits map[string]any `json:"limits"`
	}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return subscriptionSnapshot{}, fmt.Errorf("invalid entitlement snapshot: %w", err)
	}
	limits := make(map[LimitKey]int64, len(decoded.Limits))
	for key, value := range decoded.Limits {
		amount, ok := limitValueAsInt64(value)
		if !ok {
			return subscriptionSnapshot{}, fmt.Errorf("invalid limit value for %s", key)
		}
		limits[LimitKey(key)] = amount
	}
	return subscriptionSnapshot{Limits: limits}, nil
}

func limitValueAsInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), typed >= 0 && typed == float64(int64(typed))
	case int64:
		return typed, typed >= 0
	case int:
		return int64(typed), typed >= 0
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil && parsed >= 0
	default:
		return 0, false
	}
}
