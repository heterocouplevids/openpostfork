package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadProductionPrimitiveDefaults(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://openpost.example.com")

	cfg := Load()

	require.Equal(t, EditionSelfHost, cfg.Edition)
	require.Equal(t, DatabaseDriverSQLite, cfg.DatabaseDriver)
	require.Equal(t, "file:openpost.db?cache=shared&mode=rwc", cfg.DatabaseDSN())
	require.Equal(t, StorageDriverLocal, cfg.StorageDriver)
	require.Empty(t, cfg.DatabaseURL)
	require.Empty(t, cfg.S3Bucket)
	require.Empty(t, cfg.PolarAccessToken)
	require.Equal(t, "https://api.polar.sh/v1", cfg.PolarAPIBaseURL)
	require.Empty(t, cfg.PolarWebhookSecret)
}

func TestLoadCloudPostgresAndS3Primitives(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://app.openpost.social")
	t.Setenv("OPENPOST_EDITION", "cloud")
	t.Setenv("OPENPOST_DATABASE_DRIVER", "postgres")
	t.Setenv("OPENPOST_DATABASE_URL", "postgres://openpost:secret@db.internal:5432/openpost?sslmode=require")
	t.Setenv("OPENPOST_STORAGE_DRIVER", "s3")
	t.Setenv("OPENPOST_S3_ENDPOINT", "https://r2.example.com")
	t.Setenv("OPENPOST_S3_REGION", "auto")
	t.Setenv("OPENPOST_S3_BUCKET", "openpost-media")
	t.Setenv("OPENPOST_S3_ACCESS_KEY_ID", "access-key")
	t.Setenv("OPENPOST_S3_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("OPENPOST_S3_PUBLIC_BASE_URL", "https://media.openpost.social")
	t.Setenv("OPENPOST_S3_FORCE_PATH_STYLE", "true")

	cfg := Load()

	require.Equal(t, EditionCloud, cfg.Edition)
	require.Equal(t, DatabaseDriverPostgres, cfg.DatabaseDriver)
	require.Equal(t, "postgres://openpost:secret@db.internal:5432/openpost?sslmode=require", cfg.DatabaseDSN())
	require.Equal(t, StorageDriverS3, cfg.StorageDriver)
	require.Equal(t, "https://r2.example.com", cfg.S3Endpoint)
	require.Equal(t, "auto", cfg.S3Region)
	require.Equal(t, "openpost-media", cfg.S3Bucket)
	require.Equal(t, "access-key", cfg.S3AccessKeyID)
	require.Equal(t, "secret-key", cfg.S3SecretAccessKey)
	require.Equal(t, "https://media.openpost.social", cfg.S3PublicBaseURL)
	require.True(t, cfg.S3ForcePathStyle)
}

func TestValidateRuntimeAllowsSelfHostedLocalDefaults(t *testing.T) {
	cfg := &Config{
		Edition:        EditionSelfHost,
		DatabaseDriver: DatabaseDriverSQLite,
		DatabasePath:   "file:openpost.db?cache=shared&mode=rwc",
		StorageDriver:  StorageDriverLocal,
	}

	require.NoError(t, cfg.ValidateRuntime())
}

func TestValidateRuntimeAllowsCloudPostgresAndS3(t *testing.T) {
	cfg := validCloudRuntimeConfig()

	require.NoError(t, cfg.ValidateRuntime())
}

func TestValidateRuntimeRejectsCloudLocalDefaults(t *testing.T) {
	cfg := &Config{
		Edition:        EditionCloud,
		DatabaseDriver: DatabaseDriverSQLite,
		DatabasePath:   "file:openpost.db?cache=shared&mode=rwc",
		StorageDriver:  StorageDriverLocal,
	}

	err := cfg.ValidateRuntime()

	require.Error(t, err)
	require.ErrorContains(t, err, "OPENPOST_EDITION=cloud")
	require.ErrorContains(t, err, "OPENPOST_DATABASE_DRIVER=postgres")
	require.ErrorContains(t, err, "OPENPOST_DATABASE_URL")
	require.ErrorContains(t, err, "OPENPOST_STORAGE_DRIVER=s3")
}

func TestValidateRuntimeRejectsCloudMissingS3Primitives(t *testing.T) {
	cfg := validCloudRuntimeConfig()
	cfg.S3Region = ""
	cfg.S3Bucket = ""
	cfg.S3AccessKeyID = ""
	cfg.S3SecretAccessKey = ""
	cfg.S3PublicBaseURL = ""

	err := cfg.ValidateRuntime()

	require.Error(t, err)
	require.ErrorContains(t, err, "OPENPOST_S3_REGION")
	require.ErrorContains(t, err, "OPENPOST_S3_BUCKET")
	require.ErrorContains(t, err, "OPENPOST_S3_ACCESS_KEY_ID")
	require.ErrorContains(t, err, "OPENPOST_S3_SECRET_ACCESS_KEY")
	require.ErrorContains(t, err, "OPENPOST_S3_PUBLIC_BASE_URL")
}

func TestValidateRuntimeRejectsCloudMissingPolarPrimitives(t *testing.T) {
	cfg := validCloudRuntimeConfig()
	cfg.PolarAccessToken = ""
	cfg.PolarWebhookSecret = ""
	cfg.PolarCheckoutURL = ""
	cfg.PolarReturnURL = ""
	cfg.PolarStarterProductID = ""
	cfg.PolarCreatorProductID = ""
	cfg.PolarProProductID = ""

	err := cfg.ValidateRuntime()

	require.Error(t, err)
	require.ErrorContains(t, err, "OPENPOST_POLAR_ACCESS_TOKEN")
	require.ErrorContains(t, err, "OPENPOST_POLAR_WEBHOOK_SECRET")
	require.ErrorContains(t, err, "OPENPOST_POLAR_CHECKOUT_SUCCESS_URL")
	require.ErrorContains(t, err, "OPENPOST_POLAR_RETURN_URL")
	require.ErrorContains(t, err, "OPENPOST_POLAR_STARTER_PRODUCT_ID")
	require.ErrorContains(t, err, "OPENPOST_POLAR_CREATOR_PRODUCT_ID")
	require.ErrorContains(t, err, "OPENPOST_POLAR_PRO_PRODUCT_ID")
}

func validCloudRuntimeConfig() *Config {
	return &Config{
		Edition:               EditionCloud,
		DatabaseDriver:        DatabaseDriverPostgres,
		DatabaseURL:           "postgres://openpost:secret@db.internal:5432/openpost?sslmode=require",
		StorageDriver:         StorageDriverS3,
		S3Region:              "auto",
		S3Bucket:              "openpost-media",
		S3AccessKeyID:         "access-key",
		S3SecretAccessKey:     "secret-key",
		S3PublicBaseURL:       "https://media.openpost.social",
		S3ForcePathStyle:      true,
		PolarAccessToken:      "polar-token",
		PolarWebhookSecret:    "whsec_secret",
		PolarCheckoutURL:      "https://app.openpost.social/settings/billing?checkout_id={CHECKOUT_ID}",
		PolarReturnURL:        "https://app.openpost.social/settings/billing",
		PolarStarterProductID: "starter-product",
		PolarCreatorProductID: "creator-product",
		PolarProProductID:     "pro-product",
	}
}

func TestLoadPolarPrimitives(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://app.openpost.social")
	t.Setenv("OPENPOST_POLAR_ACCESS_TOKEN", "polar-token")
	t.Setenv("OPENPOST_POLAR_API_BASE_URL", "https://sandbox-api.polar.sh/v1/")
	t.Setenv("OPENPOST_POLAR_WEBHOOK_SECRET", "whsec_secret")
	t.Setenv("OPENPOST_POLAR_CHECKOUT_SUCCESS_URL", "https://app.openpost.social/settings/billing/")
	t.Setenv("OPENPOST_POLAR_RETURN_URL", "https://app.openpost.social/settings/billing/")
	t.Setenv("OPENPOST_POLAR_STARTER_PRODUCT_ID", "starter-product")
	t.Setenv("OPENPOST_POLAR_CREATOR_PRODUCT_ID", "creator-product")
	t.Setenv("OPENPOST_POLAR_PRO_PRODUCT_ID", "pro-product")

	cfg := Load()

	require.Equal(t, "polar-token", cfg.PolarAccessToken)
	require.Equal(t, "https://sandbox-api.polar.sh/v1", cfg.PolarAPIBaseURL)
	require.Equal(t, "whsec_secret", cfg.PolarWebhookSecret)
	require.Equal(t, "https://app.openpost.social/settings/billing", cfg.PolarCheckoutURL)
	require.Equal(t, "https://app.openpost.social/settings/billing", cfg.PolarReturnURL)
	require.Equal(t, "starter-product", cfg.PolarStarterProductID)
	require.Equal(t, "creator-product", cfg.PolarCreatorProductID)
	require.Equal(t, "pro-product", cfg.PolarProProductID)
}

func TestLoadPolarReturnURLLegacyAlias(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://app.openpost.social")
	t.Setenv("OPENPOST_POLAR_CUSTOMER_PORTAL_URL", "https://app.openpost.social/settings/billing/")

	cfg := Load()

	require.Equal(t, "https://app.openpost.social/settings/billing", cfg.PolarReturnURL)
}

func TestLoadBuildsProviderAppRegistryFromLegacyEnv(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://app.openpost.social")
	t.Setenv("X_CLIENT_ID", "x-client")
	t.Setenv("X_CLIENT_SECRET", "x-secret")
	t.Setenv("LINKEDIN_CLIENT_ID", "linkedin-client")
	t.Setenv("LINKEDIN_CLIENT_SECRET", "linkedin-secret")
	t.Setenv("THREADS_CLIENT_ID", "threads-client")
	t.Setenv("THREADS_CLIENT_SECRET", "threads-secret")
	t.Setenv("MASTODON_SERVERS", `[{"name":"Personal","client_id":"masto-client","client_secret":"masto-secret","instance_url":"https://masto.pt/"}]`)

	cfg := Load()

	require.Len(t, cfg.ProviderApps, 5)
	require.Equal(t, "bluesky", cfg.ProviderApps[0].Provider)
	require.Equal(t, "x", cfg.ProviderApps[1].Provider)
	require.Equal(t, "https://app.openpost.social/api/v1/accounts/x/callback", cfg.ProviderApps[1].RedirectURI)
	require.Equal(t, "mastodon", cfg.ProviderApps[2].Provider)
	require.Equal(t, "https://masto.pt", cfg.ProviderApps[2].InstanceURL)
	require.Equal(t, "linkedin", cfg.ProviderApps[3].Provider)
	require.Equal(t, "https://app.openpost.social/api/v1/accounts/linkedin/callback", cfg.ProviderApps[3].RedirectURI)
	require.Equal(t, "threads", cfg.ProviderApps[4].Provider)
	require.Equal(t, "https://app.openpost.social/api/v1/accounts/threads/callback", cfg.ProviderApps[4].RedirectURI)
}

func TestLoadMergesStructuredProviderApps(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://app.openpost.social")
	t.Setenv("X_CLIENT_ID", "legacy-x-client")
	t.Setenv("X_CLIENT_SECRET", "legacy-x-secret")
	t.Setenv("OPENPOST_PROVIDER_APPS", `[
		{"provider":"x","client_id":"cloud-x-client","client_secret":"cloud-x-secret"},
		{"provider":"mastodon","name":"Community","client_id":"masto-client","client_secret":"masto-secret","instance_url":"https://community.example"}
	]`)

	cfg := Load()

	require.Len(t, cfg.ProviderApps, 3)
	require.Equal(t, "bluesky", cfg.ProviderApps[0].Provider)
	require.Equal(t, "cloud-x-client", cfg.ProviderApps[1].ClientID)
	require.Equal(t, "https://app.openpost.social/api/v1/accounts/x/callback", cfg.ProviderApps[1].RedirectURI)
	require.Equal(t, "mastodon", cfg.ProviderApps[2].Provider)
	require.Equal(t, "urn:ietf:wg:oauth:2.0:oob", cfg.ProviderApps[2].RedirectURI)
}

func TestLoadInvalidProductionPrimitiveEnumsFallback(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://openpost.example.com")
	t.Setenv("OPENPOST_EDITION", "enterprise")
	t.Setenv("OPENPOST_DATABASE_DRIVER", "mysql")
	t.Setenv("OPENPOST_STORAGE_DRIVER", "gcs")

	cfg := Load()

	require.Equal(t, EditionSelfHost, cfg.Edition)
	require.Equal(t, DatabaseDriverSQLite, cfg.DatabaseDriver)
	require.Equal(t, StorageDriverLocal, cfg.StorageDriver)
}

func TestDatabaseDSNFallsBackToDatabasePathForPostgres(t *testing.T) {
	cfg := &Config{
		DatabaseDriver: DatabaseDriverPostgres,
		DatabasePath:   "postgres://legacy/path",
	}

	require.Equal(t, "postgres://legacy/path", cfg.DatabaseDSN())
}

// TestOauthRedirectFromFrontendPreferesExplicitEnv pins the contract
// that an operator-set env var (or its alias) wins over the derived
// default from FrontendURL. This matters for users who proxy their
// installation under a path or behind a hostname the binary can't see
// from OPENPOST_APP_URL.
func TestOauthRedirectFromFrontendPreferesExplicitEnv(t *testing.T) {
	t.Setenv("X_REDIRECT_URI", "https://proxy.example.com/api/v1/accounts/x/callback")
	t.Setenv("TWITTER_REDIRECT_URI", "")
	got := oauthRedirectFromFrontend("X_REDIRECT_URI", "TWITTER_REDIRECT_URI", "https://openpost.example.com", "/api/v1/accounts/x/callback")
	require.Equal(t, "https://proxy.example.com/api/v1/accounts/x/callback", got)
}

// TestOauthRedirectFromFrontendFallsBackToAlias covers the case where
// the primary env var isn't set but the legacy alias is.
func TestOauthRedirectFromFrontendFallsBackToAlias(t *testing.T) {
	t.Setenv("X_REDIRECT_URI", "")
	t.Setenv("TWITTER_REDIRECT_URI", "https://proxy.example.com/api/v1/accounts/x/callback")
	got := oauthRedirectFromFrontend("X_REDIRECT_URI", "TWITTER_REDIRECT_URI", "https://openpost.example.com", "/api/v1/accounts/x/callback")
	require.Equal(t, "https://proxy.example.com/api/v1/accounts/x/callback", got)
}

// TestOauthRedirectFromFrontendDerivesFromFrontendURL is the regression
// test for the operator footgun (P0.3): when nothing is set, the
// default is derived from FrontendURL rather than hardcoded to
// localhost:8080 (which would 404 in production).
func TestOauthRedirectFromFrontendDerivesFromFrontendURL(t *testing.T) {
	t.Setenv("X_REDIRECT_URI", "")
	t.Setenv("TWITTER_REDIRECT_URI", "")
	got := oauthRedirectFromFrontend("X_REDIRECT_URI", "TWITTER_REDIRECT_URI", "https://openpost.example.com", "/api/v1/accounts/x/callback")
	require.Equal(t, "https://openpost.example.com/api/v1/accounts/x/callback", got)
}

// TestOauthRedirectFromFrontendStripsTrailingSlash covers the common
// case where the operator sets OPENPOST_APP_URL with a trailing slash.
func TestOauthRedirectFromFrontendStripsTrailingSlash(t *testing.T) {
	t.Setenv("LINKEDIN_REDIRECT_URI", "")
	got := oauthRedirectFromFrontend("LINKEDIN_REDIRECT_URI", "", "https://openpost.example.com/", "/api/v1/accounts/linkedin/callback")
	require.Equal(t, "https://openpost.example.com/api/v1/accounts/linkedin/callback", got)
}

// TestOauthRedirectFromFrontendNoAlias covers the LinkedIn / Threads
// case where there is no legacy alias. Passing an empty alias should
// not cause a panic and should derive from FrontendURL.
func TestOauthRedirectFromFrontendNoAlias(t *testing.T) {
	t.Setenv("THREADS_REDIRECT_URI", "")
	got := oauthRedirectFromFrontend("THREADS_REDIRECT_URI", "", "https://openpost.example.com", "/api/v1/accounts/threads/callback")
	require.Equal(t, "https://openpost.example.com/api/v1/accounts/threads/callback", got)
}

// TestOauthRedirectFromFrontendEmptyFrontendDerivesPathOnly documents
// the (unusual) edge case where FrontendURL is empty. The result is
// still well-formed (a path-only URL), but the operator probably wants
// to set OPENPOST_APP_URL.
func TestOauthRedirectFromFrontendEmptyFrontendDerivesPathOnly(t *testing.T) {
	t.Setenv("X_REDIRECT_URI", "")
	t.Setenv("TWITTER_REDIRECT_URI", "")
	got := oauthRedirectFromFrontend("X_REDIRECT_URI", "TWITTER_REDIRECT_URI", "", "/api/v1/accounts/x/callback")
	require.Equal(t, "/api/v1/accounts/x/callback", got)
}

// TestWarnOnPlaceholderURLNoEnvSet documents the warning fires when
// neither OPENPOST_APP_URL nor its legacy alias is set, so the operator
// is on the binary's default and the loud-warn is the right move.
func TestWarnOnPlaceholderURLNoEnvSet(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "")
	t.Setenv("OPENPOST_FRONTEND_URL", "")
	// The function logs but does not return a value; we just verify it
	// does not panic when the env is empty.
	require.NotPanics(t, func() {
		warnOnPlaceholderURL(&Config{FrontendURL: "http://localhost:8080"})
	})
}

// TestWarnOnPlaceholderURLExplicitEnvSkipsWarn documents the contract
// that the warning is suppressed once the operator has set
// OPENPOST_APP_URL (or its alias). `devenv shell` and any real
// deployment must not see the warning.
func TestWarnOnPlaceholderURLExplicitEnvSkipsWarn(t *testing.T) {
	t.Setenv("OPENPOST_APP_URL", "https://openpost.example.com")
	t.Setenv("OPENPOST_FRONTEND_URL", "")
	require.NotPanics(t, func() {
		warnOnPlaceholderURL(&Config{FrontendURL: "https://openpost.example.com"})
	})
}
