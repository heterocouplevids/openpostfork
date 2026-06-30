package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/openpost/backend/internal/services/usage"
	"github.com/uptrace/bun"
)

const (
	mcpProtocolVersion  = "2025-06-18"
	mcpToolWorkspaces   = "list_workspaces"
	mcpToolAccounts     = "list_accounts"
	mcpToolCreateDraft  = "create_draft"
	mcpToolSchedulePost = "schedule_post"
	mcpToolGetPost      = "get_post_status"
	mcpToolCancelPost   = "cancel_post"
	mcpToolSuggestSlot  = "suggest_next_slot"
	mcpToolUploadURL    = "upload_media_from_url"
	maxRemoteMediaBytes = 50 * 1024 * 1024
)

type MCPHandler struct {
	db                *bun.DB
	auth              middleware.Authenticator
	entitlement       entitlements.Service
	usage             *usage.Service
	mediaStorage      mediastore.BlobStorage
	mediaURLHTTP      *http.Client
	mediaURLValidator func(context.Context, *url.URL) error
}

func NewMCPHandler(db *bun.DB, authenticator middleware.Authenticator, entitlement ...entitlements.Service) *MCPHandler {
	entitlementService := entitlements.Service(entitlements.NewSelfHostedService())
	if len(entitlement) > 0 && entitlement[0] != nil {
		entitlementService = entitlement[0]
	}
	return &MCPHandler{
		db:          db,
		auth:        authenticator,
		entitlement: entitlementService,
		usage:       usage.NewService(db),
	}
}

func (h *MCPHandler) SetUsage(usageService *usage.Service) {
	if usageService != nil {
		h.usage = usageService
	}
}

func (h *MCPHandler) SetMediaStorage(storage mediastore.BlobStorage) {
	h.mediaStorage = storage
}

func (h *MCPHandler) SetMediaURLHTTPClient(client *http.Client) {
	h.mediaURLHTTP = client
}

func (h *MCPHandler) SetMediaURLValidator(validator func(context.Context, *url.URL) error) {
	h.mediaURLValidator = validator
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
			mcpSchedulePostTool(),
			mcpGetPostStatusTool(),
			mcpCancelPostTool(),
			mcpSuggestNextSlotTool(),
			mcpUploadMediaFromURLTool(),
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

func mcpSchedulePostTool() map[string]any {
	return map[string]any{
		"name":        mcpToolSchedulePost,
		"title":       "Schedule post",
		"description": "Create a scheduled OpenPost post and queue it for publishing.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Post content.",
				},
				"scheduled_at": map[string]any{
					"type":        "string",
					"format":      "date-time",
					"description": "Publish time as an RFC3339 timestamp.",
				},
				"social_account_ids": map[string]any{
					"type":        "array",
					"description": "Destination account IDs returned by list_accounts.",
					"items":       map[string]any{"type": "string"},
					"minItems":    1,
				},
			},
			"required":             []string{"workspace_id", "content", "scheduled_at", "social_account_ids"},
			"additionalProperties": false,
		},
	}
}

func mcpGetPostStatusTool() map[string]any {
	return map[string]any{
		"name":        mcpToolGetPost,
		"title":       "Get post status",
		"description": "Read the current OpenPost status and destination status for a post.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"post_id": map[string]any{
					"type":        "string",
					"description": "Post ID returned by create_draft or schedule_post.",
				},
			},
			"required":             []string{"workspace_id", "post_id"},
			"additionalProperties": false,
		},
	}
}

func mcpCancelPostTool() map[string]any {
	return map[string]any{
		"name":        mcpToolCancelPost,
		"title":       "Cancel scheduled post",
		"description": "Cancel a queued scheduled post and leave it as an editable draft.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"post_id": map[string]any{
					"type":        "string",
					"description": "Scheduled post ID returned by schedule_post.",
				},
			},
			"required":             []string{"workspace_id", "post_id"},
			"additionalProperties": false,
		},
	}
}

func mcpSuggestNextSlotTool() map[string]any {
	return map[string]any{
		"name":        mcpToolSuggestSlot,
		"title":       "Suggest next slot",
		"description": "Suggest the next free configured posting slot for a workspace.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"set_id": map[string]any{
					"type":        "string",
					"description": "Optional social media set ID to filter schedules.",
				},
				"after": map[string]any{
					"type":        "string",
					"format":      "date-time",
					"description": "Optional RFC3339 lower bound. Defaults to the current time.",
				},
			},
			"required":             []string{"workspace_id"},
			"additionalProperties": false,
		},
	}
}

func mcpUploadMediaFromURLTool() map[string]any {
	return map[string]any{
		"name":        mcpToolUploadURL,
		"title":       "Upload media from URL",
		"description": "Fetch a public media URL and store it in an OpenPost workspace.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{
					"type":        "string",
					"description": "Workspace ID returned by list_workspaces.",
				},
				"url": map[string]any{
					"type":        "string",
					"format":      "uri",
					"description": "Public http(s) URL to fetch.",
				},
				"filename": map[string]any{
					"type":        "string",
					"description": "Optional filename to store for display and extension detection.",
				},
				"alt_text": map[string]any{
					"type":        "string",
					"description": "Optional accessible alt text for the media.",
				},
			},
			"required":             []string{"workspace_id", "url"},
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
	start := time.Now()
	var (
		result any
		rpcErr *mcpError
	)
	switch params.Name {
	case mcpToolWorkspaces:
		result, rpcErr = h.listWorkspaces(ctx, principal.UserID)
	case mcpToolAccounts:
		result, rpcErr = h.listAccounts(ctx, principal.UserID, params.Arguments)
	case mcpToolCreateDraft:
		result, rpcErr = h.createDraft(ctx, principal.UserID, params.Arguments)
	case mcpToolSchedulePost:
		result, rpcErr = h.schedulePost(ctx, principal.UserID, params.Arguments)
	case mcpToolGetPost:
		result, rpcErr = h.getPostStatus(ctx, principal.UserID, params.Arguments)
	case mcpToolCancelPost:
		result, rpcErr = h.cancelPost(ctx, principal.UserID, params.Arguments)
	case mcpToolSuggestSlot:
		result, rpcErr = h.suggestNextSlot(ctx, principal.UserID, params.Arguments)
	case mcpToolUploadURL:
		result, rpcErr = h.uploadMediaFromURL(ctx, principal.UserID, params.Arguments)
	default:
		rpcErr = &mcpError{Code: -32602, Message: "unknown tool"}
	}
	h.recordToolCall(ctx, principal.UserID, params.Name, workspaceIDFromMCPArguments(params.Arguments), time.Since(start), rpcErr)
	return result, rpcErr
}

func decodeMCPArguments(args map[string]any, dest any) error {
	payload, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dest)
}

func (h *MCPHandler) recordToolCall(ctx context.Context, userID, toolName, workspaceID string, duration time.Duration, rpcErr *mcpError) {
	status := "success"
	errorMessage := ""
	if rpcErr != nil {
		status = "error"
		errorMessage = rpcErr.Message
	}
	_, _ = h.db.NewInsert().Model(&models.MCPToolCall{
		ID:           newUUID(),
		UserID:       userID,
		WorkspaceID:  workspaceID,
		ToolName:     toolName,
		Status:       status,
		ErrorMessage: errorMessage,
		DurationMs:   duration.Milliseconds(),
		CreatedAt:    time.Now().UTC(),
	}).Exec(ctx)
}

func workspaceIDFromMCPArguments(args map[string]any) string {
	if args == nil {
		return ""
	}
	if workspaceID, ok := args["workspace_id"].(string); ok {
		return strings.TrimSpace(workspaceID)
	}
	return ""
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

type mcpPostDestination struct {
	SocialAccountID string `json:"social_account_id"`
	Platform        string `json:"platform"`
	Slug            string `json:"slug"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type mcpPostStatus struct {
	ID                 string               `json:"id"`
	WorkspaceID        string               `json:"workspace_id"`
	Content            string               `json:"content"`
	Status             string               `json:"status"`
	ScheduledAt        string               `json:"scheduled_at,omitempty"`
	ActualRunAt        string               `json:"actual_run_at,omitempty"`
	RandomDelayMinutes int                  `json:"random_delay_minutes"`
	CreatedAt          string               `json:"created_at"`
	Destinations       []mcpPostDestination `json:"destinations"`
}

func (h *MCPHandler) schedulePost(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	var input struct {
		WorkspaceID      string   `json:"workspace_id"`
		Content          string   `json:"content"`
		ScheduledAt      string   `json:"scheduled_at"`
		SocialAccountIDs []string `json:"social_account_ids"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid schedule_post arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, &mcpError{Code: -32602, Message: "content is required"}
	}
	accountIDs, rpcErr := normalizeMCPIDs(input.SocialAccountIDs, "social_account_ids")
	if rpcErr != nil {
		return nil, rpcErr
	}
	if len(accountIDs) == 0 {
		return nil, &mcpError{Code: -32602, Message: "social_account_ids must contain at least one account"}
	}
	if rpcErr := h.ensureActiveAccounts(ctx, input.WorkspaceID, accountIDs); rpcErr != nil {
		return nil, rpcErr
	}
	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		return nil, &mcpError{Code: -32602, Message: "scheduled_at must be an RFC3339 timestamp"}
	}
	if scheduledAt.IsZero() {
		return nil, &mcpError{Code: -32602, Message: "scheduled_at is required"}
	}
	if rpcErr := h.checkScheduledPostQuota(ctx, input.WorkspaceID, 1, scheduledAt); rpcErr != nil {
		return nil, rpcErr
	}

	now := time.Now().UTC()
	post := &models.Post{
		ID:          newUUID(),
		WorkspaceID: input.WorkspaceID,
		CreatedByID: userID,
		Content:     input.Content,
		Status:      statusScheduled,
		ScheduledAt: scheduledAt,
		ActualRunAt: scheduledAt,
		CreatedAt:   now,
	}
	destinations := make([]models.PostDestination, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		destinations = append(destinations, models.PostDestination{
			ID:              newUUID(),
			PostID:          post.ID,
			SocialAccountID: accountID,
			Status:          postStatusPending,
		})
	}
	payload, err := json.Marshal(map[string]string{postIDKey: post.ID})
	if err != nil {
		return nil, &mcpError{Code: -32603, Message: "failed to create publish job payload"}
	}
	job := &models.Job{
		ID:      newUUID(),
		Type:    jobTypePublishPost,
		Payload: string(payload),
		Status:  "pending",
		RunAt:   scheduledAt,
	}

	err = h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(post).Exec(txCtx); err != nil {
			return err
		}
		if _, err := tx.NewInsert().Model(&destinations).Exec(txCtx); err != nil {
			return err
		}
		if _, err := tx.NewInsert().Model(job).Exec(txCtx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, &mcpError{Code: -32603, Message: "failed to schedule post"}
	}
	if rpcErr := h.recordScheduledPostUsage(ctx, input.WorkspaceID, 1, scheduledAt); rpcErr != nil {
		return nil, rpcErr
	}

	postStatus, rpcErr := h.loadMCPPostStatus(ctx, post.ID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return mcpPostToolResult("Post scheduled: "+post.ID, postStatus), nil
}

func (h *MCPHandler) getPostStatus(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	post, rpcErr := h.accessibleMCPPost(ctx, userID, args)
	if rpcErr != nil {
		return nil, rpcErr
	}
	postStatus, rpcErr := h.loadMCPPostStatus(ctx, post.ID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return mcpPostToolResult("Post status: "+post.Status, postStatus), nil
}

func (h *MCPHandler) cancelPost(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	post, rpcErr := h.accessibleMCPPost(ctx, userID, args)
	if rpcErr != nil {
		return nil, rpcErr
	}
	if post.Status == models.PostStatusPublished || post.Status == models.PostStatusPublishing {
		return nil, &mcpError{Code: -32602, Message: "cannot cancel a post that is published or being published"}
	}

	post.Status = statusDraft
	post.ScheduledAt = time.Time{}
	post.ActualRunAt = time.Time{}
	post.RandomDelayMinutes = 0
	err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().
			Model(post).
			Column("status", "scheduled_at", "actual_run_at", "random_delay_minutes").
			Where("id = ?", post.ID).
			Exec(txCtx); err != nil {
			return err
		}
		if _, err := tx.NewDelete().
			Model(&models.Job{}).
			Where("type = ? AND json_extract(payload, '$.post_id') = ?", jobTypePublishPost, post.ID).
			Exec(txCtx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, &mcpError{Code: -32603, Message: "failed to cancel post"}
	}

	postStatus, rpcErr := h.loadMCPPostStatus(ctx, post.ID)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return mcpPostToolResult("Post canceled and returned to drafts: "+post.ID, postStatus), nil
}

type mcpSlotSuggestion struct {
	WorkspaceID string                   `json:"workspace_id"`
	SetID       string                   `json:"set_id,omitempty"`
	Timezone    string                   `json:"timezone"`
	SlotTime    string                   `json:"slot_time,omitempty"`
	SlotTimeUTC string                   `json:"slot_time_utc,omitempty"`
	Slot        *PostingScheduleResponse `json:"slot,omitempty"`
	Message     string                   `json:"message"`
}

func (h *MCPHandler) suggestNextSlot(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
		SetID       string `json:"set_id"`
		After       string `json:"after"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid suggest_next_slot arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}

	var workspace models.Workspace
	err := h.db.NewSelect().
		Model(&workspace).
		Where("id = ?", input.WorkspaceID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &mcpError{Code: -32602, Message: "workspace not found"}
		}
		return nil, &mcpError{Code: -32603, Message: "failed to load workspace"}
	}

	loc, err := time.LoadLocation(workspace.Timezone)
	if err != nil {
		loc = time.UTC
		workspace.Timezone = "UTC"
	}
	now := time.Now().In(loc)
	if strings.TrimSpace(input.After) != "" {
		after, err := time.Parse(time.RFC3339, input.After)
		if err != nil {
			return nil, &mcpError{Code: -32602, Message: "after must be an RFC3339 timestamp"}
		}
		now = after.In(loc)
	}

	var schedules []models.PostingSchedule
	query := h.db.NewSelect().
		Model(&schedules).
		Where("workspace_id = ?", input.WorkspaceID).
		Where("is_active = ?", true)
	if strings.TrimSpace(input.SetID) != "" {
		query = query.Where("set_id = ?", input.SetID)
	}
	if err := query.Scan(ctx); err != nil && err != sql.ErrNoRows {
		return nil, &mcpError{Code: -32603, Message: "failed to load posting schedules"}
	}

	if len(schedules) == 0 {
		suggestion := mcpSlotSuggestion{
			WorkspaceID: input.WorkspaceID,
			SetID:       input.SetID,
			Timezone:    workspace.Timezone,
			Message:     "No posting schedules configured for this workspace.",
		}
		return mcpSlotToolResult(suggestion), nil
	}

	var scheduledPosts []models.Post
	postQuery := h.db.NewSelect().
		Model(&scheduledPosts).
		Where("workspace_id = ?", input.WorkspaceID).
		Where("status = ?", statusScheduled).
		Where("scheduled_at >= ?", now.UTC().Add(-24*time.Hour)).
		Order("scheduled_at ASC")
	if err := postQuery.Scan(ctx); err != nil && err != sql.ErrNoRows {
		return nil, &mcpError{Code: -32603, Message: "failed to load scheduled posts"}
	}

	nextSlot, nextSlotTime := findNextConfiguredScheduleSlotTime(now, loc, schedules, scheduledPosts)
	suggestion := mcpSlotSuggestion{
		WorkspaceID: input.WorkspaceID,
		SetID:       input.SetID,
		Timezone:    workspace.Timezone,
		Message:     "No available slots found in the next month.",
	}
	if !nextSlotTime.IsZero() {
		suggestion.SlotTime = nextSlotTime.Format(time.RFC3339)
		suggestion.SlotTimeUTC = nextSlotTime.UTC().Format(time.RFC3339)
		suggestion.Message = "Next available slot found."
		if nextSlot != nil {
			slot := postingScheduleResponseForWorkspace(nextSlotTime, loc, *nextSlot)
			suggestion.Slot = &slot
		}
	}
	return mcpSlotToolResult(suggestion), nil
}

type mcpMedia struct {
	ID        string `json:"id"`
	MimeType  string `json:"mime_type"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	Deduped   bool   `json:"deduped"`
	Filename  string `json:"filename"`
	AltText   string `json:"alt_text,omitempty"`
	SourceURL string `json:"source_url"`
}

func (h *MCPHandler) uploadMediaFromURL(ctx context.Context, userID string, args map[string]any) (any, *mcpError) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
		URL         string `json:"url"`
		Filename    string `json:"filename"`
		AltText     string `json:"alt_text"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid upload_media_from_url arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}
	if h.mediaStorage == nil {
		return nil, &mcpError{Code: -32603, Message: "media storage is not configured"}
	}

	remote, filename, declaredMimeType, content, rpcErr := h.fetchRemoteMedia(ctx, input.URL, input.Filename)
	if rpcErr != nil {
		return nil, rpcErr
	}
	mediaHandler := &MediaHandler{
		db:      h.db,
		storage: h.mediaStorage,
		quota:   h.entitlement,
		usage:   h.usage,
	}
	result, err := mediaHandler.processUploadBytes(ctx, mediaUploadBytesInput{
		WorkspaceID:      input.WorkspaceID,
		Filename:         filename,
		DeclaredMimeType: declaredMimeType,
		Size:             int64(len(content)),
		Content:          content,
		AltText:          input.AltText,
	})
	if err != nil {
		return nil, &mcpError{Code: -32602, Message: err.Error()}
	}

	media := mcpMedia{
		ID:        stringFromMap(result, "id"),
		MimeType:  stringFromMap(result, "mime_type"),
		URL:       stringFromMap(result, "url"),
		Size:      int64FromMap(result, "size"),
		Deduped:   boolFromMap(result, "deduped"),
		Filename:  filename,
		AltText:   input.AltText,
		SourceURL: remote.String(),
	}
	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: "Media uploaded: " + media.ID,
		}},
		"structuredContent": map[string]any{
			"media": media,
		},
	}, nil
}

func (h *MCPHandler) fetchRemoteMedia(ctx context.Context, rawURL, requestedFilename string) (*url.URL, string, string, []byte, *mcpError) {
	remote, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || remote == nil || remote.Host == "" {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: "url must be an absolute http(s) URL"}
	}
	if rpcErr := h.validateMediaURL(ctx, remote); rpcErr != nil {
		return nil, "", "", nil, rpcErr
	}

	client := h.mediaURLHTTP
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, _ []*http.Request) error {
				if err := h.defaultValidateMediaURL(req.Context(), req.URL); err != nil {
					return err
				}
				return nil
			},
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remote.String(), nil)
	if err != nil {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: "invalid url"}
	}
	req.Header.Set("User-Agent", "openpost-mcp-media/0.1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: "failed to fetch media url"}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: fmt.Sprintf("media url returned HTTP %d", resp.StatusCode)}
	}
	finalURL := resp.Request.URL
	if rpcErr := h.validateMediaURL(ctx, finalURL); rpcErr != nil {
		return nil, "", "", nil, rpcErr
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteMediaBytes+1))
	if err != nil {
		return nil, "", "", nil, &mcpError{Code: -32603, Message: "failed to read remote media"}
	}
	if len(content) == 0 {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: "remote media is empty"}
	}
	if len(content) > maxRemoteMediaBytes {
		return nil, "", "", nil, &mcpError{Code: -32602, Message: "file size exceeds 50MB limit"}
	}

	filename := cleanRemoteMediaFilename(requestedFilename)
	if filename == "" {
		filename = cleanRemoteMediaFilename(path.Base(finalURL.Path))
	}
	if filename == "" || filename == "." || filename == "/" {
		filename = "remote-media"
	}
	return finalURL, filename, resp.Header.Get("Content-Type"), content, nil
}

func (h *MCPHandler) validateMediaURL(ctx context.Context, remote *url.URL) *mcpError {
	validator := h.mediaURLValidator
	if validator == nil {
		validator = h.defaultValidateMediaURL
	}
	if err := validator(ctx, remote); err != nil {
		return &mcpError{Code: -32602, Message: err.Error()}
	}
	return nil
}

func (h *MCPHandler) defaultValidateMediaURL(ctx context.Context, remote *url.URL) error {
	if remote == nil || remote.Hostname() == "" {
		return fmt.Errorf("url must be absolute")
	}
	if remote.Scheme != "http" && remote.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, remote.Hostname())
	if err != nil {
		return fmt.Errorf("failed to resolve url host")
	}
	if len(ips) == 0 {
		return fmt.Errorf("url host did not resolve")
	}
	for _, addr := range ips {
		ip := addr.IP
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
			return fmt.Errorf("url host resolves to a private or local address")
		}
	}
	return nil
}

func (h *MCPHandler) accessibleMCPPost(ctx context.Context, userID string, args map[string]any) (*models.Post, *mcpError) {
	var input struct {
		WorkspaceID string `json:"workspace_id"`
		PostID      string `json:"post_id"`
	}
	if err := decodeMCPArguments(args, &input); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid post arguments"}
	}
	if rpcErr := h.ensureWorkspaceAccess(ctx, userID, input.WorkspaceID); rpcErr != nil {
		return nil, rpcErr
	}
	if strings.TrimSpace(input.PostID) == "" {
		return nil, &mcpError{Code: -32602, Message: "post_id is required"}
	}
	post := new(models.Post)
	err := h.db.NewSelect().
		Model(post).
		Where("id = ? AND workspace_id = ?", input.PostID, input.WorkspaceID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &mcpError{Code: -32602, Message: "post not found in workspace"}
		}
		return nil, &mcpError{Code: -32603, Message: "failed to load post"}
	}
	return post, nil
}

func (h *MCPHandler) loadMCPPostStatus(ctx context.Context, postID string) (mcpPostStatus, *mcpError) {
	var post models.Post
	if err := h.db.NewSelect().Model(&post).Where("id = ?", postID).Scan(ctx); err != nil {
		return mcpPostStatus{}, &mcpError{Code: -32603, Message: "failed to load post"}
	}

	var rows []struct {
		models.PostDestination `bun:",extend"`
		Platform               string `bun:"platform"`
		Slug                   string `bun:"slug"`
	}
	err := h.db.NewSelect().
		Model(&rows).
		ColumnExpr("post_destination.*").
		ColumnExpr("sa.platform").
		ColumnExpr("sa.slug").
		Join("JOIN social_accounts AS sa ON sa.id = post_destination.social_account_id").
		Where("post_destination.post_id = ?", post.ID).
		OrderExpr("sa.platform ASC, sa.slug ASC").
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return mcpPostStatus{}, &mcpError{Code: -32603, Message: "failed to load post destinations"}
	}

	destinations := make([]mcpPostDestination, 0, len(rows))
	for _, row := range rows {
		destinations = append(destinations, mcpPostDestination{
			SocialAccountID: row.SocialAccountID,
			Platform:        row.Platform,
			Slug:            row.Slug,
			Status:          row.Status,
			ErrorMessage:    row.ErrorMessage,
		})
	}

	status := mcpPostStatus{
		ID:                 post.ID,
		WorkspaceID:        post.WorkspaceID,
		Content:            post.Content,
		Status:             post.Status,
		RandomDelayMinutes: post.RandomDelayMinutes,
		CreatedAt:          post.CreatedAt.Format(time.RFC3339),
		Destinations:       destinations,
	}
	if !post.ScheduledAt.IsZero() {
		status.ScheduledAt = post.ScheduledAt.Format(time.RFC3339)
	}
	if !post.ActualRunAt.IsZero() {
		status.ActualRunAt = post.ActualRunAt.Format(time.RFC3339)
	}
	return status, nil
}

func mcpPostToolResult(text string, post mcpPostStatus) map[string]any {
	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: text,
		}},
		"structuredContent": map[string]any{
			"post": post,
		},
	}
}

func mcpSlotToolResult(suggestion mcpSlotSuggestion) map[string]any {
	return map[string]any{
		"content": []mcpContent{{
			Type: "text",
			Text: suggestion.Message,
		}},
		"structuredContent": map[string]any{
			"suggestion": suggestion,
		},
	}
}

func (h *MCPHandler) checkScheduledPostQuota(ctx context.Context, workspaceID string, amount int64, scheduledAt time.Time) *mcpError {
	current, err := h.usage.CurrentMonthly(ctx, workspaceID, entitlements.LimitScheduledPostsMonthly, scheduledAt)
	if err != nil {
		return &mcpError{Code: -32603, Message: "failed to load scheduled post usage"}
	}
	decision, err := h.entitlement.Check(ctx, entitlements.Request{
		WorkspaceID: workspaceID,
		Limit:       entitlements.LimitScheduledPostsMonthly,
		Current:     current,
		Amount:      amount,
	})
	if err != nil {
		return &mcpError{Code: -32603, Message: "failed to check scheduled post limit"}
	}
	if !decision.Allowed {
		reason := decision.Reason
		if reason == "" {
			reason = "scheduled post limit exceeded"
		}
		return &mcpError{Code: -32000, Message: reason}
	}
	return nil
}

func (h *MCPHandler) recordScheduledPostUsage(ctx context.Context, workspaceID string, amount int64, scheduledAt time.Time) *mcpError {
	if _, err := h.usage.IncrementMonthly(ctx, workspaceID, entitlements.LimitScheduledPostsMonthly, amount, scheduledAt); err != nil {
		return &mcpError{Code: -32603, Message: "failed to record scheduled post usage"}
	}
	return nil
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

func normalizeMCPIDs(ids []string, field string) ([]string, *mcpError) {
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, &mcpError{Code: -32602, Message: field + " cannot contain empty values"}
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique, nil
}

func cleanRemoteMediaFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.Trim(filename, `/\`)
	if filename == "" || filename == "." {
		return ""
	}
	filename = path.Base(filename)
	filename = strings.ReplaceAll(filename, "\x00", "")
	return filename
}

func stringFromMap(values map[string]interface{}, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func boolFromMap(values map[string]interface{}, key string) bool {
	if value, ok := values[key].(bool); ok {
		return value
	}
	return false
}

func int64FromMap(values map[string]interface{}, key string) int64 {
	switch value := values[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func newUUID() string {
	return uuid.New().String()
}
