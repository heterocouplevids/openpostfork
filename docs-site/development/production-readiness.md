# Production Readiness Plan

This is the implementation map for turning OpenPost into a production-ready self-hosted product plus OpenPost Cloud. The public repo should keep the shared product core, while private infrastructure stays in the deployment/ops layer.

## Product Direction

- Keep **OpenPost** as the product name.
- Use **OpenPost Cloud** for the official hosted service.
- Keep `openpost.social` as the marketing site, `docs.openpost.social` as the docs site, and `app.openpost.social` as the app.
- Position the product as: write one idea, adapt it into platform-native renditions, and publish intentionally.
- Keep self-hosting credible: no artificial self-hosted feature crippling.

## Architecture Principles

- Keep one shared product core in this repo.
- Keep secrets, production deployment config, provider credentials, monitoring, and private admin scripts outside this repo.
- Prefer interface-backed hosted primitives so self-hosted and cloud use the same API paths.
- Keep SQLite and local media as first-class self-hosted defaults.
- Add Postgres and S3/R2 as cloud-ready drivers, not replacements.
- Treat billing as entitlements and usage limits, not provider-specific checks scattered through handlers.
- Use background jobs for provider publishing, media processing, token refresh, and other restart-sensitive work.

## Milestones

### 1. Cloud Foundation

- Add `OPENPOST_EDITION=selfhost|cloud`.
- Add `OPENPOST_DATABASE_DRIVER=sqlite|postgres` and Postgres-backed Bun ORM initialization.
- Add `OPENPOST_STORAGE_DRIVER=local|s3` with S3-compatible storage.
- Keep runtime database expressions portable. Background job recovery, job workspace scoping, publish-job cleanup, MCP scheduling cleanup, and schedule overview date aggregation now avoid SQLite-only JSON/date expressions so cloud Postgres deployments use the same paths.
- Add usage counters and entitlement checks at API boundaries. The foundation is in place with monthly `usage_counters`, workspace-creation entitlement checks, team invitation seat checks, and scheduled-post usage accounting.
- Enforce quota boundaries for workspace, team, provider, media, scheduling, and publishing paths. Social account connection quota enforcement is in place in the shared account saver, team invitations reserve active plus pending seats, media upload quota enforcement is in place for monthly uploaded bytes and stored bytes, scheduled-post quota enforcement is in place for single posts and threads, and publishing-worker quota enforcement is in place for published posts and provider write calls.
- Add monthly usage counters for scheduled posts, published posts, uploaded bytes, stored bytes, and provider write calls. The publishing worker records published-post and provider-write usage.

### 2. Billing And Plans

- Use Polar for OpenPost Cloud checkout, subscriptions, customer portal, and webhooks. Cloud mode now requires the Polar access token, webhook secret, checkout/return URLs, and Starter/Creator/Pro product IDs at startup.
- Store local subscription state and entitlement snapshots; do not call Polar on every request. The Polar checkout, customer portal, and webhook foundation now creates hosted billing sessions, verifies signed events, deduplicates webhook deliveries, and upserts workspace subscription snapshots.
- Keep self-hosted entitlement defaults permissive and configurable.
- Keep cloud pre-checkout access constrained. Cloud mode now allows a first bootstrap workspace, then evaluates workspace expansion from active user subscription snapshots instead of falling back to self-hosted unlimited behavior.
- Suggested launch plans:
  - Starter: 3 open-web connections, Bluesky/Mastodon first, 1 workspace, 100 scheduled posts/month, 1 GB media.
  - Creator: 6 connections, X/LinkedIn/Threads/Bluesky/Mastodon, 3 workspaces, 500 scheduled posts/month, 5 GB media.
  - Pro: 15 connections, larger media/history limits, team support when ready.
- Avoid a hosted free tier at launch; use trial/beta access instead.

### 3. Provider Readiness

- Add a provider app registry for cloud and self-hosted credentials. Startup now builds adapters from a normalized registry populated by legacy env vars, optional `OPENPOST_PROVIDER_APPS` JSON, and active encrypted `provider_apps` database rows managed through instance-admin APIs for hosted/operator-managed credentials.
- Replace fixed Mastodon env-only config with dynamic instance registration for cloud.
- Keep user-supplied remote URLs guarded against SSRF. Dynamic Mastodon registration and MCP URL media ingestion now reject private/local targets, validate redirects, use guarded dial-time resolution, and ignore environment proxy settings for these fetches.
- Add production OAuth app checklists for X, LinkedIn, Threads, Facebook, Instagram, YouTube, TikTok, Mastodon, and Bluesky.
- Delay platform launch promises until provider-specific publish, refresh, media, and retry behavior is verified end to end. TikTok, Facebook, Instagram, and YouTube now have first-slice adapters, but all four still need live-account verification before being treated as fully proven production providers.

### 4. Media Pipeline

- Move cloud uploads to direct browser-to-S3/R2 upload sessions. The S3 storage driver now issues authenticated upload sessions with presigned PUT targets, pending media reservations, completion finalization, dedupe, and quota accounting.
- Track media assets separately from provider-uploaded media IDs.
- Store size, checksum, dimensions, duration, processing status, storage driver, object key, and public URL mode.
- Add provider media state for X, LinkedIn, Mastodon, Threads, Instagram,
  Facebook, YouTube, and TikTok. Destination-scoped provider media state now
  records successful upload IDs for retry reuse while avoiding cached public
  URLs for Threads, Instagram, Facebook, and TikTok.
- Keep Threads and other public-URL providers working through signed/public media URLs.

### 5. Publication Model

- Introduce **Publication** as the user-facing unit of intent.
- The first schema and API foundation is in place with `publications`, `publication_assets`, optional `posts.publication_id` links, authenticated create/list/get/update publication endpoints, and post API/MCP draft/schedule flows that preserve source-publication links.
- Keep **Renditions** as destination-specific versions with format-specific validation.
- Keep current `posts` flow working while adding publication tables behind tests.
- Migrate the composer toward source idea, destinations, renditions, and release plan. The composer now shows linked source-publication context, can reapply source copy/media, and hydrates publication media metadata for downstream provider checks.
- Support release choreography: same time, staggered posts, platform-first launches, and follow-up threads.

### 6. MCP And ChatGPT App

- Expose a remote MCP endpoint for OpenPost Cloud at `/mcp`.
- Keep the MCP server backend-owned, not frontend-owned.
- Add a local `openpost-mcp` stdio binary for desktop/self-hosted clients. The CLI now includes a stdio proxy that loads the active OpenPost profile/token and forwards frames to `/mcp`.
- Reuse CLI/API client behavior where possible, but keep MCP stdout strict.
- Start with safe semantic tools and prompts: list workspaces, list accounts, create/list source publications, create/list/update draft, set post renditions, upload media from URL, schedule post or draft, cancel post, get post status, suggest next slot, and prompt templates for planning posts, adapting renditions, and reviewing the queue. The remote MCP foundation now supports workspace/account listing, source publication creation/review, draft creation/review/revision, destination-specific rendition updates, guarded URL media upload, quota-checked scheduling for new posts and existing drafts, post status reads, scheduled-post queue inspection/cancellation, next-slot suggestions, and agentic scheduling prompt templates.
- Require auth for remote MCP, scope sessions, log tool calls, and expose revocation in settings. Tool-call logging is now persisted in `mcp_tool_calls`, recent calls are visible in settings with API-token client attribution, Apps SDK-facing protected-resource/tool security metadata, invocation status labels, and output schemas are in place, Settings can create/revoke dedicated `mcp:full` tokens, OAuth authorization-code + PKCE account linking now mints audience-bound MCP tokens, and both manual tokens and OAuth approvals can be limited to one workspace.

### 7. Marketing, SEO, And Docs

- Keep `marketing-site/` public in this repo.
- Keep `docs-site/` technical and task-oriented.
- Add pricing, blog, comparison, tips, and tools pages to `openpost.social`. The marketing site now includes crawlable tools, tips, blog, and comparison routes with sitemap coverage.
- Add SEO utilities such as post preview, thread splitter, character counter, and UTM builder. The marketing tools page now covers all four with Playwright verification for generated UTM links.
- Keep docs on install, providers, configuration, CLI, operations, and development.

### 8. Verification

- Add Playwright smoke tests for marketing, login, onboarding, composer, scheduling, accounts, settings, and media. Coverage is now in place for marketing, docs audience separation, browser registration/login/onboarding, app settings/billing/MCP activity/session revocation, Activity job pagination, provider discovery, custom Mastodon connect, plan onboarding, publication handoff, account-specific composer previews, composer scheduling through suggested slots, and media-library upload/listing.
- Add backend regression tests before each schema/service change.
- Keep `devenv shell -- lint` as the push gate.
- For hosted deployment work, verify the real app URL, docs URL, marketing URL, release workflow, database backups, and logs.

## First Implementation Order

1. Upgrade marketing-site into a real public front door.
2. Add production-readiness docs and keep links discoverable.
3. Add backend config primitives for edition, database driver, and storage driver.
4. Add storage-driver tests before implementing S3/R2.
5. Add entitlement interfaces and self-host defaults. Done for the service contract and workspace creation boundary.
6. Add usage tables and API boundary checks. Monthly usage counters, workspace quota enforcement, team invitation seat enforcement, social-account quota enforcement, media quota enforcement, scheduled-post quota enforcement, and publishing-worker usage/quota enforcement are in place.
7. Add Playwright coverage around the core app flows.
8. Start MCP with authenticated remote metadata and safe read/create/schedule tools. Remote auth, protected-resource metadata, authorization-server metadata, PKCE account linking, tool security descriptors, Apps SDK output metadata, prompt templates, workspace listing, account listing, guarded URL media upload, draft creation, scheduled posting, status reads, scheduled-post cancellation, next-slot suggestions, settings-visible tool-call activity, and dedicated `mcp:full` API-token creation are in place.
