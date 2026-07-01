package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/openpost/backend/internal/api/handlers"
	"github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/platform"
	"github.com/openpost/backend/internal/services/apitokens"
	"github.com/openpost/backend/internal/services/auth"
	"github.com/openpost/backend/internal/services/billing"
	cliauth "github.com/openpost/backend/internal/services/cli_auth"
	servicecrypto "github.com/openpost/backend/internal/services/crypto"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/mastodonapps"
	"github.com/openpost/backend/internal/services/mcpoauth"
	"github.com/openpost/backend/internal/services/mediasigner"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/openpost/backend/internal/services/mfa"
	"github.com/openpost/backend/internal/services/providerapps"
	"github.com/openpost/backend/internal/services/sessions"
	"github.com/uptrace/bun"
)

type RouteDeps struct {
	DB                           *bun.DB
	AuthService                  *auth.Service
	Authenticator                middleware.Authenticator
	SessionService               *sessions.Service
	APITokenService              *apitokens.Service
	CLIAuthService               *cliauth.Service
	MCPOAuthService              *mcpoauth.Service
	BillingService               *billing.Service
	MediaStorage                 mediastore.BlobStorage
	MediaSigner                  *mediasigner.Signer
	Entitlement                  entitlements.Service
	TokenEncryptor               *servicecrypto.TokenEncryptor
	MFAService                   *mfa.Service
	Providers                    map[string]platform.Adapter
	MastodonAppService           *mastodonapps.Service
	FrontendURL                  string
	PublicURL                    string
	DisableRegistrations         bool
	DisableLinkedInThreadReplies bool

	MediaHandler    *handlers.MediaHandler
	BillingHandler  *handlers.BillingHandler
	MCPOAuthHandler *handlers.MCPOAuthHandler
}

func RegisterHumaRoutes(api huma.API, deps RouteDeps) {
	mediaHandler := deps.MediaHandler
	if mediaHandler == nil {
		mediaHandler = handlers.NewMediaHandler(deps.DB, deps.MediaStorage, deps.AuthService, deps.Authenticator, deps.MediaSigner)
		mediaHandler.SetEntitlement(deps.Entitlement)
	}
	mediaHandler.RegisterRoutes(api)

	billingHandler := deps.BillingHandler
	if billingHandler == nil {
		billingHandler = handlers.NewBillingHandler(deps.BillingService, deps.DB, deps.Authenticator)
	}
	billingHandler.RegisterAPIRoutes(api)

	authHandler := handlers.NewAuthHandler(
		deps.DB,
		deps.AuthService,
		deps.Authenticator,
		deps.TokenEncryptor,
		deps.MFAService,
		deps.DisableRegistrations,
	)
	authHandler.SetSessionService(deps.SessionService)
	authHandler.Register(api)
	authHandler.Login(api)
	authHandler.VerifyTOTPLogin(api)
	authHandler.BeginPasskeyLogin(api)
	authHandler.FinishPasskeyLogin(api)
	authHandler.Me(api)
	authHandler.SecurityStatus(api)
	authHandler.ListSessions(api)
	authHandler.RevokeSession(api)
	authHandler.BeginTOTPSetup(api)
	authHandler.ConfirmTOTPSetup(api)
	authHandler.DisableTOTP(api)
	authHandler.BeginPasskeyRegistration(api)
	authHandler.FinishPasskeyRegistration(api)
	authHandler.RemovePasskey(api)

	handlers.NewAPITokenHandler(deps.APITokenService, deps.Authenticator, deps.DB).RegisterRoutes(api)
	handlers.NewCLIAuthHandler(deps.CLIAuthService, deps.Authenticator, deps.PublicURL).RegisterRoutes(api)
	handlers.NewMCPActivityHandler(deps.DB, deps.Authenticator).RegisterRoutes(api)
	handlers.NewProviderAppHandler(providerapps.NewService(deps.DB, deps.TokenEncryptor), deps.DB, deps.Authenticator).RegisterRoutes(api)

	mcpOAuthHandler := deps.MCPOAuthHandler
	if mcpOAuthHandler == nil {
		mcpOAuthHandler = handlers.NewMCPOAuthHandler(deps.MCPOAuthService, deps.Authenticator, deps.PublicURL)
	}
	mcpOAuthHandler.RegisterAPIRoutes(api)

	workspaceHandler := handlers.NewWorkspaceHandler(deps.DB, deps.Authenticator, deps.Entitlement)
	workspaceHandler.SetFrontendURL(deps.FrontendURL)
	workspaceHandler.CreateWorkspace(api)
	workspaceHandler.ListWorkspaces(api)
	workspaceHandler.ListWorkspaceTeam(api)
	workspaceHandler.CreateWorkspaceInvitation(api)
	workspaceHandler.RevokeWorkspaceInvitation(api)
	workspaceHandler.AcceptWorkspaceInvitation(api)
	workspaceHandler.GetWorkspaceSettings(api)
	workspaceHandler.UpdateWorkspaceSettings(api)

	postHandler := handlers.NewPostHandler(deps.DB, deps.Authenticator, deps.Entitlement)
	postHandler.CreatePost(api)
	postHandler.CreateThread(api)
	postHandler.ListPosts(api)
	postHandler.GetPost(api)
	postHandler.UpdatePost(api)
	postHandler.DeletePost(api)
	postHandler.GetScheduleOverview(api)
	postHandler.UpsertVariants(api)
	postHandler.GetVariants(api)
	postHandler.DeleteVariants(api)

	setHandler := handlers.NewSetHandler(deps.DB, deps.Authenticator)
	setHandler.CreateSet(api)
	setHandler.ListSets(api)
	setHandler.GetSet(api)
	setHandler.UpdateSet(api)
	setHandler.DeleteSet(api)
	setHandler.AddSetAccounts(api)
	setHandler.RemoveSetAccount(api)

	postingScheduleHandler := handlers.NewPostingScheduleHandler(deps.DB, deps.Authenticator)
	postingScheduleHandler.ListSchedules(api)
	postingScheduleHandler.CreateSchedule(api)
	postingScheduleHandler.UpdateSchedule(api)
	postingScheduleHandler.DeleteSchedule(api)
	postingScheduleHandler.SuggestSchedule(api)
	postingScheduleHandler.GetNextAvailableSlot(api)

	promptHandler := handlers.NewPromptHandler(deps.DB, deps.Authenticator)
	promptHandler.ListPrompts(api)
	promptHandler.CreatePrompt(api)
	promptHandler.DeletePrompt(api)
	promptHandler.GetRandomPrompt(api)
	promptHandler.GetCategories(api)

	handlers.NewJobHandler(deps.DB, deps.Authenticator).RegisterRoutes(api)

	oauthHandler := handlers.NewOAuthHandler(
		deps.DB,
		deps.TokenEncryptor,
		deps.Providers,
		deps.Authenticator,
		deps.DisableLinkedInThreadReplies,
		deps.FrontendURL,
	)
	oauthHandler.SetEntitlement(deps.Entitlement)
	oauthHandler.SetMastodonAppService(deps.MastodonAppService)
	oauthHandler.ListProviders(api)
	oauthHandler.ListMastodonServers(api)
	oauthHandler.GetAuthURL(api)
	oauthHandler.Callback(api)
	oauthHandler.ExchangeCode(api)
	oauthHandler.BlueskyLogin(api)
	oauthHandler.GetAccountSelection(api)
	oauthHandler.CompleteAccountSelection(api)
	oauthHandler.ListAccounts(api)
	oauthHandler.UpdateAccount(api)
	oauthHandler.DisconnectAccount(api)

	RegisterHealth(api, deps.DB)
}

func RegisterHealth(api huma.API, db *bun.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, func(_ context.Context, _ *struct{}) (*struct {
		Body struct {
			Status string `json:"status" doc:"Health status"`
		}
	}, error) {
		resp := &struct {
			Body struct {
				Status string `json:"status" doc:"Health status"`
			}
		}{}
		resp.Body.Status = "ok"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "readiness-check",
		Method:      http.MethodGet,
		Path:        "/ready",
		Summary:     "Readiness check",
		Tags:        []string{"System"},
		Errors:      []int{503},
	}, func(ctx context.Context, _ *struct{}) (*struct {
		Body struct {
			Status   string `json:"status" doc:"Readiness status"`
			Database string `json:"database" doc:"Database dependency status"`
		}
	}, error) {
		if db == nil {
			return nil, huma.NewError(http.StatusServiceUnavailable, "database is not ready")
		}
		var one int
		if err := db.NewSelect().ColumnExpr("1").Scan(ctx, &one); err != nil {
			return nil, huma.NewError(http.StatusServiceUnavailable, "database is not ready")
		}
		resp := &struct {
			Body struct {
				Status   string `json:"status" doc:"Readiness status"`
				Database string `json:"database" doc:"Database dependency status"`
			}
		}{}
		resp.Body.Status = "ready"
		resp.Body.Database = "ok"
		return resp, nil
	})
}
