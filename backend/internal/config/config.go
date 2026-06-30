package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/openpost/backend/internal/platform"
)

type MastodonServerConfig struct {
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	InstanceURL  string `json:"instance_url"`
}

type Config struct {
	Edition              string
	Port                 string
	DatabaseDriver       string
	DatabasePath         string
	DatabaseURL          string
	JWTSecret            string
	EncryptionKey        string
	DisableRegistrations bool
	FrontendURL          string
	PublicURL            string
	CORSOrigins          []string
	WebAuthnRPID         string

	TwitterClientID     string
	TwitterClientSecret string
	TwitterRedirectURI  string

	MastodonRedirectURI string
	MastodonServers     []MastodonServerConfig

	LinkedInClientID             string
	LinkedInClientSecret         string
	LinkedInRedirectURI          string
	DisableLinkedInThreadReplies bool

	ThreadsClientID     string
	ThreadsClientSecret string
	ThreadsRedirectURI  string

	ProviderApps []platform.AppConfig

	StorageDriver     string
	MediaPath         string
	MediaURL          string
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3PublicBaseURL   string
	S3ForcePathStyle  bool

	PolarAccessToken      string
	PolarAPIBaseURL       string
	PolarWebhookSecret    string
	PolarCheckoutURL      string
	PolarReturnURL        string
	PolarStarterProductID string
	PolarCreatorProductID string
	PolarProProductID     string
}

const minSecretLength = 32

const (
	EditionSelfHost = "selfhost"
	EditionCloud    = "cloud"

	DatabaseDriverSQLite   = "sqlite"
	DatabaseDriverPostgres = "postgres"

	StorageDriverLocal = "local"
	StorageDriverS3    = "s3"
)

func Load() *Config {
	// FrontendURL is computed up front so the platform-specific OAuth
	// redirect URIs can be derived from it when the operator hasn't set
	// the *_REDIRECT_URI env vars. This avoids the previous footgun
	// where copying `.env.example` to `.env` produced a working-looking
	// setup that emitted OAuth callbacks pointing at localhost:5173
	// (Vite's dev port) regardless of where the binary was actually
	// deployed.
	frontendURL := strings.TrimRight(getEnvWithFallbacks("OPENPOST_APP_URL", "http://localhost:8080", "OPENPOST_FRONTEND_URL"), "/")

	cfg := &Config{
		Edition:              getEnvEnum("OPENPOST_EDITION", EditionSelfHost, EditionSelfHost, EditionCloud),
		Port:                 getEnvWithFallbacks("OPENPOST_PORT", "8080"),
		DatabaseDriver:       getEnvEnum("OPENPOST_DATABASE_DRIVER", DatabaseDriverSQLite, DatabaseDriverSQLite, DatabaseDriverPostgres),
		DatabasePath:         getEnvWithFallbacks("OPENPOST_DATABASE_PATH", "file:openpost.db?cache=shared&mode=rwc", "OPENPOST_DB_PATH"),
		DatabaseURL:          getEnvWithFallbacks("OPENPOST_DATABASE_URL", "", "DATABASE_URL"),
		JWTSecret:            getEnvWithFallbacks("OPENPOST_JWT_SECRET", "", "JWT_SECRET"),
		EncryptionKey:        getEnvWithFallbacks("OPENPOST_ENCRYPTION_KEY", "", "ENCRYPTION_KEY"),
		DisableRegistrations: getEnvBoolWithAliases(false, "OPENPOST_DISABLE_REGISTRATIONS"),
		FrontendURL:          frontendURL,
		PublicURL:            getEnvWithFallbacks("OPENPOST_PUBLIC_URL", "", "OPENPOST_APP_URL", "OPENPOST_FRONTEND_URL"),

		TwitterClientID:     getEnvWithFallbacks("X_CLIENT_ID", "", "TWITTER_CLIENT_ID"),
		TwitterClientSecret: getEnvWithFallbacks("X_CLIENT_SECRET", "", "TWITTER_CLIENT_SECRET"),
		TwitterRedirectURI:  oauthRedirectFromFrontend("X_REDIRECT_URI", "TWITTER_REDIRECT_URI", frontendURL, "/api/v1/accounts/x/callback"),

		// Mastodon's OOB flow uses a special URI scheme rather than a
		// real callback URL, so we don't derive from FrontendURL here.
		// Operators who need a real URL can still override via env.
		MastodonRedirectURI: getEnvDefault("MASTODON_REDIRECT_URI", "urn:ietf:wg:oauth:2.0:oob"),

		LinkedInClientID:             getEnvWithFallbacks("LINKEDIN_CLIENT_ID", ""),
		LinkedInClientSecret:         getEnvWithFallbacks("LINKEDIN_CLIENT_SECRET", ""),
		LinkedInRedirectURI:          oauthRedirectFromFrontend("LINKEDIN_REDIRECT_URI", "", frontendURL, "/api/v1/accounts/linkedin/callback"),
		DisableLinkedInThreadReplies: getEnvBoolWithAliases(false, "LINKEDIN_DISABLE_THREAD_REPLIES", "OPENPOST_DISABLE_LINKEDIN_THREAD_REPLIES"),

		ThreadsClientID:     getEnvWithFallbacks("THREADS_CLIENT_ID", ""),
		ThreadsClientSecret: getEnvWithFallbacks("THREADS_CLIENT_SECRET", ""),
		ThreadsRedirectURI:  oauthRedirectFromFrontend("THREADS_REDIRECT_URI", "", frontendURL, "/api/v1/accounts/threads/callback"),

		StorageDriver:     getEnvEnum("OPENPOST_STORAGE_DRIVER", StorageDriverLocal, StorageDriverLocal, StorageDriverS3),
		MediaPath:         getEnvDefault("OPENPOST_MEDIA_PATH", "./media"),
		MediaURL:          getEnvDefault("OPENPOST_MEDIA_URL", "/media"),
		S3Endpoint:        getEnvDefault("OPENPOST_S3_ENDPOINT", ""),
		S3Region:          getEnvDefault("OPENPOST_S3_REGION", ""),
		S3Bucket:          getEnvDefault("OPENPOST_S3_BUCKET", ""),
		S3AccessKeyID:     getEnvDefault("OPENPOST_S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getEnvDefault("OPENPOST_S3_SECRET_ACCESS_KEY", ""),
		S3PublicBaseURL:   strings.TrimRight(getEnvDefault("OPENPOST_S3_PUBLIC_BASE_URL", ""), "/"),
		S3ForcePathStyle:  getEnvBoolWithAliases(false, "OPENPOST_S3_FORCE_PATH_STYLE"),

		PolarAccessToken:      getEnvDefault("OPENPOST_POLAR_ACCESS_TOKEN", ""),
		PolarAPIBaseURL:       strings.TrimRight(getEnvDefault("OPENPOST_POLAR_API_BASE_URL", "https://api.polar.sh/v1"), "/"),
		PolarWebhookSecret:    getEnvDefault("OPENPOST_POLAR_WEBHOOK_SECRET", ""),
		PolarCheckoutURL:      strings.TrimRight(getEnvDefault("OPENPOST_POLAR_CHECKOUT_SUCCESS_URL", ""), "/"),
		PolarReturnURL:        strings.TrimRight(getEnvWithFallbacks("OPENPOST_POLAR_RETURN_URL", "", "OPENPOST_POLAR_CUSTOMER_PORTAL_URL"), "/"),
		PolarStarterProductID: getEnvDefault("OPENPOST_POLAR_STARTER_PRODUCT_ID", ""),
		PolarCreatorProductID: getEnvDefault("OPENPOST_POLAR_CREATOR_PRODUCT_ID", ""),
		PolarProProductID:     getEnvDefault("OPENPOST_POLAR_PRO_PRODUCT_ID", ""),
	}

	if cfg.PublicURL == "" {
		cfg.PublicURL = cfg.FrontendURL
	}
	if parsed, err := url.Parse(cfg.PublicURL); err == nil && parsed.Hostname() != "" {
		cfg.WebAuthnRPID = parsed.Hostname()
	} else {
		cfg.WebAuthnRPID = "localhost"
	}

	if raw := os.Getenv("MASTODON_SERVERS"); raw != "" {
		var servers []MastodonServerConfig
		if err := json.Unmarshal([]byte(raw), &servers); err != nil {
			log.Printf("WARNING: failed to parse MASTODON_SERVERS JSON: %v", err)
		} else {
			cfg.MastodonServers = servers
		}
	}
	cfg.ProviderApps = providerAppsFromLegacyConfig(cfg)
	if raw := os.Getenv("OPENPOST_PROVIDER_APPS"); raw != "" {
		var apps []platform.AppConfig
		if err := json.Unmarshal([]byte(raw), &apps); err != nil {
			log.Printf("WARNING: failed to parse OPENPOST_PROVIDER_APPS JSON: %v", err)
		} else {
			cfg.ProviderApps = mergeProviderApps(cfg.ProviderApps, defaultProviderAppConfig(cfg, apps)...)
		}
	}

	// Build CORS origins list
	corsOrigins := []string{cfg.FrontendURL, "http://localhost:5173"}
	if extra := getEnvWithFallbacks("OPENPOST_EXTRA_CORS_ORIGINS", "", "OPENPOST_CORS_EXTRA_ORIGINS"); extra != "" {
		for _, origin := range strings.Split(extra, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				corsOrigins = append(corsOrigins, trimmed)
			}
		}
	}
	// Always allow Capacitor origins
	corsOrigins = append(corsOrigins, "capacitor://localhost", "http://localhost", "https://localhost")
	cfg.CORSOrigins = corsOrigins

	warnOnPlaceholderURL(cfg)

	return cfg
}

func providerAppsFromLegacyConfig(cfg *Config) []platform.AppConfig {
	apps := []platform.AppConfig{{Provider: "bluesky"}}
	if cfg.TwitterClientID != "" {
		apps = append(apps, platform.AppConfig{
			Provider:     "x",
			ClientID:     cfg.TwitterClientID,
			ClientSecret: cfg.TwitterClientSecret,
			RedirectURI:  cfg.TwitterRedirectURI,
		})
	}
	for _, server := range cfg.MastodonServers {
		apps = append(apps, platform.AppConfig{
			Provider:     "mastodon",
			Name:         server.Name,
			ClientID:     server.ClientID,
			ClientSecret: server.ClientSecret,
			RedirectURI:  cfg.MastodonRedirectURI,
			InstanceURL:  server.InstanceURL,
		})
	}
	if cfg.LinkedInClientID != "" {
		apps = append(apps, platform.AppConfig{
			Provider:     "linkedin",
			ClientID:     cfg.LinkedInClientID,
			ClientSecret: cfg.LinkedInClientSecret,
			RedirectURI:  cfg.LinkedInRedirectURI,
		})
	}
	if cfg.ThreadsClientID != "" {
		apps = append(apps, platform.AppConfig{
			Provider:     "threads",
			ClientID:     cfg.ThreadsClientID,
			ClientSecret: cfg.ThreadsClientSecret,
			RedirectURI:  cfg.ThreadsRedirectURI,
		})
	}
	return defaultProviderAppConfig(cfg, apps)
}

func defaultProviderAppConfig(cfg *Config, apps []platform.AppConfig) []platform.AppConfig {
	out := make([]platform.AppConfig, 0, len(apps))
	for _, app := range apps {
		app = platform.NormalizeAppConfig(app)
		if app.RedirectURI == "" {
			app.RedirectURI = providerRedirectURI(cfg, app.Provider)
		}
		out = append(out, app)
	}
	return out
}

func providerRedirectURI(cfg *Config, provider string) string {
	redirects := map[string]string{
		"x":        cfg.TwitterRedirectURI,
		"mastodon": cfg.MastodonRedirectURI,
		"linkedin": cfg.LinkedInRedirectURI,
		"threads":  cfg.ThreadsRedirectURI,
	}
	return redirects[provider]
}

func mergeProviderApps(base []platform.AppConfig, overrides ...platform.AppConfig) []platform.AppConfig {
	merged := append([]platform.AppConfig{}, base...)
	indexByKey := make(map[string]int, len(merged))
	for i, app := range merged {
		indexByKey[providerAppMergeKey(app)] = i
	}
	for _, app := range overrides {
		key := providerAppMergeKey(app)
		if i, ok := indexByKey[key]; ok {
			merged[i] = app
			continue
		}
		indexByKey[key] = len(merged)
		merged = append(merged, app)
	}
	return merged
}

func providerAppMergeKey(app platform.AppConfig) string {
	app = platform.NormalizeAppConfig(app)
	keys := map[string]string{
		"mastodon": app.Provider + ":" + app.InstanceURL,
	}
	if key, ok := keys[app.Provider]; ok {
		return key
	}
	return app.Provider
}

func (c *Config) DatabaseDSN() string {
	if c.DatabaseDriver == DatabaseDriverPostgres && c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return c.DatabasePath
}

func (c *Config) ValidateRuntime() error {
	if c.Edition != EditionCloud {
		return nil
	}

	missing := append(c.missingCloudDataPlaneConfig(), c.missingCloudBillingConfig()...)
	if len(missing) > 0 {
		return fmt.Errorf("OPENPOST_EDITION=cloud requires: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *Config) missingCloudDataPlaneConfig() []string {
	missing := make([]string, 0, 8)
	if c.DatabaseDriver != DatabaseDriverPostgres {
		missing = append(missing, "OPENPOST_DATABASE_DRIVER=postgres")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		missing = append(missing, "OPENPOST_DATABASE_URL")
	}
	if c.StorageDriver != StorageDriverS3 {
		missing = append(missing, "OPENPOST_STORAGE_DRIVER=s3")
	}
	if strings.TrimSpace(c.S3Region) == "" {
		missing = append(missing, "OPENPOST_S3_REGION")
	}
	if strings.TrimSpace(c.S3Bucket) == "" {
		missing = append(missing, "OPENPOST_S3_BUCKET")
	}
	if strings.TrimSpace(c.S3AccessKeyID) == "" {
		missing = append(missing, "OPENPOST_S3_ACCESS_KEY_ID")
	}
	if strings.TrimSpace(c.S3SecretAccessKey) == "" {
		missing = append(missing, "OPENPOST_S3_SECRET_ACCESS_KEY")
	}
	if strings.TrimSpace(c.S3PublicBaseURL) == "" {
		missing = append(missing, "OPENPOST_S3_PUBLIC_BASE_URL")
	}
	return missing
}

func (c *Config) missingCloudBillingConfig() []string {
	missing := make([]string, 0, 7)
	if strings.TrimSpace(c.PolarAccessToken) == "" {
		missing = append(missing, "OPENPOST_POLAR_ACCESS_TOKEN")
	}
	if strings.TrimSpace(c.PolarWebhookSecret) == "" {
		missing = append(missing, "OPENPOST_POLAR_WEBHOOK_SECRET")
	}
	if strings.TrimSpace(c.PolarCheckoutURL) == "" {
		missing = append(missing, "OPENPOST_POLAR_CHECKOUT_SUCCESS_URL")
	}
	if strings.TrimSpace(c.PolarReturnURL) == "" {
		missing = append(missing, "OPENPOST_POLAR_RETURN_URL")
	}
	if strings.TrimSpace(c.PolarStarterProductID) == "" {
		missing = append(missing, "OPENPOST_POLAR_STARTER_PRODUCT_ID")
	}
	if strings.TrimSpace(c.PolarCreatorProductID) == "" {
		missing = append(missing, "OPENPOST_POLAR_CREATOR_PRODUCT_ID")
	}
	if strings.TrimSpace(c.PolarProProductID) == "" {
		missing = append(missing, "OPENPOST_POLAR_PRO_PRODUCT_ID")
	}
	return missing
}

// warnOnPlaceholderURL emits a loud startup warning when the operator is
// running with the binary's default `http://localhost:8080` for
// OPENPOST_APP_URL/OPENPOST_PUBLIC_URL, which is almost always wrong in
// production. The check only fires when neither env var was set
// explicitly, so `devenv shell` and any operator who has set a real URL
// are not affected.
func warnOnPlaceholderURL(cfg *Config) {
	if _, explicit := os.LookupEnv("OPENPOST_APP_URL"); explicit {
		return
	}
	if _, explicit := os.LookupEnv("OPENPOST_FRONTEND_URL"); explicit {
		return
	}
	log.Printf("============================================================")
	log.Printf("WARNING: OPENPOST_APP_URL is not set; falling back to")
	log.Printf("         %s. OAuth callbacks, WebAuthn origins, and the", cfg.FrontendURL)
	log.Printf("         public media URL will all advertise this address.")
	log.Printf("         Set OPENPOST_APP_URL=https://your-public-host in")
	log.Printf("         production. See .env.example for details.")
	log.Printf("============================================================")
}

func getEnvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvWithFallbacks(primary, fallback string, aliases ...string) string {
	if value := os.Getenv(primary); value != "" {
		return value
	}
	for _, alias := range aliases {
		if value := os.Getenv(alias); value != "" {
			return value
		}
	}
	return fallback
}

func getEnvBoolWithAliases(fallback bool, keys ...string) bool {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}

		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("WARNING: invalid boolean for %s=%q, using default %t", key, value, fallback)
			return fallback
		}
		return parsed
	}

	return fallback
}

func getEnvEnum(key, fallback string, allowed ...string) string {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}

	log.Printf("WARNING: invalid value for %s=%q, using default %q", key, value, fallback)
	return fallback
}

// oauthRedirectFromFrontend returns the OAuth redirect URI to register
// with an external provider, preferring the explicit env var (and any
// aliases) when set. If nothing is set, it derives a sensible default
// from the FrontendURL — this prevents the footgun where copying
// `.env.example` produced OAuth callbacks pointing at localhost:5173
// (Vite's dev port) regardless of where the binary was deployed.
func oauthRedirectFromFrontend(primary, alias, frontend, path string) string {
	keys := []string{primary}
	if alias != "" {
		keys = append(keys, alias)
	}
	if v := getEnvWithFallbacks(keys[0], "", keys[1:]...); v != "" {
		return v
	}
	return strings.TrimRight(frontend, "/") + path
}

func Init() {
	jwtSecret := getEnvWithFallbacks("OPENPOST_JWT_SECRET", "", "JWT_SECRET")
	encryptionKey := getEnvWithFallbacks("OPENPOST_ENCRYPTION_KEY", "", "ENCRYPTION_KEY")

	if jwtSecret == "" {
		log.Fatal("FATAL: OPENPOST_JWT_SECRET is required")
	}
	if len(jwtSecret) < minSecretLength {
		log.Fatalf("FATAL: OPENPOST_JWT_SECRET must be at least %d characters (got %d)", minSecretLength, len(jwtSecret))
	}
	if encryptionKey == "" {
		log.Fatal("FATAL: OPENPOST_ENCRYPTION_KEY is required")
	}
	if len(encryptionKey) < minSecretLength {
		log.Fatalf("FATAL: OPENPOST_ENCRYPTION_KEY must be at least %d characters (got %d)", minSecretLength, len(encryptionKey))
	}
}
