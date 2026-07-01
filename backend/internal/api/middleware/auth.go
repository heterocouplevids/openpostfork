package middleware

import (
	"context"
	"net/http"
	"net/netip"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/apitokens"
	"github.com/openpost/backend/internal/services/auth"
	"github.com/uptrace/bun"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	EmailKey       contextKey = "email"
	WorkspaceIDKey contextKey = "workspace_id"
	ClientIPKey    contextKey = "client_ip"
	errorKey       contextKey = "error"

	// bearerPrefix is the canonical HTTP Authorization scheme this
	// middleware accepts. Centralised as a const to satisfy
	// golangci-lint's goconst rule across all three middleware
	// implementations (AuthMiddleware, JWTMiddleware, BearerMiddleware).
	bearerPrefix = "Bearer"
)

type Principal struct {
	UserID      string
	Email       string
	Scope       string
	WorkspaceID string
	Audience    string
	ClientID    string
	ClientName  string
	TokenPrefix string
}

type Authenticator interface {
	AuthenticateBearer(ctx context.Context, token string) (*Principal, error)
}

type JWTAuthenticator struct {
	service *auth.Service
}

func NewJWTAuthenticator(service *auth.Service) *JWTAuthenticator {
	return &JWTAuthenticator{service: service}
}

func (a *JWTAuthenticator) AuthenticateBearer(_ context.Context, token string) (*Principal, error) {
	claims, err := a.service.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return &Principal{UserID: claims.UserID, Email: claims.Email}, nil
}

type CompositeService struct {
	jwt       Authenticator
	apiTokens *apitokens.Service
}

func NewCompositeService(jwtService *auth.Service, apiTokenService *apitokens.Service) *CompositeService {
	return &CompositeService{
		jwt:       NewJWTAuthenticator(jwtService),
		apiTokens: apiTokenService,
	}
}

func (s *CompositeService) AuthenticateBearer(ctx context.Context, token string) (*Principal, error) {
	principal, err := s.jwt.AuthenticateBearer(ctx, token)
	if err == nil {
		return principal, nil
	}
	if s.apiTokens == nil {
		return nil, err
	}

	apiPrincipal, apiErr := s.apiTokens.ValidateToken(ctx, token)
	if apiErr != nil {
		return nil, err
	}
	return &Principal{
		UserID:      apiPrincipal.UserID,
		Email:       apiPrincipal.Email,
		Scope:       apiPrincipal.Scope,
		WorkspaceID: apiPrincipal.WorkspaceID,
		Audience:    apiPrincipal.Audience,
		ClientID:    apiPrincipal.TokenID,
		ClientName:  apiPrincipal.TokenName,
		TokenPrefix: apiPrincipal.TokenPrefix,
	}, nil
}

func AuthMiddleware(api huma.API, authenticator Authenticator) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing authorization header")
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != bearerPrefix {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		principal, err := authenticator.AuthenticateBearer(ctx.Context(), tokenParts[1])
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx = huma.WithValue(ctx, UserIDKey, principal.UserID)
		ctx = huma.WithValue(ctx, EmailKey, principal.Email)
		if principal.WorkspaceID != "" {
			ctx = huma.WithValue(ctx, WorkspaceIDKey, principal.WorkspaceID)
		}
		next(ctx)
	}
}

func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func GetWorkspaceID(ctx context.Context) string {
	if v, ok := ctx.Value(WorkspaceIDKey).(string); ok {
		return v
	}
	return ""
}

func GetClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(ClientIPKey).(string); ok {
		return v
	}
	return ""
}

func RequestMetadataMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		next(huma.WithValue(ctx, ClientIPKey, requestClientIP(ctx)))
	}
}

// WorkspaceAccessMiddleware validates that the user has access to the workspace specified in the request.
// This should be used after AuthMiddleware.
func WorkspaceAccessMiddleware(api huma.API, _ *bun.DB) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		userID := GetUserID(ctx.Context())
		if userID == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Get workspace_id from query or body - this is a simplified version
		// In practice, you'd need to extract it from the specific input structure
		// This middleware serves as a pattern that handlers can follow
		next(ctx)
	}
}

// CheckWorkspaceAccess is a helper function to verify workspace access.
func CheckWorkspaceAccess(ctx context.Context, db *bun.DB, workspaceID, userID string) (bool, error) {
	if !WorkspaceScopeAllows(ctx, workspaceID) {
		return false, nil
	}
	var memberCount int
	memberCount, err := db.NewSelect().Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return memberCount > 0, nil
}

func WorkspaceScopeAllows(ctx context.Context, workspaceID string) bool {
	scopedWorkspaceID := strings.TrimSpace(GetWorkspaceID(ctx))
	if scopedWorkspaceID == "" {
		return true
	}
	return scopedWorkspaceID == strings.TrimSpace(workspaceID)
}

func JWTMiddleware(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "missing authorization header"})
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != bearerPrefix {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "invalid authorization header format"})
			}

			claims, err := authService.ValidateToken(tokenParts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "invalid or expired token"})
			}

			c.Set(string(UserIDKey), claims.UserID)
			c.Set(string(EmailKey), claims.Email)

			return next(c)
		}
	}
}

// BearerMiddleware is the Echo-shaped counterpart of AuthMiddleware.
// It accepts a JWT session token OR an API/CLI token (op_cli_...) via
// the unified Authenticator, and exposes the resolved principal on the
// Echo context. Use it on legacy Echo routes (e.g. /api/v1/media/upload)
// that need to support CLI tokens but cannot be expressed as Huma ops.
func BearerMiddleware(authenticator Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "missing authorization header"})
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != bearerPrefix {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "invalid authorization header format"})
			}

			principal, err := authenticator.AuthenticateBearer(c.Request().Context(), tokenParts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{string(errorKey): "invalid or expired token"})
			}

			c.Set(string(UserIDKey), principal.UserID)
			c.Set(string(EmailKey), principal.Email)
			requestCtx := context.WithValue(c.Request().Context(), UserIDKey, principal.UserID)
			requestCtx = context.WithValue(requestCtx, EmailKey, principal.Email)
			if principal.WorkspaceID != "" {
				c.Set(string(WorkspaceIDKey), principal.WorkspaceID)
				requestCtx = context.WithValue(requestCtx, WorkspaceIDKey, principal.WorkspaceID)
			}
			c.SetRequest(c.Request().WithContext(requestCtx))

			return next(c)
		}
	}
}

func requestClientIP(ctx huma.Context) string {
	if forwarded := strings.TrimSpace(strings.Split(ctx.Header("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	if realIP := strings.TrimSpace(ctx.Header("X-Real-Ip")); realIP != "" {
		return realIP
	}
	if addr, err := netip.ParseAddrPort(ctx.RemoteAddr()); err == nil {
		return addr.Addr().String()
	}
	return ctx.RemoteAddr()
}
