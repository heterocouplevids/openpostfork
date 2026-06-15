# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Added
- Scaffold for a new `cli/` Go module (`github.com/openpost/cli`) — a standalone HTTP client for a running OpenPost instance. Includes the config layer (XDG config dir, profile precedence, flag > env > file), the OS keyring token store with an explicit --insecure-storage fallback, a typed API client, a JSON/table output printer, an account-picker that resolves `--accounts x,x:@main,mastodon:server.example` to social_account IDs, a schedule parser that handles RFC3339 / absolute layouts / natural-language ("tomorrow 2pm", "in 3 hours", "next monday 9am") / `now` / `draft`, and Cobra-based `root` and `completion` commands. The CLI does not embed the server, does not open SQLite, and does not import `backend/internal/...`.
- First-class API tokens for the CLI and other long-lived automation clients. `api_tokens` table stores sha256-hashed opaque tokens with the format `op_cli_<8-hex-prefix>_<base64url-secret>`; the JWT web path is unchanged. New `Authenticator` interface and `CompositeService` (JWT → API token fallback) wrap the existing `*auth.Service`. Huma handlers now accept the interface; the Echo `JWTMiddleware` is preserved. `GET /api/v1/api-tokens`, `POST /api/v1/api-tokens`, and `DELETE /api/v1/api-tokens/{id}` let users manage tokens from web/CLI; the raw token is returned exactly once on create.
- `cli_auth_sessions` table for the device-flow authorization flow that will land in the next phase (RFC 8628-style). Both device code and user code are stored as sha256 hashes; only the plain user_code ever leaves the server. Migration 008 and 009 are idempotent and auto-applied on startup, matching the 007 pattern.
- CLI device-flow authorization endpoints under `/api/v1/cli/auth/`: `POST /start` (opens a session, returns device_code + user_code + verification_url, rate-limited per client IP), `POST /poll` (1 req/s minimum, slow-down + retry-after), `GET /session?user_code=…` (the web approval page), `POST /approve` (mints an `APIToken` via the existing apitokens service, raw token returned once), `POST /deny`. New `internal/services/cli_auth` package wraps the session lifecycle and `CleanupExpired`. The CLI never handles the user's password, TOTP, or passkey — the user approves in the web UI.
- `/cli/authorize` web page that gates on auth, fetches the pending session, and shows the client identity, requested scopes, and Approve/Deny buttons. Same-origin `?redirect=` support on `/login` so the page round-trips through login when needed. New shadcn-style `Badge` primitive used by the requested-scopes chips.
- CLI skeleton: `openpost auth login|status|logout`, `openpost auth token list|revoke`, `openpost instance add|list|use|remove`, `openpost workspace list|use|create`. Login flows through the device-flow endpoints (browser-open by default, `--device` for SSH/headless, `--with-token` for stdin paste, `--insecure-storage` to opt out of the OS keyring). Config lives in `~/.config/openpost/config.toml` (XDG-aware); tokens live in the OS keyring by default. `--json`, `--quiet`, `--yes`, `--no-color`, and shell completions for `bash`, `zsh`, `fish`, and `powershell` are wired in. The CLI does not embed the server, does not open SQLite, and does not import `backend/internal/...`.
- CLI account and media commands: `openpost account list|disconnect` and `openpost media upload|list`. The `--accounts` picker resolves platform aliases (`x`, `linkedin`, `x:@username`) and account IDs against the workspace's social accounts, with a friendly disambiguation hint when a platform has multiple accounts. Accountpicker has table-driven unit tests covering the empty / single / multiple-match paths.
- CLI posting commands: `openpost post create|list|view|update|delete` and `openpost thread create <file>`. `--schedule` accepts RFC3339, `now`, `draft`, or natural language (`tomorrow 2pm`, `in 3 hours`, `next monday 9am`) and resolves against the workspace's timezone with a friendly confirmation prompt for natural-language inputs. `--accounts` resolves platform aliases via the picker. `--media` accepts existing media IDs or local file paths (uploaded first). Thread files use front matter for metadata and `---` separators between posts; the splitter has table-driven tests for front-matter, embedded dashes, empty segments, and mixed CRLF/LF. `openpost jobs list` surfaces the server's job queue.
- Release artifacts and documentation for the CLI: GitHub releases now build `openpost-cli-*` binaries for Linux, macOS, and Windows alongside the unchanged `openpost-server-*` artifacts; `scripts/install-cli.sh` installs the latest release binary with `curl | sh`; and new CLI docs cover installation, authentication, posting, and automation.
- Pre-commit hooks for the `cli/` Go module: `cli-gofmt`, `cli-golangci-lint`, and `cli-go-test` mirror the existing `backend/` hooks via the same `devenv`-generated `pre-commit-config.json` and run only for changes under `cli/`. The CLI's gofmt and golangci-lint were not previously gated at commit time; they are now.

### Changed
- Moved thread draft state out of `posts.content` (where it lived as a `__openpost_thread__:` JSON blob) into a dedicated `thread_drafts` table. The composer now sends the encoded draft as a typed `thread_draft` field on the create/update POST/PATCH and reads it back from the same field on get. The blob-in-content path is preserved as a fallback for data that was saved before the migration. Migration 007 is idempotent and runs automatically on startup.
- Replaced the `WHERE payload LIKE '%<uuid>%'` job-cancellation query in `posts.go` with a `type = 'publish_post' AND json_extract(payload, '$.post_id') = ?` match, so cancelling one post's jobs can no longer accidentally cancel other jobs (e.g. `media_cleanup`, `refresh_token`) whose payload happened to contain the post ID as a substring. Added a regression test in `posts_cancellation_test.go`.
- Made OAuth callback redirects absolute: the `Location` header on error and success paths now uses the configured `OPENPOST_APP_URL` as the base, so the redirect works correctly behind subpath reverse proxies and non-root mounts.
- Aligned the Go config's `*_REDIRECT_URI` defaults with `.env.example`: when an env var is unset the value is now derived from `OPENPOST_APP_URL` (with `urn:ietf:wg:oauth:2.0:oob` for Mastodon, matching the documented example).
- The Go binary now panics loudly at startup with a clear message if the embedded `index.html` is missing or empty. Previously a build that skipped the frontend step would silently serve a blank HTML page with HTTP 200.

### Removed
- Deleted the dead `frontend/messages/es.json` stub. Spanish was listed as a supported language in the docs and the ROADMAP, but the locale wasn't registered in Paraglide and the file only contained a single placeholder key. Both `frontend/README.md` and `ROADMAP.md` now reflect that Spanish is not yet shipped.
- Dropped `openpost account connect <platform>` from the CLI. Account connection is web-UI-only: provider credentials live on the server, the OAuth/Bluesky-app-password dance is server-side, and the CLI's only account-management surface is `list` and `disconnect`. The `account` cobra group has a `Long:` description pointing at `<instance>/accounts`, and `account list` against an empty workspace prints the URL to the web UI so the path stays discoverable. Unit tests cover the URL-construction and empty-state helpers.

### Fixed
- CLI list/single-resource endpoints (`ListAccounts`, `ListMedia`, `ListPosts`, `ListJobs`, `GetWorkspaceSettings`, `CreatePost`, `GetPost`, `CreateAPIToken`) used to decode Huma responses into a `struct{ Body T }` envelope. Huma v2 flattens the `Body` field on the wire, so the decode failed with `cannot unmarshal array into Go value of type struct { Body … }` and the CLI silently lost media data on `media list` (decoding `null` into a nil slice, then rendering "no media uploaded"). All endpoints now decode the flat wire format directly. 8 new `httptest`-backed regression tests in `cli/internal/api/client_test.go` lock the format for the next refactor.
- Legacy Echo media routes (`/api/v1/media/upload`, `/api/v1/media/batch-upload`, `/api/v1/media/metadata`) only accepted JWT web sessions because they wired `middleware.JWTMiddleware(h.auth)`. CLI users got a 401 (`invalid or expired token`) on every upload. New `middleware.BearerMiddleware(Authenticator)` is the Echo-shaped counterpart of `AuthMiddleware` and accepts both JWT and `op_cli_…` tokens via the unified `CompositeService`. The three legacy routes now use it. The bare `"Bearer"` literal was lifted to a `bearerPrefix` const to satisfy `goconst` across all three middleware implementations. 4 new `httptest` tests in `backend/internal/api/middleware/auth_test.go` cover success, missing header, malformed header, and rejected-token paths.

### Changed
- Expanded the README launch messaging around the Typefully-like workflow, target users, support snapshot, and current limitations.
- Filled in the thin operator docs with a more complete quickstart, single-binary install guide, backup and restore process, provider support matrix, and stronger security guidance.

## [1.0.9] - 2026-05-16

### Fixed
- Corrected Bluesky video service auth to use the user's PDS DID from the access JWT audience instead of assuming `bsky.social`.
- Corrected LinkedIn video status polling to percent-encode video URNs as Rest.li path variables.

## [1.0.8] - 2026-05-16

### Changed
- Media library deletion now allows media that is unused or only attached to already published posts, while still blocking media needed by draft, scheduled, publishing, or failed posts.

### Added
- Added a media library download action for saved images and videos.

## [1.0.7] - 2026-05-16

### Fixed
- Corrected Bluesky video service auth to use the documented GET query endpoint, parse wrapped video job responses, and poll video jobs with the service token.
- Prevented LinkedIn video posts from sending image-only media overrides and waited for finalized videos to become available before creating the post.
- Allowed dropdown sub-menus to overflow the quick-settings menu surface so the language picker is not clipped in production builds.

## [1.0.6] - 2026-05-14

### Changed
- Clarified the README and docs to reflect the actual provider-by-provider video implementation state instead of treating video support as universally absent.

### Fixed
- Corrected the launch TODO and public docs after auditing the current X, Mastodon, Bluesky, LinkedIn, and Threads video code paths.
- Reduced repeated backend string literals called out by `golangci-lint` `goconst` checks so local Go linting passes again.
- Added a real Bluesky video embed path, MIME-aware Threads media publishing, and LinkedIn video upload finalization with required file sizes.
- Updated composer and social previews to render attached videos as videos and warn about provider-specific media limitations.

## [1.0.5] - 2026-05-10

### Changed
- Refactored composer preview rendering so desktop and mobile previews share the same derived Svelte state model.
- Extended account-specific post variants to track media attachments independently from the synced post media.

### Fixed
- Fixed stale composer preview and textarea sizing when switching between synced and account-specific social media variants.
- Prevented media that is only attached to account-specific variants from being deleted as unused media.

## [1.0.4] - 2026-05-09

### Added
- Documentation page explaining why to self-host OpenPost, plus clearer provider/platform limitations coverage.
- Capacitor app asset generation and refreshed Android launcher/splash assets derived from the project brand icon.
- PWA manifest configuration for the frontend build.

### Changed
- Refreshed launch messaging across the README and docs site around the self-hosted Buffer/Hootsuite positioning, target users, and current product limitations.
- Android release builds now use the consolidated `build:capacitor` flow so frontend build, Capacitor sync, and mobile asset generation stay in one path.
- Asset sync now prepares the frontend logo source used by Capacitor asset generation.

### Fixed
- Stopped tracking the repository root `TODO.md` while ignoring the local file, so personal launch notes can remain in the working directory without showing up in git.
- Corrected Bluesky token expiry handling by deriving expiry times from the JWT on login and refresh, which keeps automatic refresh jobs scheduled correctly instead of relying on a hardcoded login window or stale timestamps.

## [1.0.3] - 2026-05-04

### Fixed
- Restored authenticated media rendering in the frontend by allowing media image requests to authorize with the current JWT and updating UI image URLs to include that access token.

## [1.0.2] - 2026-05-04

### Fixed
- Restored Mastodon OAuth validation and callback state handling so missing `server_name` requests fail cleanly and browser redirects can complete without requiring the callback query to repeat the server selection.
- Corrected workspace-scoped job listing to apply visibility filtering before `limit`, so non-admin users get full pages of jobs from accessible workspaces.
- Signed Threads media URLs now target the app media endpoint by media ID instead of the underlying file basename.

## [1.0.1] - 2026-05-03

### Fixed
- Docker release builds now copy the repo `scripts/` directory so the frontend asset-sync step works in GitHub Actions and container releases complete successfully.

## [1.0.0] - 2026-05-03

### Added
- Account-level MFA with QR-based TOTP enrollment, passkey registration, and step-up login verification, plus settings UI for managing both methods.
- VitePress documentation site scaffold under `docs-site/`, including landing page, sidebar/navigation config, OpenPost-themed styling, and first-pass operator/contributor docs.
- Shared asset sync pipeline that copies canonical repo assets into frontend and docs public directories.
- GitHub Pages workflow for building and deploying the docs site.
- Token refresh job scheduling plus backend tests covering queued refresh execution and provider-specific refresh credentials.
- Dedicated account-connection success callback page for returning OAuth users to `/accounts`.
- Workspace migration scaffold for configurable draft gap minutes.
- Workspace setting for `draft_gap_minutes`, used by suggested queue times when a day's configured schedule slots are already occupied.

### Changed
- Settings now include account-security controls, while login can require a second factor when TOTP or passkeys are enabled.
- Optimized GitHub Actions CI by priming a shared Nix store cache before lint/test jobs, caching Go/lint/Bun dependencies, skipping unaffected backend/frontend jobs, and moving Go race tests off pull request runs.
- README reduced to a shorter front door that points detailed setup and operations content at the docs site.
- Docs site base-path handling now defaults to `/` for custom-domain hosting, with `OPENPOST_DOCS_BASE` available as an explicit override for repository-path deployments like `/openpost/`.
- README docs links now point at the custom docs domain `https://op.rgo.pt`.
- Docs now include a Nix module deployment page backed by a build-time sync of the production module from `rodrgds/nix-config`.
- Token refresh handling now declares platform capabilities explicitly, retries publish attempts on any supported expired account, and routes OAuth success redirects through the new callback screen.
- Workspace settings no longer auto-overwrite shared timezone and week-start values from the first browser locale that opens a workspace.
- Posting schedule settings now use a local-time weekly grid with per-day toggles and row-based time management instead of a flat UTC slot list.
- Suggested posting times now consider already scheduled posts and fall back to the configured minimum draft gap when a day has no unused schedule slots left.
- Weekly posting schedules now preserve the configured workspace-local time across DST changes instead of drifting by the current UTC offset.

### Fixed
- Mastodon accounts now persist their configured `instance_url` as the canonical provider key, avoiding publish/token-refresh mismatches after OAuth connection.
- The default Mastodon callback URI now matches the documented backend callback endpoint on `localhost:8080`.
- Mastodon server listings now avoid duplicate entries when adapters are registered with both UI labels and canonical instance-url keys.

## [0.4.4] - 2026-04-19

Changes since `v0.4.3`.

### Added
- X OAuth request store handler for temporary request-state persistence.
- Frontend OpenAPI snapshot and generated API TypeScript declarations tracked in-repo for CI consistency.
- Placeholder file in embedded web public directory to keep `go:embed` stable in clean checkouts.

### Changed
- X OAuth handler and platform integration flow refinements.
- Backend model and database updates supporting the latest auth/request-state behavior.
- Frontend pre-commit/devenv validation flow now runs deterministic generation/check steps for i18n and OpenAPI types.
- Frontend dashboard and media routes fixed strict TypeScript nullability errors found in CI.
- Frontend ignore/format rules adjusted to avoid generated-file drift during hooks.

## [0.4.3] - 2026-04-19

Changes since `v0.4.2`.

### Added
- Prompt management backend API (`/prompts`, `/prompts/random`, `/prompts/categories`, create/delete custom prompts).
- Built-in prompt catalog seeding and prompt category support.
- Posting schedule backend API (`/posting-schedules` list/create/update/delete).
- Prompt browsing UI at `/prompts` with category filtering, random prompt selection, and custom prompt creation.
- Compose flow integration for using prompts directly in new posts.
- Settings UI support for posting schedule slot management.

### Changed
- Post handler logic expanded for improved post management and scheduling workflows.
- Media handler behavior updated for media lifecycle and cleanup flow alignment.
- Authentication middleware updated for request auth handling refinements.
- Database/model layer updated with new scheduling and prompt entities.
- Queue worker updated to process scheduling-related jobs.
- Frontend layout refactors for improved page consistency (`PageContainer`, `EmptyState`, sidebar and dashboard updates).
- Favicon assets refreshed.

### Project And Docs
- Frontend page layout refactor and onboarding/UI refinements.
- Added AI agent skill definitions and repo agent guideline updates.
- Added/expanded roadmap and planning documentation updates.

### Commit Summary Since v0.4.2
- `681e3ab` refactor(frontend): unify page layouts with PageContainer and EmptyState components
- `bde9cc1` docs(agents): add conventional commits and branches requirement
- `a6f60ee` feat(frontend): add onboarding page and UI refinements
- `a53ef22` feat(agents): add AI agent skill definitions
- `7289963` feat: implement Phase 3 - Media Management & Cleanup
- `87a1901` feat: implement Phase 2 - Platform Customization & Social Media Sets
- `80c302c` feat: enhance post management features
- `cb8a110` feat: add comprehensive roadmap for OpenPost features and priorities
