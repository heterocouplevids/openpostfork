package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/billing"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/usage"
	"github.com/uptrace/bun"
)

type BillingHandler struct {
	billing *billing.Service
	db      *bun.DB
	auth    middleware.Authenticator
	usage   *usage.Service
}

func NewBillingHandler(billingService *billing.Service, deps ...any) *BillingHandler {
	handler := &BillingHandler{billing: billingService}
	if len(deps) > 0 {
		if db, ok := deps[0].(*bun.DB); ok {
			handler.db = db
			handler.usage = usage.NewService(db)
		}
	}
	if len(deps) > 1 {
		if auth, ok := deps[1].(middleware.Authenticator); ok {
			handler.auth = auth
		}
	}
	return handler
}

type PolarWebhookOutput struct {
	OK        bool   `json:"ok"`
	Duplicate bool   `json:"duplicate"`
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
}

func (h *BillingHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/v1/billing/polar/webhook", h.handlePolarWebhook)
}

func (h *BillingHandler) RegisterAPIRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-billing-status",
		Method:      http.MethodGet,
		Path:        "/billing/status",
		Summary:     "Get billing status",
		Tags:        []string{"Billing"},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403},
	}, h.getStatus)

	huma.Register(api, huma.Operation{
		OperationID: "create-billing-checkout",
		Method:      http.MethodPost,
		Path:        "/billing/checkout",
		Summary:     "Create billing checkout",
		Tags:        []string{"Billing"},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403, 503},
	}, h.createCheckout)

	huma.Register(api, huma.Operation{
		OperationID: "create-billing-portal-session",
		Method:      http.MethodPost,
		Path:        "/billing/portal",
		Summary:     "Create billing portal session",
		Tags:        []string{"Billing"},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403, 503},
	}, h.createPortalSession)
}

func (h *BillingHandler) handlePolarWebhook(c echo.Context) error {
	if h.billing == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{fieldError: "billing service is not configured"})
	}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{fieldError: "failed to read webhook body"})
	}
	result, err := h.billing.ProcessPolarWebhook(c.Request().Context(), body, billing.WebhookHeaders{
		ID:        c.Request().Header.Get("webhook-id"),
		Timestamp: c.Request().Header.Get("webhook-timestamp"),
		Signature: c.Request().Header.Get("webhook-signature"),
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{fieldError: err.Error()})
	}

	return c.JSON(http.StatusOK, PolarWebhookOutput{
		OK:        true,
		Duplicate: result.Duplicate,
		EventID:   result.EventID,
		EventType: result.EventType,
	})
}

type GetBillingStatusInput struct {
	WorkspaceID string `query:"workspace_id" doc:"Workspace ID"`
}

type BillingStatusResponse struct {
	WorkspaceID       string           `json:"workspace_id" doc:"Workspace ID"`
	Provider          string           `json:"provider,omitempty" doc:"Billing provider"`
	Status            string           `json:"status" doc:"Subscription status"`
	PlanID            string           `json:"plan_id,omitempty" doc:"Plan ID"`
	CurrentPeriodEnd  string           `json:"current_period_end,omitempty" doc:"Current billing period end"`
	CancelAtPeriodEnd bool             `json:"cancel_at_period_end" doc:"Whether the subscription cancels at period end"`
	Limits            map[string]int64 `json:"limits" doc:"Entitlement limits from the local subscription snapshot"`
	Usage             map[string]int64 `json:"usage" doc:"Current-month usage counters"`
	PeriodStart       string           `json:"period_start" doc:"UTC month start for the usage counters"`
}

type BillingStatusOutput struct {
	Body BillingStatusResponse
}

func (h *BillingHandler) getStatus(ctx context.Context, input *GetBillingStatusInput) (*BillingStatusOutput, error) {
	userID := middleware.GetUserID(ctx)
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	if err := h.checkWorkspaceAccess(ctx, input.WorkspaceID, userID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	usageSnapshot, err := h.usage.SnapshotMonthly(ctx, input.WorkspaceID, now)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to load billing usage")
	}
	response := BillingStatusResponse{
		WorkspaceID: input.WorkspaceID,
		Status:      "none",
		Limits:      map[string]int64{},
		Usage:       usageSnapshotToStrings(usageSnapshot),
		PeriodStart: usage.MonthStart(now).Format(time.RFC3339),
	}

	var sub models.BillingSubscription
	err = h.db.NewSelect().
		Model(&sub).
		Where("workspace_id = ?", input.WorkspaceID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return &BillingStatusOutput{Body: response}, nil
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to load billing subscription")
	}

	response.Provider = sub.Provider
	response.Status = sub.Status
	response.PlanID = sub.PlanID
	response.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
	if !sub.CurrentPeriodEnd.IsZero() {
		response.CurrentPeriodEnd = sub.CurrentPeriodEnd.UTC().Format(time.RFC3339)
	}
	limits, err := limitsFromSnapshot(sub.EntitlementSnapshot)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to parse billing entitlements")
	}
	response.Limits = limits
	return &BillingStatusOutput{Body: response}, nil
}

type CreateBillingCheckoutInput struct {
	Body struct {
		WorkspaceID string `json:"workspace_id" doc:"Workspace ID"`
		PlanID      string `json:"plan_id" doc:"Plan ID: starter, creator, or pro"`
	}
}

type BillingURLResponse struct {
	URL string `json:"url" doc:"Redirect URL"`
	ID  string `json:"id,omitempty" doc:"Provider object ID"`
}

type BillingURLOutput struct {
	Body BillingURLResponse
}

func (h *BillingHandler) createCheckout(ctx context.Context, input *CreateBillingCheckoutInput) (*BillingURLOutput, error) {
	userID := middleware.GetUserID(ctx)
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	if err := h.checkWorkspaceAccess(ctx, input.Body.WorkspaceID, userID); err != nil {
		return nil, err
	}
	email, err := h.userEmail(ctx, userID)
	if err != nil {
		return nil, err
	}

	result, err := h.billing.CreateCheckout(ctx, billing.CreateCheckoutInput{
		WorkspaceID:   input.Body.WorkspaceID,
		UserID:        userID,
		CustomerEmail: email,
		PlanID:        input.Body.PlanID,
	})
	if err != nil {
		return nil, billingAPIError(err)
	}
	return &BillingURLOutput{Body: BillingURLResponse{ID: result.ID, URL: result.URL}}, nil
}

type CreateBillingPortalInput struct {
	Body struct {
		WorkspaceID string `json:"workspace_id" doc:"Workspace ID"`
	}
}

func (h *BillingHandler) createPortalSession(ctx context.Context, input *CreateBillingPortalInput) (*BillingURLOutput, error) {
	userID := middleware.GetUserID(ctx)
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	if err := h.checkWorkspaceAccess(ctx, input.Body.WorkspaceID, userID); err != nil {
		return nil, err
	}

	result, err := h.billing.CreateCustomerPortalSession(ctx, input.Body.WorkspaceID)
	if err != nil {
		return nil, billingAPIError(err)
	}
	return &BillingURLOutput{Body: BillingURLResponse{ID: result.ID, URL: result.URL}}, nil
}

func billingAPIError(err error) error {
	if billing.IsConfigurationError(err) {
		return huma.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return huma.Error400BadRequest(err.Error())
}

func (h *BillingHandler) ensureReady() error {
	if h.billing == nil || h.db == nil || h.auth == nil || h.usage == nil {
		return huma.Error500InternalServerError("billing API is not configured")
	}
	return nil
}

func (h *BillingHandler) checkWorkspaceAccess(ctx context.Context, workspaceID, userID string) error {
	if !middleware.WorkspaceScopeAllows(ctx, workspaceID) {
		return huma.Error403Forbidden("workspace not accessible")
	}
	count, err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return huma.Error500InternalServerError("failed to check workspace access")
	}
	if count == 0 {
		return huma.Error403Forbidden("workspace not accessible")
	}
	return nil
}

func (h *BillingHandler) userEmail(ctx context.Context, userID string) (string, error) {
	var user models.User
	err := h.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return "", huma.Error403Forbidden("user not found")
	}
	if err != nil {
		return "", huma.Error500InternalServerError("failed to load user")
	}
	return user.Email, nil
}

func usageSnapshotToStrings(snapshot map[entitlements.LimitKey]int64) map[string]int64 {
	out := make(map[string]int64, len(snapshot))
	for key, value := range snapshot {
		out[string(key)] = value
	}
	return out
}

func limitsFromSnapshot(raw string) (map[string]int64, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]int64{}, nil
	}
	var decoded struct {
		Limits map[string]any `json:"limits"`
	}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, err
	}
	limits := make(map[string]int64, len(decoded.Limits))
	for key, value := range decoded.Limits {
		amount, ok := snapshotValueAsInt64(value)
		if !ok {
			return nil, fmt.Errorf("invalid limit value for %s", key)
		}
		limits[key] = amount
	}
	return limits, nil
}

func snapshotValueAsInt64(value any) (int64, bool) {
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
