package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
