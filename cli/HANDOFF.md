# OpenPost CLI — Codex Handoff Brief

## Context

The OpenPost repo is a self-hosted social media scheduler: Go + Echo + Huma
backend, SvelteKit frontend, embedded in a single Go binary at
`backend/cmd/openpost/main.go`. The current auth model issues a 7-day JWT
through `Authorization: Bearer ...`. The web frontend already uses
`openapi-fetch` against generated OpenAPI types at
`frontend/src/lib/api/client.ts`.

The user (Rodrigo) wants a first-class CLI client for talking to a running
OpenPost instance over HTTPS. The plan is documented in
`~/.hermes/memories/MEMORY.md` (and in the long-form user message we got
before this brief). The decisions you must follow are below — they are
non-negotiable because they were already made deliberately by the user
and the parent agent:

- The CLI is a **standalone client**, not a subcommand of the server
  binary. It lives in `cli/` as a separate Go module
  (`github.com/openpost/cli`).
- The CLI must NOT import anything from `backend/internal/...`. It only
  talks to the running server over `/api/v1`. This is enforced by the
  fact that `cli/go.mod` is a separate module — the backend's
  `internal/...` paths simply aren't importable.
- Auth uses a new opaque API token (`op_cli_<prefix>_<secret>`) stored
  server-side as a sha256 hash. Web JWT auth is unchanged. Existing
  handlers still receive the same `user_id` from the request context —
  the only thing that changes is that the bearer authenticator tries
  JWT first, then API token.
- The CLI login flow is a device-code flow (RFC 8628-style) by default,
  served from the web UI. The CLI never handles the user's password,
  TOTP code, or passkey. CLI auth is also accessible via a `--with-token`
  flag reading from stdin (for CI / headless).
- The CLI token is stored client-side in the OS keyring
  (`github.com/zalando/go-keyring`), with a documented
  `--insecure-storage` fallback to a 0600 `credentials.json`.
- Social-provider OAuth (X, Mastodon, LinkedIn, Threads, Bluesky) stays
  server-side. `openpost account connect x` simply calls the existing
  `GET /api/v1/accounts/{platform}/auth-url` and opens the browser. We do
  NOT try to do provider OAuth in the terminal.

## What is already done (committed as 91d05de)

The parent agent laid down the file layout. The following files are
already in place — read them before writing more code so you understand
the shape that's been chosen:

- `cli/go.mod` — module `github.com/openpost/cli`, Go 1.25.0, no deps yet.
- `cli/cmd/openpost/main.go` — entrypoint that calls `commands.NewRoot(version)`.
- `cli/internal/commands/root.go` — Cobra root, global flags, calls
  `config.Load(...)` and stashes the result on the command context.
  Subcommands are added via `root.AddCommand(newAuthCmd())` etc. — write
  the factories with that signature.
- `cli/internal/commands/completion.go` — `openpost completion <shell>`.
- `cli/internal/commands/version.go` — `openpost version`.
- `cli/internal/config/config.go` — TOML config, `FlagOverrides`, `Runtime`
  struct, XDG paths, env precedence. **Use `config.FromCommand(cmd)`
  to pull the runtime out of the cobra command's context in subcommands.**
- `cli/internal/api/client.go` — typed HTTP client with method-per-endpoint
  surface: `Health`, `Me`, `ListWorkspaces`, `GetWorkspaceSettings`,
  `ListAccounts`, `ListMedia`, `UploadMedia`, `DeleteMedia`, `ListPosts`,
  `CreatePost`, `GetPost`, `DeletePost`, `CreateThread`, `ListAPITokens`,
  `CreateAPIToken`, `RevokeAPIToken`, `StartCLIAuth`, `PollCLIAuth`,
  `ListJobs`. **Add new typed methods here, do not call `do()` directly
  from subcommand code.**
- `cli/internal/auth/token_store.go` — `Store` interface, `KeyringStore`
  (default), `InsecureStore` (fallback), `NewStore(runtime)`, `HasToken`.
- `cli/internal/output/output.go` — `Printer` with `PrintJSON`,
  `Printf`, `Errorf`, `Table`. Use it everywhere.
- `cli/internal/schedule/parse.go` — `Parse(input, options)` with strict
  RFC3339, absolute layouts, natural-language, `now` / `draft` aliases.
  Returns `Result{Time, Source, Original, Warning}`. Empty time means
  "draft" (no schedule).
- `cli/internal/accountpicker/accountpicker.go` — `Resolve(workspaceID,
  selectors, accounts)` — turn `--accounts x,x:@main,mastodon:host`
  into account IDs.

## Codebase facts that affect your work

Read these files before you touch the backend:

- `backend/internal/api/middleware/auth.go` — the current JWT-only
  authenticator. You will **add a method, not replace the function**.
  The cleanest approach: introduce an `Authenticator` interface with
  `AuthenticateBearer(ctx, token) (*Principal, error)`, and have
  `AuthMiddleware` use a `*auth.CompositeService` that wraps the JWT
  service and the new `apitokens` service. JWT still works for the web.
- `backend/internal/services/auth/auth.go` — the JWT service. Add
  `apitokens` as a sibling service, then a `Composite` wrapper, and
  wire the wrapper into `main.go` where `authService` is currently
  passed to `middleware.NewAuthMiddleware` / `middleware.JWTMiddleware`.
  Several call sites use the JWT service directly (e.g. `authService.ValidateToken`);
  these are fine and do not need to change. Only the *middleware entry
  point* changes.
- `backend/internal/api/handlers/auth.go` — existing auth handlers. Do
  NOT add CLI handlers here; put them in a new
  `backend/internal/api/handlers/cli_auth.go` to keep the surface
  coherent. Same for `backend/internal/api/handlers/api_tokens.go`.
- `backend/internal/models/models.go` — add `APIToken` and
  `CLIAuthSession` structs (the exact shape is documented below).
  Register them in `backend/internal/database/database.go` `CreateSchema`.
- `backend/internal/database/migrations/` — add `008_api_tokens.sql`
  and `009_cli_auth_sessions.sql` (NOT `007_...` — `007_thread_drafts`
  is already taken). Use the same split-statements pattern as
  `007_thread_drafts.sql`. Migrations are auto-discovered by
  `migrations.go` (it embeds `*.sql` and runs in order). Add
  `api_tokens_test.go` and `cli_auth_sessions_test.go` for regression
  coverage following the `007_thread_drafts_test.go` style.
- `backend/cmd/openpost/main.go` — register the new handler methods
  on the Huma API. **Do NOT change the way the server boots, do NOT
  change ports, do NOT touch provider registration, do NOT change the
  Echo CORS config, do NOT change the worker, do NOT touch the
  shutdown sequence.** Just call the new handler methods in the same
  block where existing handlers are registered.
- `frontend/src/lib/api/client.ts` — frontend API client. The web
  Settings page for managing API tokens needs a route at
  `frontend/src/routes/settings/tokens/+page.svelte` (and probably
  `+page.ts` for data loading). Read an existing settings page to
  match the style.
- `frontend/src/routes/cli/authorize/+page.svelte` — the CLI device
  approval page. Reads `?user_code=XXXX` from the URL, shows the
  client name + machine + requested scopes + expiry, has Approve/Deny
  buttons that POST to `/api/v1/cli/auth/approve` /
  `/cli/auth/deny`. If the user is not logged in, the page redirects
  to `/login?redirect=/cli/authorize?...`. Use the existing Paraglide
  i18n strings (don't add new locales for this).

## Hard constraints

- **No SQLite access from CLI.** Even if it would be faster, even for
  read-only operations. The CLI is an external client.
- **No direct imports of `backend/internal/...` from `cli/`.** Different
  Go module, different dependency graph.
- **JWT auth keeps working unchanged.** The web frontend's
  `localStorage.getItem('token')` flow is not touched. New endpoints
  use the same `Authorization: Bearer ...` header; the middleware just
  tries two authenticators.
- **Social provider OAuth stays server-side.** No provider-specific
  OAuth client code in the CLI.
- **Token format is `op_cli_<8-hex-prefix>_<43+-char-secret>`**. The
  prefix is for display/lookup; the secret is the bearer. Store
  `sha256(prefix + ":" + secret)` server-side. **Never log the raw
  token anywhere — including errors, debug output, and tests.** Use
  `t.T.Helper()` and assert on the prefix, not the full string.
- **Token expiry default: 90 days.** Support explicit longer expiry
  on self-hosted instances via `--expires-in 0` (= never) in the API,
  but never make "never expires" the default in the web UI. The CLI
  `auth login` does not expose this flag in v1.
- **Token scopes v1**: a single scope `cli:full` that grants
  workspace/account/post/media/jobs read+write on every workspace the
  user has access to. Per-workspace fine-grained scoping is a later
  change. Store the scope as a string column, parse at validation time
  — do NOT introduce a `token_scopes` table for v1.
- **Huma/OpenAPI compat**: every new handler must be registered via
  `huma.Register` so the spec picks it up. The CLI does not generate
  a client from the spec for v1 — it uses the hand-written methods in
  `cli/internal/api/client.go` — but the spec must still be valid
  because the docs site and the web frontend read it.
- **Conventional Commits, conventional branches.** One logical unit
  per commit. Do NOT bundle backend + frontend + CLI into one commit
  — they should land as separate commits even when they belong to the
  same feature phase.
- **`devenv shell` for anything that needs `go`/`bun`/`prek` on the
  host.** The host PATH does not have these; the pre-commit hooks
  only work inside the devenv shell. Commit via
  `devenv shell /tmp/commit-N.sh` (see prior commits).
- **`go clean -testcache` before claiming backend tests pass.** The
  pre-commit `backend-go-test` hook does NOT pass `-count=1`, so
  cached test runs can lie.

## Phasing — commit this work in this order

### Phase 1 — Backend: API tokens

Files to create / modify:
- `backend/internal/models/models.go` — add `APIToken` and `CLIAuthSession`.
- `backend/internal/database/database.go` — register both in `CreateSchema`.
- `backend/internal/database/migrations/008_api_tokens.sql`
- `backend/internal/database/migrations/009_cli_auth_sessions.sql`
- `backend/internal/database/migrations/api_tokens_test.go` — covers
  schema, data migration, idempotency, FK cascade, prefix lookup, hash
  uniqueness.
- `backend/internal/database/migrations/cli_auth_sessions_test.go` —
  covers schema, idempotency, user_code uniqueness, status enum,
  expiry cleanup.
- `backend/internal/services/apitokens/service.go` — `GenerateToken`,
  `HashToken`, `ValidateToken`, `ListTokens`, `RevokeToken`,
  `TouchLastUsedAt`. Token format: `op_cli_<8-hex-prefix>_<32-byte-base64url>`.
  Storage: `token_prefix` = first 8 hex of sha256(secret), `token_hash`
  = sha256(secret). One constant-time lookup by `token_prefix` + verify
  hash on the candidate row.
- `backend/internal/services/apitokens/service_test.go`.
- `backend/internal/api/handlers/api_tokens.go` —
  `GET /api/v1/api-tokens` (list, never returns the raw token),
  `POST /api/v1/api-tokens` (mint — returns raw token ONCE),
  `DELETE /api/v1/api-tokens/{id}` (revoke).
- `backend/internal/api/middleware/auth.go` — add an
  `Authenticator` interface and a `CompositeService` that wraps
  `*auth.Service` and `*apitokens.Service`. New behavior: try JWT,
  fall back to API token. Old JWT-only behavior is preserved
  byte-for-byte when the token is a JWT.
- `backend/cmd/openpost/main.go` — instantiate the new services,
  register the new handler.

Acceptance: `curl -H "Authorization: Bearer op_cli_..." http://localhost:8080/api/v1/auth/me`
returns the same user info as the JWT path. **Commit and stop.**

### Phase 2 — Backend: CLI device-flow endpoints + web approval page

Files to create:
- `backend/internal/api/handlers/cli_auth.go`:
  - `POST /api/v1/cli/auth/start` — body:
    `{ client_name, client_version, client_os, requested_scopes }`,
    response: `{ device_code, user_code, verification_url, expires_in, interval }`.
    Device code and user code are sha256-hashed at rest; only the
    plain user_code is returned to the client (it goes in the URL
    the user types). Verification URL is
    `<instance>/cli/authorize?user_code=<plain>`.
  - `POST /api/v1/cli/auth/poll` — body: `{ device_code }`, response:
    `authorization_pending` (default), `approved` (carries the new
    raw token), `access_denied`, `expired_token`. Honours the
    `interval` field on retry; never polls faster than 1 req/sec.
  - `POST /api/v1/cli/auth/approve` — auth required, body:
    `{ device_code, scopes (override), name (optional token name) }`.
    Mints an `APIToken` with the requested scopes, marks the session
    approved, returns success.
  - `POST /api/v1/cli/auth/deny` — auth required, body:
    `{ device_code }`. Marks the session denied.
- `backend/internal/api/handlers/cli_auth_test.go` — covers the
  full happy path, denial, expiry, polling interval, rate limit,
  and that the user_code is the only thing the CLI gets.
- `frontend/src/routes/cli/authorize/+page.svelte` — read
  `user_code` from query, show client info, Approve/Deny buttons.
  Redirect to /login if not authenticated. Use existing Paraglide
  strings.

Also wire the new `cli/auth/*` and `api-tokens` routes into
`backend/cmd/openpost/main.go`. **Commit and stop.**

### Phase 3 — CLI: subcommand skeleton, auth, completion

Files to create:
- `cli/internal/commands/auth.go` — `openpost auth login|status|logout|token list|token revoke <id>`.
  - `login <instance>` — defaults to browser flow. `--device` to
    print user_code + URL and poll (no browser). `--with-token` to
    read raw token from stdin. `--no-browser` to skip the auto-open.
  - `status` — shows profile, instance, token store name, keyring
    entry exists y/n, expiry.
  - `logout` — deletes token from active store.
  - `token list` — calls `GET /api/v1/api-tokens`, prints a table.
  - `token revoke <id>` — `DELETE /api/v1/api-tokens/{id}`.
- `cli/internal/commands/instance.go` — `instance add|list|use|remove`.
  `instance add <name> <url>` writes a profile with that name into
  `config.toml`. `instance use <name>` sets `current_profile`.
- `cli/internal/commands/workspace.go` — `workspace list|use|create`.
  `workspace use <name|id>` updates the active profile's
  `workspace_id` and `workspace_name`.
- `cli/internal/commands/completion.go` — already exists; nothing to do.
- `cli/go.mod` — add deps via `go mod tidy`:
  `github.com/spf13/cobra`,
  `github.com/zalando/go-keyring`,
  `github.com/markusmobius/go-dateparser`,
  `github.com/BurntSushi/toml`.
- `cli/README.md` — install / build / quickstart.

Verification: `go build ./cli/cmd/openpost` succeeds; `openpost --help`,
`openpost auth login --help`, `openpost auth status`, `openpost
instance --help` all render and exit 0. `openpost completion bash` emits
a valid script. **Commit and stop.**

### Phase 4 — CLI: workspace / account / media

Files to create:
- `cli/internal/commands/account.go` — `account list`, `account
  connect <platform> [--server <name>]` (calls
  `GET /api/v1/accounts/{platform}/auth-url`, opens browser, polls the
  callback? No — server-side OAuth means we just open the URL and the
  user comes back. Print a message about the active workspace.),
  `account disconnect <id>`.
- `cli/internal/commands/media.go` — `media upload <file> [--alt
  "..."]`, `media list [--limit N]`, `media delete <id>`.

Verification: against a real local OpenPost instance, `openpost
account list` returns the user's accounts and `openpost media upload
./test.png --alt "..."` works. **Commit and stop.**

### Phase 5 — CLI: post / thread / jobs

Files to create:
- `cli/internal/commands/post.go` — `post create|list|view|update|delete|publish-now`.
  - `create` flags: `--content` (or `--file`), `--accounts`,
    `--schedule`, `--media <id|path>`, `--alt`, `--workspace`,
    `--random-delay`. Parses schedule via `internal/schedule`. If
    `--media` points to an existing file, uploads first. Prints
    confirmation in human form or JSON.
  - `list` flags: `--workspace`, `--status`, `--date YYYY-MM-DD`,
    `--limit`.
  - `view <id>`, `update <id> --content ... --schedule ...`,
    `delete <id>`, `publish-now <id>`.
- `cli/internal/commands/thread.go` — `thread create <markdown-file>`.
  Parse the markdown (frontmatter: `workspace`, `accounts`,
  `schedule`, `random_delay`; body split by `---` on a line by itself).
  Build a `CreateThreadInput`. `POST /api/v1/posts/thread`.
- `cli/internal/commands/jobs.go` — `jobs list`, `jobs retry <id>`.
  (Retry is a `POST /api/v1/jobs/{id}/retry`; check the existing
  routes for the exact shape and the JWT-or-token middleware will
  accept CLI tokens.)
- `cli/internal/commands/account.go` is updated with account
  resolution by alias for `--accounts`.

Verification: `openpost post create --content "Hello" --accounts x
--schedule "tomorrow 2pm" --workspace personal` creates a scheduled
post. JSON output: `--json`. **Commit and stop.**

### Phase 6 — Release, docs, polish

Files to create / modify:
- `.github/workflows/release.yml` — add a `build-cli` job that
  cross-compiles for linux/darwin/windows × amd64/arm64 and uploads
  `openpost-cli-${os}-${arch}` artifacts alongside the server
  binaries. Use the same devenv shell pattern as `build-binaries`.
  Do NOT touch the existing server / docker / android jobs.
- `scripts/install-cli.sh` — `curl ... | sh` style installer that
  detects OS/arch, downloads the matching release binary, verifies
  its sha256 against the GitHub release checksums, and installs to
  `/usr/local/bin/openpost` (or `$HOME/.local/bin` if not root).
- `docs-site/.vitepress/config.ts` (or equivalent sidebar config) —
  add a `CLI` section with sub-pages: `installation.md`,
  `authentication.md`, `posting.md`, `scheduling.md`, `automation.md`.
- `docs-site/cli/installation.md` — quickstart: install, login, post.
- `docs-site/cli/authentication.md` — device flow, --with-token,
  env vars, profile semantics.
- `docs-site/cli/posting.md` — `openpost post create` examples for
  single posts, threads, with media, scheduled.
- `docs-site/cli/scheduling.md` — natural-language examples, timezone
  precedence, DST corner cases.
- `docs-site/cli/automation.md` — CI examples, OPENPOST_TOKEN
  patterns, profile per environment.
- `CHANGELOG.md` — Unreleased section entries per phase.

Verification: `devenv shell go build ./cli/cmd/openpost` succeeds,
`./openpost --version` prints the version, `devenv shell go test ./...`
in the backend passes (with `go clean -testcache` first). **Commit
and stop.**

## Things you must NOT do

- Do NOT add a TUI / interactive mode beyond the existing `Printer`
  patterns. We agreed not to. `huh` / `bubbletea` / `lipgloss` are
  out of scope for v1.
- Do NOT bundle the CLI into the server binary. The user explicitly
  chose the separate-module path. Don't second-guess.
- Do NOT touch the existing JWT auth code in
  `backend/internal/services/auth/auth.go`. Wrap it.
- Do NOT change the Huma `OpenAPI` mount at `/openapi.json`. Just
  add handlers; the spec is built dynamically.
- Do NOT add new env vars to `.env.example` for CLI behaviour — the
  CLI reads from its own `~/.config/openpost/config.toml` and
  `OPENPOST_*` env vars consumed by the CLI binary, not the server.
- Do NOT generate an OpenAPI-typed Go client in v1. Hand-written
  methods in `cli/internal/api/client.go` only.
- Do NOT do provider-specific OAuth in the CLI.
- Do NOT store raw tokens in `~/.config/openpost/config.toml`. The
  config file is non-secret metadata. Tokens live in the keyring (or
  the explicit `credentials.json` fallback file).

## Time budget

This is a multi-phase build, but each phase is small. Aim for one
commit per phase plus any obvious fixups. If you hit the model's
iteration cap mid-phase, stop cleanly (commit what compiles, leave a
`# TODO(phase-N): ...` marker), and report which phases you finished
and which you stubbed. The parent agent will review your diff, finish
the gaps, and commit.

## What "done" looks like for the parent

After your run, the parent will:
1. `git log --oneline feature/cli ^main` to see your commits.
2. Re-read every file you modified (your subagent self-reports can lie).
3. Re-run the gate suite in `devenv shell`:
   - `cd backend && go clean -testcache && go test ./...`
   - `cd cli && go build ./... && go test ./...`
   - `cd frontend && bun run typecheck && bun test`
4. Fix any rough edges, add missing CHANGELOG entries, push the
   branch, and write the final user-facing summary.

Make the parent's job easy: keep diffs narrow, keep tests in place,
keep conventional-commit messages honest, don't refactor files outside
your scope.
