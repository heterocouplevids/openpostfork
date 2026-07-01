package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/queue"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/uptrace/bun"
)

type WorkspaceHandler struct {
	db          *bun.DB
	auth        middleware.Authenticator
	entitlement entitlements.Service
	frontendURL string
}

func NewWorkspaceHandler(db *bun.DB, authenticator middleware.Authenticator, entitlement ...entitlements.Service) *WorkspaceHandler {
	entitlementService := entitlements.Service(entitlements.NewSelfHostedService())
	if len(entitlement) > 0 && entitlement[0] != nil {
		entitlementService = entitlement[0]
	}
	return &WorkspaceHandler{db: db, auth: authenticator, entitlement: entitlementService}
}

func (h *WorkspaceHandler) SetFrontendURL(frontendURL string) {
	h.frontendURL = strings.TrimRight(strings.TrimSpace(frontendURL), "/")
}

type CreateWorkspaceInput struct {
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"100" doc:"Workspace name"`
	}
}

type CreateWorkspaceOutput struct {
	Body struct {
		WorkspaceID        string `json:"id"`
		WorkspaceName      string `json:"name"`
		WorkspaceCreatedAt string `json:"created_at"`
	}
}

type ListWorkspacesOutput struct {
	Body []struct {
		WorkspaceID        string `json:"id"`
		WorkspaceName      string `json:"name"`
		WorkspaceCreatedAt string `json:"created_at"`
	}
}

type WorkspaceMemberResponse struct {
	UserID string `json:"user_id" doc:"User ID"`
	Email  string `json:"email" doc:"User email"`
	Role   string `json:"role" doc:"Workspace role"`
}

type WorkspaceInvitationResponse struct {
	ID               string  `json:"id" doc:"Invitation ID"`
	WorkspaceID      string  `json:"workspace_id" doc:"Workspace ID"`
	Email            string  `json:"email" doc:"Invited email"`
	Role             string  `json:"role" doc:"Workspace role to grant"`
	InvitedByUserID  string  `json:"invited_by_user_id" doc:"Inviting user ID"`
	AcceptedByUserID *string `json:"accepted_by_user_id,omitempty" doc:"Accepting user ID"`
	Token            string  `json:"token,omitempty" doc:"Raw invite token returned once on creation"`
	AcceptURL        string  `json:"accept_url,omitempty" doc:"Browser URL that accepts the invitation"`
	ExpiresAt        string  `json:"expires_at" doc:"Invitation expiry time"`
	AcceptedAt       *string `json:"accepted_at,omitempty" doc:"When the invitation was accepted"`
	RevokedAt        *string `json:"revoked_at,omitempty" doc:"When the invitation was revoked"`
	CreatedAt        string  `json:"created_at" doc:"Invitation creation time"`
}

type WorkspaceTeamOutput struct {
	Body struct {
		Members      []WorkspaceMemberResponse     `json:"members"`
		Invitations  []WorkspaceInvitationResponse `json:"invitations"`
		CurrentSeats int64                         `json:"current_seats"`
	}
}

type WorkspaceTeamInput struct {
	PathID string `path:"id" doc:"Workspace ID"`
}

type CreateWorkspaceInvitationInput struct {
	PathID string `path:"id" doc:"Workspace ID"`
	Body   struct {
		Email string `json:"email" format:"email" doc:"Email address to invite"`
		Role  string `json:"role" enum:"admin,editor,viewer" doc:"Workspace role to grant"`
	}
}

type CreateWorkspaceInvitationOutput struct {
	Body WorkspaceInvitationResponse
}

type RevokeWorkspaceInvitationInput struct {
	PathID       string `path:"id" doc:"Workspace ID"`
	InvitationID string `path:"invitation_id" doc:"Invitation ID"`
}

type RevokeWorkspaceInvitationOutput struct {
	Body struct {
		Revoked bool `json:"revoked"`
	}
}

type AcceptWorkspaceInvitationInput struct {
	Body struct {
		Token string `json:"token" minLength:"16" doc:"Raw invitation token"`
	}
}

type AcceptWorkspaceInvitationOutput struct {
	Body struct {
		WorkspaceID string `json:"workspace_id"`
		Role        string `json:"role"`
		Accepted    bool   `json:"accepted"`
	}
}

func (h *WorkspaceHandler) CreateWorkspace(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-workspace",
		Method:        http.MethodPost,
		Path:          "/workspaces",
		Summary:       "Create a new workspace",
		Tags:          []string{tagWorkspaces},
		DefaultStatus: http.StatusOK,
		Middlewares:   huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, input *CreateWorkspaceInput) (*CreateWorkspaceOutput, error) {
		userID := middleware.GetUserID(ctx)
		if middleware.GetWorkspaceID(ctx) != "" {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}
		if err := h.checkCreateWorkspaceEntitlement(ctx, userID); err != nil {
			return nil, err
		}

		workspace := &models.Workspace{
			ID:        uuid.New().String(),
			Name:      input.Body.Name,
			CreatedAt: time.Now().UTC(),
		}

		member := &models.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      userID,
			Role:        models.WorkspaceRoleAdmin,
		}

		err := h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			if _, err := tx.NewInsert().Model(workspace).Exec(txCtx); err != nil {
				return err
			}
			if _, err := tx.NewInsert().Model(member).Exec(txCtx); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to create workspace")
		}

		resp := &CreateWorkspaceOutput{}
		resp.Body.WorkspaceID = workspace.ID
		resp.Body.WorkspaceName = workspace.Name
		resp.Body.WorkspaceCreatedAt = workspace.CreatedAt.Format(time.RFC3339)
		return resp, nil
	})
}

func (h *WorkspaceHandler) ListWorkspaceTeam(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-workspace-team",
		Method:      http.MethodGet,
		Path:        "/workspaces/{id}/team",
		Summary:     "List workspace members and pending invitations",
		Tags:        []string{tagWorkspaces},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *WorkspaceTeamInput) (*WorkspaceTeamOutput, error) {
		if _, err := h.requireWorkspaceMember(ctx, input.PathID, middleware.GetUserID(ctx)); err != nil {
			return nil, err
		}

		members, err := h.listWorkspaceMembers(ctx, input.PathID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch workspace members")
		}
		invitations, err := h.listPendingWorkspaceInvitations(ctx, input.PathID, time.Now().UTC())
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch workspace invitations")
		}

		resp := &WorkspaceTeamOutput{}
		resp.Body.Members = members
		resp.Body.Invitations = workspaceInvitationResponses(invitations, "", "")
		resp.Body.CurrentSeats = int64(len(members) + len(invitations))
		return resp, nil
	})
}

func (h *WorkspaceHandler) CreateWorkspaceInvitation(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-workspace-invitation",
		Method:        http.MethodPost,
		Path:          "/workspaces/{id}/invitations",
		Summary:       "Create a workspace invitation",
		Tags:          []string{tagWorkspaces},
		DefaultStatus: http.StatusOK,
		Middlewares:   huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:        []int{400, 402, 403, 404, 409},
	}, func(ctx context.Context, input *CreateWorkspaceInvitationInput) (*CreateWorkspaceInvitationOutput, error) {
		userID := middleware.GetUserID(ctx)
		if err := h.requireWorkspaceAdmin(ctx, input.PathID, userID); err != nil {
			return nil, err
		}

		email := normalizeWorkspaceInvitationEmail(input.Body.Email)
		if email == "" {
			return nil, huma.Error400BadRequest("email is required")
		}
		role := strings.TrimSpace(input.Body.Role)
		if role == "" {
			role = models.WorkspaceRoleEditor
		}
		if !isWorkspaceRole(role) {
			return nil, huma.Error400BadRequest("invalid workspace role")
		}

		now := time.Now().UTC()
		if err := h.ensureCanInviteWorkspaceSeat(ctx, input.PathID, email, now); err != nil {
			return nil, err
		}

		token, tokenHash, err := generateWorkspaceInvitationToken()
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to generate invitation token")
		}
		invitation := &models.WorkspaceInvitation{
			ID:              uuid.New().String(),
			WorkspaceID:     input.PathID,
			Email:           email,
			Role:            role,
			InvitedByUserID: userID,
			TokenHash:       tokenHash,
			ExpiresAt:       now.Add(7 * 24 * time.Hour),
			CreatedAt:       now,
		}
		if _, err := h.db.NewInsert().Model(invitation).Exec(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to create workspace invitation")
		}

		resp := &CreateWorkspaceInvitationOutput{}
		resp.Body = workspaceInvitationResponse(*invitation, token, h.acceptWorkspaceInvitationURL(token))
		return resp, nil
	})
}

func (h *WorkspaceHandler) RevokeWorkspaceInvitation(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "revoke-workspace-invitation",
		Method:      http.MethodDelete,
		Path:        "/workspaces/{id}/invitations/{invitation_id}",
		Summary:     "Revoke a pending workspace invitation",
		Tags:        []string{tagWorkspaces},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *RevokeWorkspaceInvitationInput) (*RevokeWorkspaceInvitationOutput, error) {
		if err := h.requireWorkspaceAdmin(ctx, input.PathID, middleware.GetUserID(ctx)); err != nil {
			return nil, err
		}

		res, err := h.db.NewUpdate().
			Model((*models.WorkspaceInvitation)(nil)).
			Set("revoked_at = ?", time.Now().UTC()).
			Where("id = ? AND workspace_id = ? AND accepted_at IS NULL AND revoked_at IS NULL", input.InvitationID, input.PathID).
			Exec(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to revoke workspace invitation")
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil, huma.Error404NotFound("workspace invitation not found")
		}

		return &RevokeWorkspaceInvitationOutput{Body: struct {
			Revoked bool `json:"revoked"`
		}{Revoked: true}}, nil
	})
}

func (h *WorkspaceHandler) AcceptWorkspaceInvitation(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "accept-workspace-invitation",
		Method:        http.MethodPost,
		Path:          "/workspace-invitations/accept",
		Summary:       "Accept a workspace invitation",
		Tags:          []string{tagWorkspaces},
		DefaultStatus: http.StatusOK,
		Middlewares:   huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:        []int{400, 403, 404, 409},
	}, func(ctx context.Context, input *AcceptWorkspaceInvitationInput) (*AcceptWorkspaceInvitationOutput, error) {
		userID := middleware.GetUserID(ctx)
		userEmail := normalizeWorkspaceInvitationEmail(middleware.GetUserEmail(ctx))
		tokenHash := hashWorkspaceInvitationToken(input.Body.Token)
		now := time.Now().UTC()

		var invitation models.WorkspaceInvitation
		err := h.db.NewSelect().
			Model(&invitation).
			Where("token_hash = ?", tokenHash).
			Scan(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, huma.Error404NotFound("workspace invitation not found")
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch workspace invitation")
		}
		if !middleware.WorkspaceScopeAllows(ctx, invitation.WorkspaceID) {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}
		if !invitation.AcceptedAt.IsZero() {
			return nil, huma.NewError(http.StatusConflict, "workspace invitation already accepted")
		}
		if !invitation.RevokedAt.IsZero() {
			return nil, huma.NewError(http.StatusConflict, "workspace invitation was revoked")
		}
		if !invitation.ExpiresAt.After(now) {
			return nil, huma.NewError(http.StatusConflict, "workspace invitation expired")
		}
		if invitation.Email != userEmail {
			return nil, huma.Error403Forbidden("workspace invitation belongs to a different email address")
		}

		err = h.db.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
			member := &models.WorkspaceMember{
				WorkspaceID: invitation.WorkspaceID,
				UserID:      userID,
				Role:        invitation.Role,
			}
			if _, err := tx.NewInsert().
				Model(member).
				On("CONFLICT (workspace_id, user_id) DO NOTHING").
				Exec(txCtx); err != nil {
				return err
			}
			res, err := tx.NewUpdate().
				Model((*models.WorkspaceInvitation)(nil)).
				Set("accepted_by_user_id = ?", userID).
				Set("accepted_at = ?", now).
				Where("id = ? AND accepted_at IS NULL AND revoked_at IS NULL", invitation.ID).
				Exec(txCtx)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return sql.ErrNoRows
			}
			return nil
		})
		if errors.Is(err, sql.ErrNoRows) {
			return nil, huma.NewError(http.StatusConflict, "workspace invitation is no longer pending")
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to accept workspace invitation")
		}

		return &AcceptWorkspaceInvitationOutput{Body: struct {
			WorkspaceID string `json:"workspace_id"`
			Role        string `json:"role"`
			Accepted    bool   `json:"accepted"`
		}{
			WorkspaceID: invitation.WorkspaceID,
			Role:        invitation.Role,
			Accepted:    true,
		}}, nil
	})
}

func (h *WorkspaceHandler) checkCreateWorkspaceEntitlement(ctx context.Context, userID string) error {
	var current int
	if err := h.db.NewSelect().
		ColumnExpr("COUNT(*)").
		Model((*models.WorkspaceMember)(nil)).
		Where("user_id = ?", userID).
		Scan(ctx, &current); err != nil {
		return huma.Error500InternalServerError("failed to check workspace limit")
	}

	decision, err := h.entitlement.Check(ctx, entitlements.Request{
		UserID:  userID,
		Limit:   entitlements.LimitWorkspaces,
		Current: int64(current),
		Amount:  1,
	})
	if err != nil {
		return huma.Error500InternalServerError("failed to check workspace limit")
	}
	if !decision.Allowed {
		reason := decision.Reason
		if reason == "" {
			reason = "workspace limit exceeded"
		}
		return huma.NewError(http.StatusPaymentRequired, reason)
	}
	return nil
}

func (h *WorkspaceHandler) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) (*models.WorkspaceMember, error) {
	if !middleware.WorkspaceScopeAllows(ctx, workspaceID) {
		return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	var member models.WorkspaceMember
	err := h.db.NewSelect().
		Model(&member).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
	}
	if err != nil {
		return nil, huma.Error500InternalServerError(errValidateWorkspaceAccess)
	}
	return &member, nil
}

func (h *WorkspaceHandler) requireWorkspaceAdmin(ctx context.Context, workspaceID, userID string) error {
	member, err := h.requireWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if member.Role != models.WorkspaceRoleAdmin {
		return huma.Error403Forbidden("workspace admin role required")
	}
	return nil
}

func (h *WorkspaceHandler) listWorkspaceMembers(ctx context.Context, workspaceID string) ([]WorkspaceMemberResponse, error) {
	var rows []WorkspaceMemberResponse
	err := h.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		ModelTableExpr("workspace_members AS wm").
		ColumnExpr("wm.user_id, u.email, wm.role").
		Join("JOIN users AS u ON u.id = wm.user_id").
		Where("wm.workspace_id = ?", workspaceID).
		Order("u.email ASC").
		Scan(ctx, &rows)
	return rows, err
}

func (h *WorkspaceHandler) listPendingWorkspaceInvitations(ctx context.Context, workspaceID string, now time.Time) ([]models.WorkspaceInvitation, error) {
	var invitations []models.WorkspaceInvitation
	err := h.db.NewSelect().
		Model(&invitations).
		Where("workspace_id = ? AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > ?", workspaceID, now).
		Order("created_at DESC").
		Scan(ctx)
	return invitations, err
}

func (h *WorkspaceHandler) ensureCanInviteWorkspaceSeat(ctx context.Context, workspaceID, email string, now time.Time) error {
	var existingMemberCount int
	if err := h.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("workspace_members AS wm").
		Join("JOIN users AS u ON u.id = wm.user_id").
		Where("wm.workspace_id = ? AND u.email = ?", workspaceID, email).
		Scan(ctx, &existingMemberCount); err != nil {
		return huma.Error500InternalServerError("failed to check workspace members")
	}
	if existingMemberCount > 0 {
		return huma.NewError(http.StatusConflict, "user is already a workspace member")
	}

	var pendingForEmail int
	if err := h.db.NewSelect().
		ColumnExpr("COUNT(*)").
		Model((*models.WorkspaceInvitation)(nil)).
		Where("workspace_id = ? AND email = ? AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > ?", workspaceID, email, now).
		Scan(ctx, &pendingForEmail); err != nil {
		return huma.Error500InternalServerError("failed to check workspace invitations")
	}
	if pendingForEmail > 0 {
		return huma.NewError(http.StatusConflict, "workspace invitation already pending")
	}

	var memberCount int
	if err := h.db.NewSelect().
		ColumnExpr("COUNT(*)").
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ?", workspaceID).
		Scan(ctx, &memberCount); err != nil {
		return huma.Error500InternalServerError("failed to check team member limit")
	}
	var pendingCount int
	if err := h.db.NewSelect().
		ColumnExpr("COUNT(*)").
		Model((*models.WorkspaceInvitation)(nil)).
		Where("workspace_id = ? AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > ?", workspaceID, now).
		Scan(ctx, &pendingCount); err != nil {
		return huma.Error500InternalServerError("failed to check team member limit")
	}

	decision, err := h.entitlement.Check(ctx, entitlements.Request{
		WorkspaceID: workspaceID,
		Limit:       entitlements.LimitTeamMembers,
		Current:     int64(memberCount + pendingCount),
		Amount:      1,
	})
	if err != nil {
		return huma.Error500InternalServerError("failed to check team member limit")
	}
	if !decision.Allowed {
		reason := decision.Reason
		if reason == "" {
			reason = "team member limit exceeded"
		}
		return huma.NewError(http.StatusPaymentRequired, reason)
	}
	return nil
}

func (h *WorkspaceHandler) ListWorkspaces(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-workspaces",
		Method:      http.MethodGet,
		Path:        "/workspaces",
		Summary:     "List workspaces for the current user",
		Tags:        []string{tagWorkspaces},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
	}, func(ctx context.Context, _ *struct{}) (*ListWorkspacesOutput, error) {
		userID := middleware.GetUserID(ctx)

		var workspaces []models.Workspace
		query := h.db.NewSelect().
			Model(&workspaces).
			Join("JOIN workspace_members AS wm ON wm.workspace_id = workspace.id").
			Where("wm.user_id = ?", userID)
		if workspaceID := middleware.GetWorkspaceID(ctx); workspaceID != "" {
			query = query.Where("workspace.id = ?", workspaceID)
		}
		err := query.Scan(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to fetch workspaces")
		}

		resp := &ListWorkspacesOutput{Body: []struct {
			WorkspaceID        string `json:"id"`
			WorkspaceName      string `json:"name"`
			WorkspaceCreatedAt string `json:"created_at"`
		}{}}
		for _, ws := range workspaces {
			resp.Body = append(resp.Body, struct {
				WorkspaceID        string `json:"id"`
				WorkspaceName      string `json:"name"`
				WorkspaceCreatedAt string `json:"created_at"`
			}{
				WorkspaceID:        ws.ID,
				WorkspaceName:      ws.Name,
				WorkspaceCreatedAt: ws.CreatedAt.Format(time.RFC3339),
			})
		}
		return resp, nil
	})
}

func workspaceInvitationResponses(invitations []models.WorkspaceInvitation, rawToken, acceptURL string) []WorkspaceInvitationResponse {
	out := make([]WorkspaceInvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		out = append(out, workspaceInvitationResponse(invitation, rawToken, acceptURL))
	}
	return out
}

func workspaceInvitationResponse(invitation models.WorkspaceInvitation, rawToken, acceptURL string) WorkspaceInvitationResponse {
	return WorkspaceInvitationResponse{
		ID:               invitation.ID,
		WorkspaceID:      invitation.WorkspaceID,
		Email:            invitation.Email,
		Role:             invitation.Role,
		InvitedByUserID:  invitation.InvitedByUserID,
		AcceptedByUserID: optionalString(invitation.AcceptedByUserID),
		Token:            rawToken,
		AcceptURL:        acceptURL,
		ExpiresAt:        invitation.ExpiresAt.UTC().Format(time.RFC3339),
		AcceptedAt:       optionalTime(invitation.AcceptedAt),
		RevokedAt:        optionalTime(invitation.RevokedAt),
		CreatedAt:        invitation.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func normalizeWorkspaceInvitationEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isWorkspaceRole(role string) bool {
	switch role {
	case models.WorkspaceRoleAdmin, models.WorkspaceRoleEditor, models.WorkspaceRoleViewer:
		return true
	default:
		return false
	}
}

const workspaceInvitationTokenPrefix = "op_inv"

func generateWorkspaceInvitationToken() (string, string, error) {
	secret, err := randomWorkspaceInvitationSecret()
	if err != nil {
		return "", "", err
	}
	token := workspaceInvitationTokenPrefix + "_" + secret
	return token, hashWorkspaceInvitationToken(token), nil
}

func randomWorkspaceInvitationSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashWorkspaceInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (h *WorkspaceHandler) acceptWorkspaceInvitationURL(token string) string {
	if h.frontendURL == "" {
		return "/invite?token=" + token
	}
	return h.frontendURL + "/invite?token=" + token
}

type GetWorkspaceSettingsInput struct {
	PathID string `path:"id" doc:"Workspace ID"`
}

type GetWorkspaceSettingsOutput struct {
	Body struct {
		Timezone            string `json:"timezone"`
		WeekStart           int    `json:"week_start"`
		MediaCleanupDays    int    `json:"media_cleanup_days"`
		RandomDelayMinutes  int    `json:"random_delay_minutes"`
		DraftGapMinutes     int    `json:"draft_gap_minutes"`
		SlotStartHour       int    `json:"slot_start_hour"`
		SlotEndHour         int    `json:"slot_end_hour"`
		SlotIntervalMinutes int    `json:"slot_interval_minutes"`
	}
}

type UpdateWorkspaceSettingsInput struct {
	PathID string `path:"id" doc:"Workspace ID"`
	Body   struct {
		Timezone            *string `json:"timezone,omitempty"`
		WeekStart           *int    `json:"week_start,omitempty"`
		MediaCleanupDays    *int    `json:"media_cleanup_days,omitempty"`
		RandomDelayMinutes  *int    `json:"random_delay_minutes,omitempty"`
		DraftGapMinutes     *int    `json:"draft_gap_minutes,omitempty"`
		SlotStartHour       *int    `json:"slot_start_hour,omitempty"`
		SlotEndHour         *int    `json:"slot_end_hour,omitempty"`
		SlotIntervalMinutes *int    `json:"slot_interval_minutes,omitempty"`
	}
}

type UpdateWorkspaceSettingsOutput struct {
	Body struct {
		Timezone            string `json:"timezone"`
		WeekStart           int    `json:"week_start"`
		MediaCleanupDays    int    `json:"media_cleanup_days"`
		RandomDelayMinutes  int    `json:"random_delay_minutes"`
		DraftGapMinutes     int    `json:"draft_gap_minutes"`
		SlotStartHour       int    `json:"slot_start_hour"`
		SlotEndHour         int    `json:"slot_end_hour"`
		SlotIntervalMinutes int    `json:"slot_interval_minutes"`
	}
}

func (h *WorkspaceHandler) GetWorkspaceSettings(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-workspace-settings",
		Method:      http.MethodGet,
		Path:        "/workspaces/{id}/settings",
		Summary:     "Get workspace settings",
		Tags:        []string{tagWorkspaces},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{403, 404},
	}, func(ctx context.Context, input *GetWorkspaceSettingsInput) (*GetWorkspaceSettingsOutput, error) {
		userID := middleware.GetUserID(ctx)
		if !middleware.WorkspaceScopeAllows(ctx, input.PathID) {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		var memberCount int
		memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", input.PathID, userID).
			Count(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError(errValidateWorkspaceAccess)
		}
		if memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		var workspace models.Workspace
		err = h.db.NewSelect().Model(&workspace).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("workspace not found")
			}
			return nil, huma.Error500InternalServerError("failed to fetch workspace")
		}

		return &GetWorkspaceSettingsOutput{Body: struct {
			Timezone            string `json:"timezone"`
			WeekStart           int    `json:"week_start"`
			MediaCleanupDays    int    `json:"media_cleanup_days"`
			RandomDelayMinutes  int    `json:"random_delay_minutes"`
			DraftGapMinutes     int    `json:"draft_gap_minutes"`
			SlotStartHour       int    `json:"slot_start_hour"`
			SlotEndHour         int    `json:"slot_end_hour"`
			SlotIntervalMinutes int    `json:"slot_interval_minutes"`
		}{
			Timezone:            workspace.Timezone,
			WeekStart:           workspace.WeekStart,
			MediaCleanupDays:    workspace.MediaCleanupDays,
			RandomDelayMinutes:  workspace.RandomDelayMinutes,
			DraftGapMinutes:     workspace.DraftGapMinutes,
			SlotStartHour:       workspace.SlotStartHour,
			SlotEndHour:         workspace.SlotEndHour,
			SlotIntervalMinutes: workspace.SlotIntervalMinutes,
		}}, nil
	})
}

//nolint:gocyclo
func (h *WorkspaceHandler) UpdateWorkspaceSettings(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "update-workspace-settings",
		Method:      http.MethodPatch,
		Path:        "/workspaces/{id}/settings",
		Summary:     "Update workspace settings",
		Tags:        []string{tagWorkspaces},
		Middlewares: huma.Middlewares{middleware.AuthMiddleware(api, h.auth)},
		Errors:      []int{400, 403, 404},
	}, func(ctx context.Context, input *UpdateWorkspaceSettingsInput) (*UpdateWorkspaceSettingsOutput, error) {
		userID := middleware.GetUserID(ctx)
		if !middleware.WorkspaceScopeAllows(ctx, input.PathID) {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		var memberCount int
		memberCount, err := h.db.NewSelect().Model((*models.WorkspaceMember)(nil)).
			Where("workspace_id = ? AND user_id = ?", input.PathID, userID).
			Count(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError(errValidateWorkspaceAccess)
		}
		if memberCount == 0 {
			return nil, huma.Error403Forbidden(errWorkspaceAccessDenied)
		}

		var workspace models.Workspace
		err = h.db.NewSelect().Model(&workspace).Where("id = ?", input.PathID).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("workspace not found")
			}
			return nil, huma.Error500InternalServerError("failed to fetch workspace")
		}

		if input.Body.Timezone != nil {
			workspace.Timezone = *input.Body.Timezone
		}
		if input.Body.WeekStart != nil {
			if *input.Body.WeekStart < 0 || *input.Body.WeekStart > 1 {
				return nil, huma.Error400BadRequest("week_start must be 0 (Sunday) or 1 (Monday)")
			}
			workspace.WeekStart = *input.Body.WeekStart
		}
		if input.Body.MediaCleanupDays != nil {
			if *input.Body.MediaCleanupDays < 0 || *input.Body.MediaCleanupDays > 365 {
				return nil, huma.Error400BadRequest("media_cleanup_days must be between 0 and 365")
			}
			workspace.MediaCleanupDays = *input.Body.MediaCleanupDays
		}
		if input.Body.RandomDelayMinutes != nil {
			if *input.Body.RandomDelayMinutes < 0 || *input.Body.RandomDelayMinutes > 60 {
				return nil, huma.Error400BadRequest("random_delay_minutes must be between 0 and 60")
			}
			workspace.RandomDelayMinutes = *input.Body.RandomDelayMinutes
		}
		if input.Body.DraftGapMinutes != nil {
			if *input.Body.DraftGapMinutes < 0 || *input.Body.DraftGapMinutes > 24*60 {
				return nil, huma.Error400BadRequest("draft_gap_minutes must be between 0 and 1440")
			}
			workspace.DraftGapMinutes = *input.Body.DraftGapMinutes
		}
		if input.Body.SlotStartHour != nil {
			if *input.Body.SlotStartHour < 0 || *input.Body.SlotStartHour > 23 {
				return nil, huma.Error400BadRequest("slot_start_hour must be between 0 and 23")
			}
			workspace.SlotStartHour = *input.Body.SlotStartHour
		}
		if input.Body.SlotEndHour != nil {
			if *input.Body.SlotEndHour < 0 || *input.Body.SlotEndHour > 23 {
				return nil, huma.Error400BadRequest("slot_end_hour must be between 0 and 23")
			}
			workspace.SlotEndHour = *input.Body.SlotEndHour
		}
		if input.Body.SlotIntervalMinutes != nil {
			if *input.Body.SlotIntervalMinutes < 1 || *input.Body.SlotIntervalMinutes > 180 {
				return nil, huma.Error400BadRequest("slot_interval_minutes must be between 1 and 180")
			}
			workspace.SlotIntervalMinutes = *input.Body.SlotIntervalMinutes
		}

		_, err = h.db.NewUpdate().Model(&workspace).
			Column("timezone", "week_start", "media_cleanup_days", "random_delay_minutes", "draft_gap_minutes", "slot_start_hour", "slot_end_hour", "slot_interval_minutes").
			Where("id = ?", input.PathID).
			Exec(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to update workspace")
		}

		if input.Body.MediaCleanupDays != nil {
			_ = queue.ScheduleMediaCleanup(h.db, input.PathID, workspace.MediaCleanupDays) //nolint:errcheck
		}

		return &UpdateWorkspaceSettingsOutput{Body: struct {
			Timezone            string `json:"timezone"`
			WeekStart           int    `json:"week_start"`
			MediaCleanupDays    int    `json:"media_cleanup_days"`
			RandomDelayMinutes  int    `json:"random_delay_minutes"`
			DraftGapMinutes     int    `json:"draft_gap_minutes"`
			SlotStartHour       int    `json:"slot_start_hour"`
			SlotEndHour         int    `json:"slot_end_hour"`
			SlotIntervalMinutes int    `json:"slot_interval_minutes"`
		}{
			Timezone:            workspace.Timezone,
			WeekStart:           workspace.WeekStart,
			MediaCleanupDays:    workspace.MediaCleanupDays,
			RandomDelayMinutes:  workspace.RandomDelayMinutes,
			DraftGapMinutes:     workspace.DraftGapMinutes,
			SlotStartHour:       workspace.SlotStartHour,
			SlotEndHour:         workspace.SlotEndHour,
			SlotIntervalMinutes: workspace.SlotIntervalMinutes,
		}}, nil
	})
}
