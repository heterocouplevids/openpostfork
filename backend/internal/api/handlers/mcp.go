package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/uptrace/bun"
)

const (
	mcpProtocolVersion = "2025-06-18"
	mcpToolWorkspaces  = "list_workspaces"
	mcpToolAccounts    = "list_accounts"
	mcpToolCreateDraft = "create_draft"
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
		return map[string]any{"tools": []map[string]any{
			mcpListWorkspacesTool(),
			mcpListAccountsTool(),
			mcpCreateDraftTool(),
		}}, nil
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

func mcpListAccountsTool() map[string]any {
	return map[string]any{
		"name":        mcpToolAccounts,
		"title":       "List social accounts",
		"description": "List active social accounts connected to an OpenPost workspace.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
			},
			"required":             []string{"workspace_id"},
			"additionalProperties": false,
		},
	}
}

func mcpCreateDraftTool() map[string]any {
	return map[string]any{
		"name":        mcpToolCreateDraft,
		"title":       "Create draft",
		"description": "Create an OpenPost draft in a workspace, optionally assigning destination social accounts.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Draft post content.",
				},
				"social_account_ids": map[string]any{
					"type":        "array",
					"description": "Optional destination account IDs returned by list_accounts.",
					"items":       map[string]any{"type": "string"},
				},
			},
			"required":             []string{"workspace_id", "content"},
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
	case mcpToolAccounts:
		return h.listAccounts(ctx, principal.UserID, params.Arguments)
	case mcpToolCreateDraft:
		return h.createDraft(ctx, principal.UserID, params.Arguments)
	default:
		return nil, &mcpError{Code: -32602, Message: "unknown tool"}
	}
}

func decodeMCPArguments(args map[string]any, dest any) error {
	payload, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dest)
}

func (h *MCPHandler) ensureWorkspaceAccess(ctx context.Context, userID, workspaceID string) *mcpError {
	if strings.TrimSpace(workspaceID) == "" {
		return &mcpError{Code: -32602, Message: "workspace_id is required"}
	}
	count, err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return &mcpError{Code: -32603, Message: "failed to check workspace access"}
	}
	if count == 0 {
		return &mcpError{Code: -32602, Message: "workspace not accessible"}
	}
	return nil
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

type mcpAccount struct {
	ID              string `json:"id"`
	Platform        string `json:"platform"`
	Slug            string `json:"slug"`
	AccountID       string `json:"account_id"`
	AccountUsername string `json:"account_username,omitempty"`
	InstanceURL     string `json:"instance_url,omitempty"`
}

func (h *MCPHandler) listAccounts(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid list_accounts arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}

	var rows []models.SocialAccount
	err := h.db.NewSelect().
		Model(&rows).
		Where("workspace_id = ?", input.WorkspaceID).
		Where("is_active = ?", true).
		OrderExpr("platform ASC, slug ASC").
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, &mcpError{Code: -32603, Message: "failed to list accounts"}
	}

	accounts := make([]mcpAccount, 0, len(rows))
	labels := make([]string, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, mcpAccount{
			ID:              row.ID,
			Platform:        row.Platform,
			Slug:            row.Slug,
			AccountID:       row.AccountID,
			AccountUsername: row.AccountUsername,
			InstanceURL:     row.InstanceURL,
		})
		labels = append(labels, row.Platform+":"+row.Slug)
	}

	text := "No active social accounts connected."
	if len(labels) > 0 {
		text = "Active social accounts: " + strings.Join(labels, ", ")
	}
	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: text,
		}},
		"structuredContent": map[string]any{
			"accounts": accounts,
		},
	}, nil
}

func (h *MCPHandler) createDraft(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	var input struct {
		WorkspaceID      string   `json:"workspace_id"`
		Content          string   `json:"content"`
		SocialAccountIDs []string `json:"social_account_ids"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid create_draft arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, &mcpError{Code: -32602, Message: "content is required"}
	}
	if rpcErr := h.ensureActiveAccounts(ctx, input.WorkspaceID, input.SocialAccountIDs); rpcErr != nil {
		return nil, rpcErr
	}

	now := time.Now().UTC()
	post := &models.Post{
		ID:          newUUID(),
		WorkspaceID: input.WorkspaceID,
		CreatedByID: userID,
		Content:     input.Content,
		Status:      statusDraft,
		CreatedAt:   now,
	}
	destinations := make([]models.PostDestination, 0, len(input.SocialAccountIDs))
	for _, accountID := range input.SocialAccountIDs {
		destinations = append(destinations, models.PostDestination{
			ID:              newUUID(),
			PostID:          post.ID,
			SocialAccountID: accountID,
			Status:          postStatusPending,
		})
	}

	err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(post).Exec(txCtx); err != nil {
			return err
		}
		if len(destinations) > 0 {
			if _, err := tx.NewInsert().Model(&destinations).Exec(txCtx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, &mcpError{Code: -32603, Message: "failed to create draft"}
	}

	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: "Draft created: " + post.ID,
		}},
		"structuredContent": map[string]any{
			"post": map[string]any{
				"id":                 post.ID,
				"workspace_id":       post.WorkspaceID,
				"status":             post.Status,
				"social_account_ids": input.SocialAccountIDs,
				"created_at":         post.CreatedAt.Format(time.RFC3339),
			},
		},
	}, nil
}

func (h *MCPHandler) ensureActiveAccounts(ctx context.Context, workspaceID string, accountIDs []string) *mcpError {
	if len(accountIDs) == 0 {
		return nil
	}
	unique := make([]string, 0, len(accountIDs))
	seen := make(map[string]struct{}, len(accountIDs))
	for _, accountID := range accountIDs {
		accountID = strings.TrimSpace(accountID)
		if accountID == "" {
			return &mcpError{Code: -32602, Message: "social_account_ids cannot contain empty values"}
		}
		if _, ok := seen[accountID]; ok {
			continue
		}
		seen[accountID] = struct{}{}
		unique = append(unique, accountID)
	}
	count, err := h.db.NewSelect().
		Model((*models.SocialAccount)(nil)).
		Where("workspace_id = ?", workspaceID).
		Where("is_active = ?", true).
		Where("id IN (?)", bun.List(unique)).
		Count(ctx)
	if err != nil {
		return &mcpError{Code: -32603, Message: "failed to validate social accounts"}
	}
	if count != len(unique) {
		return &mcpError{Code: -32602, Message: "one or more social accounts are invalid, disconnected, or outside this workspace"}
	}
	return nil
}

func newUUID() string {
	return uuid.New().String()
}
