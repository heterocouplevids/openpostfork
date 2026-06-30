package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/uptrace/bun"
)

const (
	mcpProtocolVersion = "2025-06-18"
	mcpToolWorkspaces  = "list_workspaces"
)

type MCPHandler struct {
	db   *bun.DB
	auth middleware.Authenticator
}

func NewMCPHandler(db *bun.DB, authenticator middleware.Authenticator) *MCPHandler {
	return &MCPHandler{db: db, auth: authenticator}
}

func (h *MCPHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/mcp", h.handle)
}

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *mcpError `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (h *MCPHandler) handle(c echo.Context) error {
	principal, err := h.authenticate(c.Request())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req mcpRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, mcpResponse{
			JSONRPC: "2.0",
			Error:   &mcpError{Code: -32700, Message: "parse error"},
		})
	}
	if req.JSONRPC != "2.0" || req.Method == "" {
		return c.JSON(http.StatusOK, mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32600, Message: "invalid request"},
		})
	}

	result, rpcErr := h.dispatch(c.Request().Context(), principal, req)
	resp := mcpResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *MCPHandler) authenticate(r *http.Request) (*middleware.Principal, error) {
	authHeader := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("missing bearer token")
	}
	return h.auth.AuthenticateBearer(r.Context(), token)
}

func (h *MCPHandler) dispatch(ctx context.Context, principal *middleware.Principal, req mcpRequest) (any, *mcpError) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"serverInfo": map[string]string{
				"name":    "openpost",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{"listChanged": false},
			},
		}, nil
	case "tools/list":
		return map[string]any{"tools": []map[string]any{mcpListWorkspacesTool()}}, nil
	case "tools/call":
		return h.callTool(ctx, principal, req.Params)
	default:
		return nil, &mcpError{Code: -32601, Message: "method not found"}
	}
}

func mcpListWorkspacesTool() map[string]any {
	return map[string]any{
		"name":        mcpToolWorkspaces,
		"title":       "List workspaces",
		"description": "List OpenPost workspaces available to the authenticated user.",
		"inputSchema": map[string]any{
			"type":                 "object",
			"properties":           map[string]any{},
			"additionalProperties": false,
		},
	}
}

func (h *MCPHandler) callTool(ctx context.Context, principal *middleware.Principal, raw json.RawMessage) (any, *mcpError) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid tool call params"}
	}
	switch params.Name {
	case mcpToolWorkspaces:
		return h.listWorkspaces(ctx, principal.UserID)
	default:
		return nil, &mcpError{Code: -32602, Message: "unknown tool"}
	}
}

type mcpWorkspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func (h *MCPHandler) listWorkspaces(ctx context.Context, userID string) (any, *mcpError) {
	var rows []struct {
		models.Workspace `bun:",extend"`
		Role             string `bun:"role"`
	}
	err := h.db.NewSelect().
		Model(&rows).
		ColumnExpr("workspace.*").
		ColumnExpr("wm.role").
		Join("JOIN workspace_members AS wm ON wm.workspace_id = workspace.id").
		Where("wm.user_id = ?", userID).
		OrderExpr("workspace.created_at ASC").
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, &mcpError{Code: -32603, Message: "failed to list workspaces"}
	}

	workspaces := make([]mcpWorkspace, 0, len(rows))
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, mcpWorkspace{
			ID:        row.ID,
			Name:      row.Name,
			Role:      row.Role,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
		})
		names = append(names, row.Name)
	}
	text := "No workspaces available."
	if len(names) > 0 {
		text = "Available workspaces: " + strings.Join(names, ", ")
	}

	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: text,
		}},
		"structuredContent": map[string]any{
			"workspaces": workspaces,
		},
	}, nil
}
