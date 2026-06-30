package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/openpost/backend/internal/platform"
	"github.com/openpost/backend/internal/services/mastodonapps"
	"github.com/stretchr/testify/require"
)

type providerAvailabilityAdapter struct {
	instanceURL string
}

func (a providerAvailabilityAdapter) GenerateAuthURL(string) (string, map[string]string) {
	return "", nil
}

func (a providerAvailabilityAdapter) ExchangeCode(context.Context, string, map[string]string) (*platform.TokenResult, error) {
	return nil, nil
}

func (a providerAvailabilityAdapter) RefreshCapability() platform.RefreshCapability {
	return platform.RefreshCapability{}
}

func (a providerAvailabilityAdapter) RefreshToken(context.Context, platform.RefreshTokenInput) (*platform.TokenResult, error) {
	return nil, nil
}

func (a providerAvailabilityAdapter) GetProfile(context.Context, string) (*platform.UserProfile, error) {
	return nil, nil
}

func (a providerAvailabilityAdapter) UploadMedia(context.Context, string, string, string, io.Reader) (string, error) {
	return "", nil
}

func (a providerAvailabilityAdapter) Publish(context.Context, string, string, *platform.PublishRequest) (string, error) {
	return "", nil
}

func (a providerAvailabilityAdapter) InstanceURL() string {
	return a.instanceURL
}

func TestListProvidersReportsConfiguredProviders(t *testing.T) {
	t.Parallel()

	e := echo.New()
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), huma.DefaultConfig("Test", "1.0.0"))
	mastodonAdapter := providerAvailabilityAdapter{instanceURL: "https://masto.pt"}
	handler := &OAuthHandler{
		auth: testAuthenticator{},
		providers: map[string]platform.Adapter{
			"bluesky":                   providerAvailabilityAdapter{},
			"x":                         providerAvailabilityAdapter{},
			"mastodon:https://masto.pt": mastodonAdapter,
			"mastodon:Personal":         mastodonAdapter,
		},
	}
	handler.ListProviders(api)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/accounts/providers", nil)
	req.Header.Set("Authorization", "Bearer web-token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out []ProviderInfo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 5)
	require.Equal(t, ProviderInfo{Platform: "bluesky", DisplayName: "Bluesky", AuthMode: "app_password", Configured: true}, out[0])
	require.Equal(t, ProviderInfo{Platform: "x", DisplayName: "X (Twitter)", AuthMode: "oauth", Configured: true}, out[1])
	require.Equal(t, ProviderInfo{Platform: "mastodon", DisplayName: "Mastodon", AuthMode: "oauth_oob", Configured: true, Name: "Personal", InstanceURL: "https://masto.pt"}, out[2])
	require.Equal(t, ProviderInfo{Platform: "linkedin", DisplayName: "LinkedIn", AuthMode: "oauth", Configured: false}, out[3])
	require.Equal(t, ProviderInfo{Platform: "threads", DisplayName: "Threads", AuthMode: "oauth", Configured: false}, out[4])
}

func TestListProvidersIncludesUnavailableMastodonPlaceholder(t *testing.T) {
	t.Parallel()

	handler := &OAuthHandler{providers: map[string]platform.Adapter{}}
	out := handler.providerAvailability()

	require.Len(t, out, 5)
	require.Equal(t, ProviderInfo{Platform: "mastodon", DisplayName: "Mastodon", AuthMode: "oauth_oob", Configured: false}, out[2])
}

func TestListProvidersReportsDynamicMastodonAvailable(t *testing.T) {
	t.Parallel()

	handler := &OAuthHandler{
		providers:    map[string]platform.Adapter{},
		mastodonApps: mastodonapps.NewService(nil, nil, mastodonapps.Options{}),
	}
	out := handler.providerAvailability()

	require.Len(t, out, 5)
	require.Equal(t, ProviderInfo{
		Platform:    "mastodon",
		DisplayName: "Mastodon",
		AuthMode:    "oauth_oob",
		Configured:  true,
		Name:        "Custom instance",
	}, out[2])
}
