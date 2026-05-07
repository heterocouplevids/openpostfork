# OpenPost Launch TODO

> Repo audit date: 2026-05-07
>
> Scope: launch readiness for OpenPost as a self-hosted Typefully-like social media scheduler. This list is based on the current repository contents, not just the earlier ideal roadmap. Runtime/provider checks still need to be performed manually before launch.

## Current launch positioning

Use this consistently across README, docs, launch posts, and screenshots:

> OpenPost is a self-hosted Typefully-like social media scheduler for people who want to write, customize, and schedule posts across platforms without paying for another monthly SaaS.

Secondary technical proof:

> Built with Go, Svelte/SvelteKit, and SQLite. Runs as a single binary, with Docker available.

Hard limitation to show clearly:

> OpenPost currently focuses on text and image-based scheduling. Video posts are not supported yet, and OpenPost does not guarantee full feature parity with every social platform.

---

## What is already done well enough

These are not launch blockers. They only need final verification.

- [x] README has a clean front door with logo, badges, screenshot, docs links, supported platforms, and short quickstart.
- [x] README already says OpenPost is self-hosted, single binary/container, SQLite-backed, supports X/Mastodon/Bluesky/Threads/LinkedIn, has encrypted tokens, MFA/TOTP, passkeys, and threads.
- [x] Docs site exists and is organized with guide, deployment, configuration, providers, usage, operations, and development sections.
- [x] Docker Compose docs include persistent `/data`, `OPENPOST_DATABASE_PATH`, `OPENPOST_MEDIA_PATH`, `OPENPOST_MEDIA_URL`, and a healthcheck.
- [x] Single-binary docs exist.
- [x] Reverse proxy docs exist and include Caddy, Nginx, callback URLs, and the Threads public-media caveat.
- [x] Provider overview exists for X, Mastodon, Bluesky, LinkedIn, and Threads.
- [x] Troubleshooting docs cover startup, OAuth callbacks, CORS, media upload, Threads media, scheduled post failures, database path, SQLite locking, and reverse proxy URL issues.
- [x] Production checklist exists.
- [x] Security policy exists.
- [x] CONTRIBUTING.md exists and gives dev setup, tests, linting, PR expectations, and contribution categories.
- [x] CODE_OF_CONDUCT.md exists.
- [x] Bug report and feature request issue templates exist.
- [x] CHANGELOG.md exists and includes recent public releases.
- [x] Release workflow exists for Docker image, Linux binary, macOS arm64 binary, and Android APK artifacts.
- [x] CI workflow exists for backend/frontend linting and tests.
- [x] Platform adapter architecture exists for the supported providers.
- [x] Custom SQLite-backed background worker is part of the intended architecture; no Redis/Postgres requirement is a real differentiator.

---

## P0 — Must fix before any major HN/Reddit launch

### 1. Fix video-support inconsistency everywhere

This is the largest current launch blocker because the public copy and UI can overpromise.

- [ ] Update README with a visible `Current limitations` section:
  - [ ] No video posts yet.
  - [ ] No full feature parity guarantee for every platform.
  - [ ] Advanced analytics are not the current focus.
  - [ ] Enterprise approval workflows are not the current focus.
- [ ] Update `docs-site/guide/what-is-openpost.md` to explain:
  - [ ] Text and image-based scheduling is the current focus.
  - [ ] Video posts are not supported yet.
  - [ ] Platform APIs differ and some features may be unavailable.
- [ ] Fix `docs-site/providers/platform-limits.md` so it does not imply OpenPost supports video publishing.
  - [ ] Change all provider video cells to `Not supported by OpenPost yet`, unless there is a fully verified provider-specific implementation.
  - [ ] Add a note that provider-native capabilities are not the same as OpenPost-supported capabilities.
- [ ] Fix `frontend/src/lib/components/media-upload.svelte`.
  - [ ] Remove `video/*` from the file input accept list until video support is real.
  - [ ] Reject video files client-side with a clear message instead of silently uploading or pretending support exists.
  - [ ] Change the UI copy from `Drop images or videos here` to image-only copy.
- [ ] Add or confirm backend validation in the media upload handler.
  - [ ] Reject video MIME types until video support is officially implemented.
  - [ ] Return a clear error such as `video uploads are not supported yet`.
- [ ] Add a regression test for uploading a video file, expecting rejection.

### 2. Make README launch-ready, not just technically correct

- [ ] Replace the hero/tagline with a stronger user-value message:
  - [ ] `Self-hosted social media scheduling, without another monthly subscription.`
  - [ ] Mention Typefully-like workflow in the first paragraph.
- [ ] Add `Who is OpenPost for?`:
  - [ ] creators who want scheduling without another SaaS subscription
  - [ ] indie hackers managing multiple project accounts
  - [ ] small teams/agencies with separate brand workspaces
  - [ ] self-hosters who want their own social scheduler
  - [ ] open-source maintainers posting across multiple platforms
- [ ] Add `Current limitations` as above.
- [ ] Add a short feature/support table:
  - [ ] Self-hosted
  - [ ] Single binary
  - [ ] Docker support
  - [ ] SQLite
  - [ ] X/Mastodon/Bluesky/Threads/LinkedIn
  - [ ] Threads composer
  - [ ] Platform-specific variants
  - [ ] Media library
  - [ ] 2FA/TOTP
  - [ ] Passkeys
  - [ ] Video posts: not yet
  - [ ] Analytics: not yet
- [ ] Reduce security/implementation detail in the top pitch; keep encrypted tokens and AES details lower down.

### 3. Complete docs that are currently too thin

#### Quickstart

- [ ] Add a single-binary path or clearly link to it from the first screen.
- [ ] Add a minimal `first scheduled post` flow, not just `publish a test post`.
- [ ] Add a short provider setup path for the easiest provider to test first.
- [ ] Add expected first-run screens and what success looks like.
- [ ] Make the `.env` instruction usable outside a cloned repo; `cp backend/.env.example .env` only works if the user has the repo checkout.

#### Single binary install

- [ ] Add a complete `.env` example for binary installs.
- [ ] Add recommended production paths, for example `/var/lib/openpost/openpost.db` and `/var/lib/openpost/media`.
- [ ] Add a systemd service example.
- [ ] Add upgrade steps for replacing the binary safely.
- [ ] Add backup reminder before upgrade.

#### Backup and restore

- [ ] Add a real restore process, not only backup commands.
- [ ] Document how to restore:
  - [ ] SQLite DB
  - [ ] media directory
  - [ ] `.env` or secrets
  - [ ] ownership/permissions
  - [ ] service restart
- [ ] Add migration-to-another-server instructions.
- [ ] Add a `test restore` checklist.
- [ ] Mention SQLite WAL/shm files if relevant to the deployment mode.

#### Supported platforms

- [ ] Add a supported/unsupported matrix by platform:
  - [ ] text posts
  - [ ] image posts
  - [ ] threads/replies
  - [ ] scheduled posts
  - [ ] video posts
  - [ ] platform-specific variants
  - [ ] analytics
- [ ] Add provider-specific caveats:
  - [ ] X developer app requirements and OAuth setup
  - [ ] Mastodon per-instance setup
  - [ ] Bluesky app password flow
  - [ ] LinkedIn permissions/review caveats
  - [ ] Threads public media URL caveat
- [ ] Make clear that provider API restrictions can change.

#### Security docs

- [ ] Update `SECURITY.md` to mention TOTP and passkeys in the security features list.
- [ ] Add operational guidance for protecting `.env`, database, media folder, and backups.
- [ ] Add HTTPS recommendation near OAuth setup and first-run docs, not only in security/reverse proxy docs.
- [ ] Add a short note that encrypted provider tokens do not remove the need to protect `OPENPOST_ENCRYPTION_KEY`.

### 4. Verify product readiness from scratch

Static docs/code inspection is not enough. Perform these manually on a fresh machine or clean VPS.

- [ ] Install from scratch using Docker Compose docs.
- [ ] Install from scratch using binary docs.
- [ ] Confirm Docker image pulls from GHCR using the documented tag.
- [ ] Confirm binary asset names match the docs and release workflow.
- [ ] Confirm OpenPost starts with only required env vars plus one provider configured.
- [ ] Confirm persistent database survives restart.
- [ ] Confirm media survives restart.
- [ ] Confirm scheduled posts survive restart.
- [ ] Confirm first account becomes admin.
- [ ] Confirm `OPENPOST_DISABLE_REGISTRATIONS=true` still allows first account on a brand-new instance and blocks later registrations.
- [ ] Confirm 2FA/TOTP enrollment, login, recovery/removal flow.
- [ ] Confirm passkey enrollment, login, and removal flow.
- [ ] Confirm password reset/change behavior if implemented; otherwise document what exists.
- [ ] Confirm health endpoint returns expected response.
- [ ] Confirm logs are enough to debug failed OAuth and failed publishing.

### 5. Verify provider workflows before launch claims

For each provider, test both connection and at least one real post from a clean instance.

- [ ] X:
  - [ ] connect account
  - [ ] publish text post
  - [ ] publish image post if claimed
  - [ ] schedule post and confirm it publishes
  - [ ] verify graceful failure message for rejected post
- [ ] Mastodon:
  - [ ] connect account for documented instance
  - [ ] publish text post
  - [ ] publish image post if claimed
  - [ ] schedule post and confirm it publishes
  - [ ] verify instance URL is stored/used correctly
- [ ] Bluesky:
  - [ ] connect with app password
  - [ ] publish text post
  - [ ] publish image post if claimed
  - [ ] schedule post and confirm it publishes
- [ ] Threads:
  - [ ] connect account
  - [ ] publish text post
  - [ ] publish image post if claimed
  - [ ] confirm public media URL requirement is documented and works
- [ ] LinkedIn:
  - [ ] connect account
  - [ ] publish text post
  - [ ] publish image post if claimed
  - [ ] document any permission/review limitations

### 6. Tighten diagnostics and support paths

- [ ] Add a documented diagnostics command or support checklist:
  - [ ] OpenPost version
  - [ ] deployment method
  - [ ] OS/architecture
  - [ ] relevant env vars without secrets
  - [ ] provider being tested
  - [ ] last 100 logs
  - [ ] health endpoint result
- [ ] Add provider-specific troubleshooting pages or sections for the most common OAuth/publishing errors.
- [ ] Improve UI error messages for failed provider publishing where they are still generic.
- [ ] Add a `Known limitations` docs page and link it from README, quickstart, and provider overview.

### 7. GitHub/community launch prep

- [ ] Create or verify GitHub labels:
  - [ ] `bug`
  - [ ] `docs`
  - [ ] `provider`
  - [ ] `good first issue`
  - [ ] `help wanted`
  - [ ] `question`
  - [ ] `roadmap`
  - [ ] `enhancement`
- [ ] Add missing issue templates:
  - [ ] provider issue
  - [ ] docs issue
  - [ ] installation/support issue
- [ ] Enable/configure GitHub Discussions if you want support outside issues.
- [ ] Add discussion categories if enabled:
  - [ ] announcements
  - [ ] support
  - [ ] ideas
  - [ ] show and tell
- [ ] Create pinned issue: `Launch feedback thread`.
- [ ] Create pinned issue: `Known launch issues`.
- [ ] Add 5-10 real `good first issue` items.
- [ ] Confirm the repo social preview image is set in GitHub settings.
- [ ] Confirm releases are visible and latest release is the one you want people installing.

### 8. Prepare marketing assets

- [ ] Capture current screenshots:
  - [ ] main dashboard
  - [ ] composer
  - [ ] platform variants/customization
  - [ ] thread composer
  - [ ] scheduling/calendar view
  - [ ] media library
  - [ ] workspaces
  - [ ] connected accounts
  - [ ] security/account settings showing 2FA/passkeys
- [ ] Replace any stale screenshots in README/docs.
- [ ] Create a 30-60 second demo GIF/video showing:
  - [ ] open composer
  - [ ] write post
  - [ ] select multiple platforms
  - [ ] customize one platform variant
  - [ ] add image
  - [ ] schedule
  - [ ] show queued/scheduled post
  - [ ] show workspace/account separation
- [ ] Create social preview assets:
  - [ ] GitHub social preview
  - [ ] Open Graph image for docs site
  - [ ] 16:9 image for LinkedIn/X/Bluesky/Mastodon
  - [ ] Reddit screenshot album
  - [ ] small logo/icon for posts

### 9. Prepare launch copy before posting

- [ ] HN title:
  - [ ] `Show HN: OpenPost – self-hosted social media scheduler in Go, Svelte and SQLite`
- [ ] HN first comment.
- [ ] r/selfhosted title:
  - [ ] `I built OpenPost, a self-hosted Typefully-like social media scheduler`
- [ ] r/selfhosted body with screenshots and limitations.
- [ ] LinkedIn post.
- [ ] X/Bluesky/Mastodon short post.
- [ ] GitHub release notes for the launch release.
- [ ] Reply bank:
  - [ ] How is this different from Postiz?
  - [ ] Does it support videos?
  - [ ] Does it support all features of each platform?
  - [ ] Why self-host this instead of Typefully/Buffer?
  - [ ] Is this production-ready?
  - [ ] What happens when provider APIs change?
  - [ ] Can teams/agencies use it?
  - [ ] How do I back it up?

---

## P1 — Strongly recommended before launch

### Product polish

- [ ] Add or improve onboarding checklist inside the app:
  - [ ] create account
  - [ ] create workspace
  - [ ] connect provider
  - [ ] create first post
  - [ ] schedule/publish
- [ ] Add clearer empty states for dashboard, posts, media, accounts, and schedules.
- [ ] Add explicit provider connection status in the account UI.
- [ ] Add retry/requeue UX for failed scheduled posts if not already obvious.
- [ ] Add a visible app version somewhere in settings/about.
- [ ] Add copyable diagnostics in settings/about if feasible.

### Tests

- [ ] Add backend tests around scheduled publishing persistence after restart or worker restart.
- [ ] Add tests for provider rejection/failure handling.
- [ ] Add auth tests around first-admin setup and registration disabling.
- [ ] Add media validation tests for unsupported file types.
- [ ] Add frontend smoke tests for the first-run flow if practical.

### Docs and positioning

- [ ] Add `OpenPost vs Typefully/Buffer/Postiz` page:
  - [ ] neutral tone
  - [ ] avoid dunking on competitors
  - [ ] explain OpenPost is smaller, focused, and self-hosted
  - [ ] acknowledge hosted tools are easier for many users
- [ ] Add `Why self-host a scheduler?` page:
  - [ ] own your workflow
  - [ ] avoid another monthly subscription
  - [ ] keep drafts/media/schedules under your control
  - [ ] run it on your own VPS/homelab
- [ ] Add FAQ:
  - [ ] Does OpenPost support video?
  - [ ] Does it support analytics?
  - [ ] Can I use it for teams?
  - [ ] Can I run it without Docker?
  - [ ] How do I back it up?
  - [ ] How is it different from Postiz?
  - [ ] Is this production-ready?
  - [ ] What happens if a provider API changes?

### Roadmap cleanup

- [ ] Split `ROADMAP.md` into clearer sections:
  - [ ] Launch blockers
  - [ ] Near-term stabilization
  - [ ] Later growth features
  - [ ] Ideas, not commitments
- [ ] Move API keys, MCP, Directus, Genkit, analytics, and Android APK messaging out of the launch-critical path.
- [ ] Add `Not currently planned before launch` section.
- [ ] Add no-video/full-parity caveats.
- [ ] Keep the roadmap honest and less hype-driven.

---

## P2 — Nice, but do not delay launch

- [ ] Full analytics.
- [ ] Video support.
- [ ] Full provider feature parity.
- [ ] Approval workflows.
- [ ] Advanced team roles.
- [ ] API keys.
- [ ] MCP server.
- [ ] Directus integration.
- [ ] AI writing assistance.
- [ ] One-click deploy templates.
- [ ] Kubernetes/Helm.
- [ ] Homebrew package.
- [ ] Nix package.
- [ ] Mobile app polish.

---

## Soft launch checklist

Do this before Reddit/HN.

- [ ] Install OpenPost on a fresh VPS using only public docs.
- [ ] Ask 3-5 people to install it from scratch.
- [ ] Watch where they get confused.
- [ ] Fix docs immediately.
- [ ] Collect top 10 repeated questions.
- [ ] Turn repeated questions into FAQ/docs.
- [ ] Collect first screenshots/testimonials if people give permission.
- [ ] Post quietly on personal LinkedIn/X/Bluesky/Mastodon.
- [ ] Share with friends/builders/self-hosters who can give useful feedback.

---

## Main launch order

Do not post everywhere at once.

### Day 0 — r/selfhosted

- [ ] Post with screenshots.
- [ ] Lead with self-hosted and no-monthly-SaaS value.
- [ ] Be explicit that it is still early.
- [ ] Be explicit about no video support.
- [ ] Stay active in comments.
- [ ] Convert repeated feedback into docs/issues.

### Day 1 or 2 — Hacker News

- [ ] Post only after fixing install/docs issues from Reddit.
- [ ] Use more technical copy.
- [ ] Emphasize Go/Svelte/SQLite/single-binary architecture as proof, not as the whole pitch.
- [ ] Be candid about tradeoffs and API limitations.

### Day 3-7 — Secondary launches

- [ ] r/opensource.
- [ ] r/webdev.
- [ ] r/indiehackers.
- [ ] Indie Hackers.
- [ ] Mastodon self-hosting/open-source circles.
- [ ] Bluesky builder/open-source circles.
- [ ] LinkedIn.

Only post when you can actively reply.

---

## Launch-day checklist

### Before posting

- [ ] Re-test Docker install.
- [ ] Re-test binary install.
- [ ] Re-test docs links.
- [ ] Re-test GitHub release links.
- [ ] Re-test Docker image pull.
- [ ] Re-test binary download.
- [ ] Re-test screenshots and demo media.
- [ ] Open GitHub issues/discussions tabs.
- [ ] Have reply bank ready.
- [ ] Have several hours free.

### While live

- [ ] Reply quickly.
- [ ] Be honest about limitations.
- [ ] Do not argue with competitor comparisons.
- [ ] Ask for logs when people hit bugs.
- [ ] Convert repeated comments into docs updates.
- [ ] Patch docs immediately.
- [ ] Keep `Known launch issues` updated.
- [ ] Thank people who install/test it.
- [ ] Ask people what platform they care about most.
- [ ] Track feature requests, but do not promise everything.

---

## First 48 hours after launch

- [ ] Triage every bug report.
- [ ] Fix broken docs immediately.
- [ ] Fix install friction immediately.
- [ ] Add FAQ entries from repeated questions.
- [ ] Add provider-specific caveats from real feedback.
- [ ] Label all GitHub issues.
- [ ] Reply to serious comments.
- [ ] Keep known issues visible and updated.
- [ ] Push small patch releases as needed:
  - [ ] `v1.0.4`
  - [ ] `v1.0.5`
  - [ ] or whatever the next semver version is at launch time
- [ ] Write a short launch feedback summary.
- [ ] Add `what changed after launch` changelog entries.
- [ ] Thank first testers publicly.
- [ ] Create issues for top requested features.

---

## First two weeks after launch

### Product

- [ ] Fix top 5 bugs.
- [ ] Improve top 5 confusing UI flows.
- [ ] Improve provider error handling.
- [ ] Improve onboarding.
- [ ] Improve backup/restore docs.
- [ ] Improve install docs.
- [ ] Add clearer supported/unsupported platform matrix.
- [ ] Add more tests around scheduling/publishing.
- [ ] Add import/export only if users request it heavily.

### Community

- [ ] Keep GitHub issues organized.
- [ ] Convert good requests into roadmap items.
- [ ] Close duplicates politely.
- [ ] Add contributors to README if people help.
- [ ] Keep discussions alive.
- [ ] Publish weekly or biweekly updates.

### Marketing

- [ ] Post `what I learned launching OpenPost`.
- [ ] Post `OpenPost now has X improvements after launch feedback`.
- [ ] Share small clips/screenshots.
- [ ] Write `Why I built a smaller self-hosted social scheduler`.
- [ ] Submit to AlternativeTo.
- [ ] Submit to awesome-selfhosted if eligible.
- [ ] Submit to relevant GitHub awesome lists.
- [ ] Submit to self-hosted/open-source directories.

---

## First month after launch

### Stabilization

- [ ] Reliable release process.
- [ ] Better provider regression testing.
- [ ] More OAuth/provider app setup docs.
- [ ] Better diagnostics/logging.
- [ ] Backup/restore fully documented and tested.
- [ ] Clear upgrade guide.

### Growth features to choose from feedback

Do not pick these from vibes. Choose from launch feedback.

- [ ] Better calendar/schedule UX.
- [ ] Better drafts workflow.
- [ ] Better thread composer.
- [ ] More platform-specific customization.
- [ ] Bulk scheduling.
- [ ] Import/export posts.
- [ ] Webhooks.
- [ ] API keys.
- [ ] MCP support.
- [ ] Basic analytics.
- [ ] Video support, only if demand is high and provider APIs make it reasonable.

### Distribution

- [ ] Coolify docs/template.
- [ ] Unraid docs/template.
- [ ] Synology/Container Manager docs.
- [ ] TrueNAS Scale docs.
- [ ] CasaOS/YunoHost only if there is demand.
- [ ] Helm chart only if Kubernetes users ask.
- [ ] Homebrew/Nix package only if binary distribution is stable.

---

## Practical launch gate

Launch when all of these are true:

- [ ] I can install OpenPost from scratch using public docs.
- [ ] I can run it as a single binary.
- [ ] I can run it with Docker.
- [ ] I can connect at least one provider without guessing.
- [ ] I can schedule a post and see it survive restart.
- [ ] I can publish a post on each platform I publicly claim to support.
- [ ] The README clearly explains who OpenPost is for.
- [ ] The docs clearly explain what is not supported.
- [ ] No UI/docs imply video support.
- [ ] The roadmap is not stale or misleading.
- [ ] Screenshots/demo are current.
- [ ] GitHub release exists and has the right artifacts.
- [ ] Known limitations are visible.
- [ ] HN/Reddit/LinkedIn/social copy is ready.
- [ ] I have several hours free to respond.

The most important remaining work is not adding more features. It is making the first 10 minutes obvious: what OpenPost is, who it is for, how to install it, what it supports, and what it does not support.
