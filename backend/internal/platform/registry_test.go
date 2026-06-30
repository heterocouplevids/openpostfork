package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAdapterRegistryBuildsCanonicalAndAliasKeys(t *testing.T) {
	t.Parallel()

	adapters, entries, err := BuildAdapterRegistry([]AppConfig{
		{Provider: "bluesky"},
		{Provider: "x", ClientID: "x-client", ClientSecret: "x-secret", RedirectURI: "https://app.test/api/v1/accounts/x/callback"},
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
	require.Contains(t, adapters, "x")
	require.Contains(t, adapters, "mastodon:https://masto.pt")
	require.Contains(t, adapters, "mastodon:Personal")
	require.Same(t, adapters["mastodon:https://masto.pt"], adapters["mastodon:Personal"])
	require.Len(t, entries, 4)
}

func TestBuildAdapterRegistryRejectsUnsupportedOrIncompleteApps(t *testing.T) {
	t.Parallel()

	_, _, err := BuildAdapterRegistry([]AppConfig{{Provider: "youtube"}}, RegistryOptions{})
	require.ErrorContains(t, err, "unsupported provider app")

	_, _, err = BuildAdapterRegistry([]AppConfig{{Provider: "threads"}}, RegistryOptions{})
	require.ErrorContains(t, err, "threads provider app requires client_id")
}
