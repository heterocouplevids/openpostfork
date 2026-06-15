// Package accountpicker turns the user-facing --accounts flag value
// into a list of social_account.id values that match the workspace.
//
// Supported forms:
//   - <id>                  raw account id
//   - x                     first/only x account in the workspace
//   - x:@username           x account whose AccountUsername is "username"
//   - linkedin              first/only linkedin account
//   - mastodon:server.example first/only mastodon account for that server
//
// On multiple matches the picker returns an error listing the
// candidates so the user can disambiguate.
package accountpicker

import (
	"fmt"
	"strings"

	"github.com/openpost/cli/internal/api"
)

// Resolve maps each selector to a social_account.id. The returned
// slice preserves the order of the input selectors.
func Resolve(workspaceID string, selectors []string, accounts []api.SocialAccount) ([]string, error) {
	if len(selectors) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(selectors))
	seen := map[string]struct{}{}
	for _, raw := range selectors {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		id, err := resolveOne(workspaceID, s, accounts)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func resolveOne(workspaceID, sel string, accounts []api.SocialAccount) (string, error) {
	var matches []api.SocialAccount
	for _, a := range accounts {
		if !a.IsActive {
			continue
		}
		if accountMatches(a, sel) {
			matches = append(matches, a)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no %s account matched in workspace %s", sel, workspaceID)
	case 1:
		return matches[0].ID, nil
	default:
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, formatAccount(m))
		}
		return "", fmt.Errorf("multiple matches for %q: %s. Disambiguate with %s or with the raw account id", sel, strings.Join(names, ", "), disambiguateHint(sel, matches))
	}
}

func accountMatches(a api.SocialAccount, sel string) bool {
	// exact id
	if a.ID == sel {
		return true
	}
	// platform:username form
	if idx := strings.Index(sel, ":"); idx > 0 {
		platform := sel[:idx]
		handle := strings.TrimPrefix(sel[idx+1:], "@")
		if a.Platform == platform && strings.EqualFold(a.AccountUsername, handle) {
			return true
		}
		return false
	}
	// bare platform
	if a.Platform == sel {
		return true
	}
	// mastodon:server.example  → match InstanceURL host
	if strings.HasPrefix(sel, "mastodon:") {
		if a.Platform != "mastodon" {
			return false
		}
		// Account.InstanceURL is the full URL; user passes the host
		// in the bare "mastodon:server" form. Compare on host.
		want := strings.TrimPrefix(sel, "mastodon:")
		return hostOf(a.InstanceURL) == want
	}
	return false
}

func hostOf(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if i := strings.Index(u, "/"); i > 0 {
		u = u[:i]
	}
	return u
}

func formatAccount(a api.SocialAccount) string {
	username := a.AccountUsername
	if username == "" {
		username = a.AccountID
	}
	if a.InstanceURL != "" {
		return fmt.Sprintf("%s:@%s@%s", a.Platform, username, hostOf(a.InstanceURL))
	}
	return fmt.Sprintf("%s:@%s", a.Platform, username)
}

func disambiguateHint(sel string, matches []api.SocialAccount) string {
	parts := make([]string, 0, len(matches))
	for _, m := range matches {
		parts = append(parts, fmt.Sprintf("--accounts %s", formatAccount(m)))
	}
	return strings.Join(parts, " or ")
}
