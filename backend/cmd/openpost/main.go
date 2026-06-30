package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	apiroutes "github.com/openpost/backend/internal/api"
	"github.com/openpost/backend/internal/api/handlers"
	apimiddleware "github.com/openpost/backend/internal/api/middleware"
	"github.com/openpost/backend/internal/config"
	"github.com/openpost/backend/internal/database"
	"github.com/openpost/backend/internal/platform"
	"github.com/openpost/backend/internal/queue"
	"github.com/openpost/backend/internal/services/apitokens"
	"github.com/openpost/backend/internal/services/auth"
	"github.com/openpost/backend/internal/services/billing"
	cliauth "github.com/openpost/backend/internal/services/cli_auth"
	"github.com/openpost/backend/internal/services/crypto"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/mastodonapps"
	"github.com/openpost/backend/internal/services/mcpoauth"
	"github.com/openpost/backend/internal/services/mediasigner"
	"github.com/openpost/backend/internal/services/mediastore"
	"github.com/openpost/backend/internal/services/mfa"
	"github.com/openpost/backend/internal/services/publisher"
	"github.com/openpost/backend/internal/services/tokenmanager"
)

//nolint:gocyclo
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()
	config.Init()
	if err := cfg.ValidateRuntime(); err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	db, err := database.InitDBWithDriver(cfg.DatabaseDriver, cfg.DatabaseDSN())
	if err != nil {
		log.Fatal(err)
	}
	if err := database.CreateSchema(db); err != nil {
		log.Printf("CreateSchema error (already exists?): %v", err)
	}

	tokenEncryptor := crypto.NewTokenEncryptor(cfg.EncryptionKey)
	authService := auth.NewService(cfg.JWTSecret)
	apiTokenService := apitokens.NewService(db)
	billingService := billing.NewService(db, cfg.PolarWebhookSecret, billing.PolarConfig{
		AccessToken: cfg.PolarAccessToken,
		SuccessURL:  cfg.PolarCheckoutURL,
		ReturnURL:   cfg.PolarReturnURL,
		Plans: billing.DefaultPlanCatalog(
			cfg.PolarStarterProductID,
			cfg.PolarCreatorProductID,
			cfg.PolarProProductID,
		),
	})
	entitlementService := entitlements.Service(entitlements.NewSelfHostedService())
	if cfg.Edition == config.EditionCloud {
		entitlementService = entitlements.NewSubscriptionService(db, entitlements.NewCloudBootstrapService())
	}
	authenticator := apimiddleware.NewCompositeService(authService, apiTokenService)
	cliAuthService := cliauth.NewService(db, apiTokenService)
	mcpOAuthService := mcpoauth.NewService(db, apiTokenService)
	mediaSigner := mediasigner.New(cfg.EncryptionKey)
	mfaService, err := mfa.NewService("OpenPost", mfa.RelyingPartyConfig{
		Name:    "OpenPost",
		ID:      cfg.WebAuthnRPID,
		Origins: []string{cfg.PublicURL},
	})
	if err != nil {
		log.Fatal(err)
	}
	tokenManager := tokenmanager.NewTokenManager(db, tokenEncryptor)
	publishSvc := publisher.NewService(db, tokenManager)
	publishSvc.SetEntitlement(entitlementService)
	publishSvc.SetDisableLinkedInThreadReplies(cfg.DisableLinkedInThreadReplies)
	publishSvc.SetMediaSigner(mediaSigner)
	if cfg.MediaURL != "" && !strings.HasPrefix(cfg.MediaURL, "/") {
		publishSvc.SetPublicMediaURL(cfg.MediaURL)
	}

	platform.RegisterAllMediaValidators()
	providers, providerEntries, err := platform.BuildAdapterRegistry(cfg.ProviderApps, platform.RegistryOptions{
		DisableLinkedInThreadReplies: cfg.DisableLinkedInThreadReplies,
	})
	if err != nil {
		log.Fatalf("failed to build provider app registry: %v", err)
	}
	for _, entry := range providerEntries {
		log.Printf("Registered provider adapter: %s", entry.Key)
	}

	for name, adapter := range providers {
		tokenManager.SetProvider(name, adapter)
		publishSvc.SetProvider(name, adapter)
	}

	storage, err := mediastore.New(context.Background(), mediastore.Config{
		Driver:    cfg.StorageDriver,
		LocalPath: cfg.MediaPath,
		BaseURL:   cfg.MediaURL,
		S3: mediastore.S3Config{
			Endpoint:        cfg.S3Endpoint,
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			PublicBaseURL:   cfg.S3PublicBaseURL,
			ForcePathStyle:  cfg.S3ForcePathStyle,
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize media storage: %v", err)
	}
	publishSvc.SetStorage(storage)
	mediaHandler := handlers.NewMediaHandler(db, storage, authService, authenticator, mediaSigner)
	mediaHandler.SetEntitlement(entitlementService)

	worker := queue.NewWorker(db, "worker-1", 1*time.Second, publishSvc, tokenManager, storage)

	apiGroup := e.Group("/api/v1")
	humaConfig := huma.DefaultConfig("OpenPost API", "1.0.0")
	api := humaecho.NewWithGroup(e, apiGroup, humaConfig)

	mediaHandler.RegisterLegacyRoutes(e)
	billingHandler := handlers.NewBillingHandler(billingService, db, authenticator)
	billingHandler.RegisterRoutes(e)

	e.GET("/openapi.json", func(c echo.Context) error {
		spec := api.OpenAPI()
		data, err := json.Marshal(spec)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to marshal spec"})
		}
		return c.Blob(http.StatusOK, "application/json", data)
	})

	robotsHandler := func(c echo.Context) error {
		robots := "User-agent: *\nAllow: /\nUser-agent: facebookexternalhit\nAllow: /\nUser-agent: Twitterbot\nAllow: /\nUser-agent: LinkedInBot\nAllow: /"
		return c.String(http.StatusOK, robots)
	}

	e.GET("/robots.txt", robotsHandler)
	e.HEAD("/robots.txt", robotsHandler)

	mastodonAppService := mastodonapps.NewService(db, tokenEncryptor, mastodonapps.Options{
		RedirectURI: cfg.MastodonRedirectURI,
		Website:     cfg.PublicURL,
	})

	mcpHandler := handlers.NewMCPHandler(db, authenticator, entitlementService)
	mcpHandler.SetMediaStorage(storage)
	mcpHandler.SetPublicURL(cfg.PublicURL)
	mcpHandler.SetProviderCatalog(providers, mastodonAppService != nil)
	mcpHandler.RegisterRoutes(e)
	mcpOAuthHandler := handlers.NewMCPOAuthHandler(mcpOAuthService, authenticator, cfg.PublicURL)
	mcpOAuthHandler.RegisterEchoRoutes(e)

	apiroutes.RegisterHumaRoutes(api, apiroutes.RouteDeps{
		DB:                           db,
		AuthService:                  authService,
		Authenticator:                authenticator,
		APITokenService:              apiTokenService,
		CLIAuthService:               cliAuthService,
		MCPOAuthService:              mcpOAuthService,
		BillingService:               billingService,
		MediaStorage:                 storage,
		MediaSigner:                  mediaSigner,
		Entitlement:                  entitlementService,
		TokenEncryptor:               tokenEncryptor,
		MFAService:                   mfaService,
		Providers:                    providers,
		MastodonAppService:           mastodonAppService,
		FrontendURL:                  cfg.FrontendURL,
		PublicURL:                    cfg.PublicURL,
		DisableRegistrations:         cfg.DisableRegistrations,
		DisableLinkedInThreadReplies: cfg.DisableLinkedInThreadReplies,
		MediaHandler:                 mediaHandler,
		BillingHandler:               billingHandler,
		MCPOAuthHandler:              mcpOAuthHandler,
	})

	RegisterSpaRoutes(e)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()

	log.Println("Starting OpenPost on :" + cfg.Port)
	log.Println("OpenAPI spec available at http://localhost:" + cfg.Port + "/openapi.json")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- e.Start(":" + cfg.Port)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("Shutting down after %s...", sig)
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			cancel()
			signal.Stop(sigCh)
			worker.Stop()
			wg.Wait()
			log.Printf("Server error: %v", err)
			return
		}
		cancel()
		worker.Stop()
		wg.Wait()
		log.Println("Server stopped")
		return
	}

	cancel()
	worker.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("Echo shutdown error: %v", err)
	}

	if err := <-serverErrCh; err != nil && err != http.ErrServerClosed {
		log.Printf("Echo server error: %v", err)
	}

	wg.Wait()
	log.Println("Server stopped")
}
