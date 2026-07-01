# OpenPost Roadmap

> Status: June 2026 — post-v1 priorities.

OpenPost is a lightweight, self-hosted social media scheduler. The core product now includes the web app, CLI, API tokens, active web-session revocation, social sets, TOTP, passkeys, workspaces, media library, and provider publishing flows. The next phase should focus on reliability, provider correctness, and operator polish before larger integrations.

---

## Current Focus

1. **Release hardening**
   Keep Docker, binary, CLI, Android, and docs release paths reproducible. Release tags should run backend, frontend, CLI, and docs checks before artifacts are published.

2. **Scheduler reliability**
   Improve retry behavior and activity-state reporting. Stale processing jobs are now recovered by workers with SQLite/Postgres-compatible database queries.

3. **Provider verification**
   Re-test real-account publishing flows for text, images, threads, and video where support is claimed. Keep public docs conservative where provider APIs or review requirements are uncertain.

4. **Operator documentation**
   Keep `.env.example`, Docker, reverse proxy, backup/restore, provider setup, and production checklist docs aligned with runtime behavior.

---

## Upcoming Features

1. **Enhanced thread management**
   Add safer atomic updates to scheduled or failed threads.

2. **Full pagination for remaining list endpoints**
   Add cursor or offset pagination metadata for the remaining large lists. Posts and background jobs now support offset pagination across the API and CLI, with Activity page load-more coverage for jobs.

3. **Analytics and engagement tracking**
   Poll provider APIs for engagement metrics and display them in an analytics dashboard.

4. **MCP server**
   Add an official automation server for local tools and agents.

5. **Writing assistance**
   Add optional rewrite and content brainstorming workflows without making OpenPost depend on a hosted provider.

6. **Directus integration**
   Explore two-way sync with Directus for users who want a separate content archive.

7. **Spanish localization**
   Complete Spanish translations with a real translation pass. The old stub was removed in v1.0.x.

---

## Technical Debt and Polish

- Expand backend test coverage around publishing, authentication, provider adapters, and queue recovery.
- Add browser-level smoke tests for first-run setup, workspace creation, draft creation, scheduling, and activity status.
- Improve mobile and Android app polish after the web flow is stable.
- Keep the composer fast, readable, and conservative about provider-specific limitations.
