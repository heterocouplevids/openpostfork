---
layout: home

hero:
  name: OpenPost
  text: Self-hosted Buffer/Hootsuite alternative.
  tagline: Schedule posts to X, Mastodon, Bluesky, Threads, LinkedIn, Facebook Pages, and TikTok from your own server. One binary or container. No Redis, no Postgres, no external queue.
  image:
    src: /assets/brand/logo-docs.svg
    alt: OpenPost logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/rodrgds/openpost

features:
  - title: Composer
    details: Write once, customize per platform, and preview posts before scheduling.
  - title: Threads
    details: Build multi-post threads and publish them in sequence.
  - title: Scheduling
    details: Plan posts ahead with durable jobs that survive restarts.
  - title: Media library
    details: Upload once and reuse media across drafts and scheduled posts.
  - title: Workspaces
    details: Keep separate brands, accounts, media, prompts, and schedules organized.
  - title: CLI
    details: Create posts, upload media, use social sets, and automate workflows from scripts or CI.
  - title: MCP
    details: Let assistants inspect workspaces, draft publications, adapt renditions, and schedule posts through authenticated tools.
  - title: Android app
    details: Install the release APK and connect it to your self-hosted instance.
---

<p>
  <img
    src="/assets/screenshots/main-dark.png"
    alt="OpenPost main dashboard"
    style="width: 100%; max-width: 1200px; border-radius: 16px; border: 1px solid var(--vp-c-divider);"
  >
</p>

## Install in a minute

```yaml
services:
  openpost:
    image: ghcr.io/rodrgds/openpost:latest
    container_name: openpost
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "8080:8080"
    volumes:
      - openpost_data:/data
    environment:
      - OPENPOST_PORT=8080
      - OPENPOST_DATABASE_PATH=/data/db/openpost.db
      - OPENPOST_MEDIA_PATH=/data/media

volumes:
  openpost_data:
```

::: tip
New to OpenPost? Start with the [Quickstart](/guide/quickstart) guide.
:::

## Choose the right docs

- **[User-facing docs](/usage/)** cover the web app, CLI, and MCP workflows for creating publications, adapting platform renditions, scheduling posts, and automating OpenPost as a product user.
- **[Self-hosting docs](/self-hosting/)** cover installation, configuration, provider app setup, media/database storage, backups, upgrades, and troubleshooting for operators.
- **[Developer docs](/development/)** cover architecture, API reference, backend/frontend internals, platform adapters, MCP implementation, billing infrastructure, testing, and the production-readiness plan.

## More ways to use OpenPost

- Use the [CLI](/cli/) for terminal workflows, cron jobs, and CI automation.
- Connect an assistant through [MCP](/mcp/) for agentic drafting, rendition, and scheduling workflows.
- Install the [Android app](/installation/android) from the APK shipped with each GitHub release.
