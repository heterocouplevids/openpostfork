package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAdapterRegistryBuildsCanonicalAndAliasKeys(t *testing.T) {
	t.Parallel()

	adapters, entries, err := BuildAdapterRegistry([]AppConfig{
		{Provider: "bluesky"},
		{Provider: "facebook", ClientID: "facebook-client", ClientSecret: "facebook-secret", RedirectURI: "https://app.test/api/v1/accounts/facebook/callback"},
		{Provider: "instagram", ClientID: "instagram-client", ClientSecret: "instagram-secret", RedirectURI: "https://app.test/api/v1/accounts/instagram/callback"},
		{Provider: "x", ClientID: "x-client", ClientSecret: "x-secret", RedirectURI: "https://app.test/api/v1/accounts/x/callback"},
		{Provider: "tiktok", ClientID: "tiktok-client", ClientSecret: "tiktok-secret", RedirectURI: "https://app.test/api/v1/accounts/tiktok/callback"},
		{
			Provider:     "mastodon",
			Name:         "Personal",
			ClientID:     "masto-client",
			ClientSecret: "masto-secret",
			RedirectURI:  "urn:ietf:wg:oauth:2.0:oob",
			InstanceURL:  "https://masto.pt/",
		},
	}, RegistryOptions{})

	require.NoError(t, err)
	require.Contains(t, adapters, "bluesky")
	require.Contains(t, adapters, "facebook")
	require.Contains(t, adapters, "instagram")
	require.Contains(t, adapters, "x")
	require.Contains(t, adapters, "tiktok")
	require.Contains(t, adapters, "mastodon:https://masto.pt")
	require.Contains(t, adapters, "mastodon:Personal")
	require.Same(t, adapters["mastodon:https://masto.pt"], adapters["mastodon:Personal"])
	require.Len(t, entries, 7)
}

func TestBuildAdapterRegistryRejectsUnsupportedOrIncompleteApps(t *testing.T) {
	t.Parallel()

	_, _, err := BuildAdapterRegistry([]AppConfig{{Provider: "youtube"}}, RegistryOptions{})
	require.ErrorContains(t, err, "unsupported provider app")

	_, _, err = BuildAdapterRegistry([]AppConfig{{Provider: "threads"}}, RegistryOptions{})
	require.ErrorContains(t, err, "threads provider app requires client_id")

	_, _, err = BuildAdapterRegistry([]AppConfig{{Provider: "facebook"}}, RegistryOptions{})
	require.ErrorContains(t, err, "facebook provider app requires client_id")

	_, _, err = BuildAdapterRegistry([]AppConfig{{Provider: "instagram"}}, RegistryOptions{})
	require.ErrorContains(t, err, "instagram provider app requires client_id")

	_, _, err = BuildAdapterRegistry([]AppConfig{{Provider: "tiktok"}}, RegistryOptions{})
	require.ErrorContains(t, err, "tiktok provider app requires client_id")
}
