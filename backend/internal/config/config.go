package config

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type MastodonServerConfig struct {
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	InstanceURL  string `json:"instance_url"`
}

type Config struct {
	Port                 string
	DatabasePath         string
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

	MediaPath string
	MediaURL  string
}

const minSecretLength = 32

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
		Port:                 getEnvWithFallbacks("OPENPOST_PORT", "8080"),
		DatabasePath:         getEnvWithFallbacks("OPENPOST_DATABASE_PATH", "file:openpost.db?cache=shared&mode=rwc", "OPENPOST_DB_PATH"),
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

		MediaPath: getEnvDefault("OPENPOST_MEDIA_PATH", "./media"),
		MediaURL:  getEnvDefault("OPENPOST_MEDIA_URL", "/media"),
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

	return cfg
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
