package handlers

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/services/billing"
)

type BillingHandler struct {
	billing *billing.Service
}

func NewBillingHandler(billingService *billing.Service) *BillingHandler {
	return &BillingHandler{billing: billingService}
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

func (h *BillingHandler) handlePolarWebhook(c echo.Context) error {
	if h.billing == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "billing service is not configured"})
	}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read webhook body"})
	}
	result, err := h.billing.ProcessPolarWebhook(c.Request().Context(), body, billing.WebhookHeaders{
		ID:        c.Request().Header.Get("webhook-id"),
		Timestamp: c.Request().Header.Get("webhook-timestamp"),
		Signature: c.Request().Header.Get("webhook-signature"),
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, PolarWebhookOutput{
		OK:        true,
		Duplicate: result.Duplicate,
		EventID:   result.EventID,
		EventType: result.EventType,
	})
}
