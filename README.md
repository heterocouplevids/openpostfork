<p align="center">
  <a href="https://github.com/rodrgds/openpost">
    <img alt="OpenPost Logo" src="./assets/brand/logo.svg" width="280"/>
  </a>
</p>

<p align="center">
  <a href="https://github.com/rodrgds/openpost/releases">
    <img src="https://img.shields.io/github/v/release/rodrgds/openpost?sort=semver&label=Release" alt="Latest Release">
  </a>
  <a href="https://github.com/rodrgds/openpost/pkgs/container/openpost">
    <img src="https://img.shields.io/github/v/release/rodrgds/openpost?sort=semver&label=Image&include_prereleases" alt="Container Image">
  </a>
  <a href="https://github.com/rodrgds/openpost/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/rodrgds/openpost/ci.yml?label=CI" alt="CI Status">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT">
  </a>
  <a href="SECURITY.md">
    <img src="https://img.shields.io/badge/Security-Security%20Policy-blue" alt="Security Policy">
  </a>
</p>

<div align="center">
  <strong>
    <h2>Self-hosted social media scheduling, without another monthly subscription.</h2>
  </strong>
  OpenPost is a self-hosted Typefully-like social media scheduler for people who want to write, customize, and schedule posts across platforms from their own server.
</div>

<div align="center">
  <br/>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/x-white.svg">
    <img alt="X (Twitter)" src="./assets/logos/x.svg" width="24">
  </picture>
  &nbsp;
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/mastodon-white.svg">
    <img alt="Mastodon" src="./assets/logos/mastodon.svg" width="24">
  </picture>
  &nbsp;
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/bluesky-white.svg">
    <img alt="Bluesky" src="./assets/logos/bluesky.svg" width="24">
  </picture>
  &nbsp;
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/threads-white.svg">
    <img alt="Threads" src="./assets/logos/threads.svg" width="24">
  </picture>
  &nbsp;
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/linkedin-white.svg">
    <img alt="LinkedIn" src="./assets/logos/linkedin.svg" width="24">
  </picture>
</div>

<p align="center">
  <br/>
  <a href="https://docs.openpost.social/"><strong>Documentation</strong></a>
  ·
  <a href="https://docs.openpost.social/guide/quickstart"><strong>Quickstart</strong></a>
  ·
  <a href="https://github.com/rodrgds/openpost/releases"><strong>Releases</strong></a>
</p>

<p align="center">
  <img alt="OpenPost main dashboard screenshot" src="./assets/screenshots/main-dark.png" width="960">
</p>

## Why OpenPost

OpenPost is for people who want the core social media scheduling workflow without relying on another hosted SaaS.

- **Typefully-like composer**: write once, customize per platform with account-specific variants
- **Thread support**: publish multi-post threads in sequence
- **Scheduling that stays queued**: plan posts ahead, queued posts survive restarts
- **Workspaces**: separate brands, projects, or clients into different workspaces
- **Reusable media library**: upload once, reuse across posts
- **Self-hosted**: your data, schedules, and tokens stay on your server

Built with Go, SvelteKit, and SQLite. Runs as a single binary or container with no Redis, no Postgres, and no external queue.

## Who is this for?

OpenPost is especially useful for:

- **Creators** who want scheduling without another SaaS subscription
- **Indie hackers** who want a cheaper or free alternative to Typefully, Buffer, or Hootsuite
- **Small teams** that want control over credentials and data
- **Open-source maintainers** managing multiple platform presences
- **Self-hosters** who want a lightweight tool instead of a full marketing suite
- **Agencies** managing separate brand workspaces

## Feature Snapshot

| Capability | Status |
|---|---|
| Self-hosted | Yes |
| Single binary | Yes |
| Docker support | Yes |
| Android app (Capacitor) | Yes (APK in every release) |
| SQLite | Yes |
| X, Mastodon, Bluesky, Threads, LinkedIn | Yes |
| Threads composer | Yes |
| Platform-specific variants | Yes |
| Media library | Yes |
| 2FA / TOTP | Yes |
| Passkeys | Yes |
| Video posts | Partial, provider-dependent |
| Analytics | Not a launch feature |

## Current Limitations

- **Video support is provider-dependent** — some platforms have implementation paths in the codebase, but not every provider is verified end to end
- **No full feature parity guarantee** — each platform has different capabilities
- **Advanced analytics are not the current focus** — engagement reporting is not a launch feature
- **Enterprise approval workflows are not the current focus** — OpenPost is optimized for core scheduling workflows

## Security And Operations

OAuth tokens are encrypted at rest, and OpenPost supports account-level 2FA/TOTP and passkeys. In production, keep `OPENPOST_JWT_SECRET` and `OPENPOST_ENCRYPTION_KEY` unique and private, run behind HTTPS for OAuth callbacks, and back up the database, media directory, and secrets together.

## Quickstart

```bash
docker compose up -d
```

Set fresh values for `OPENPOST_JWT_SECRET` and `OPENPOST_ENCRYPTION_KEY` before using OpenPost. Both secrets are required and must be at least 32 characters long. The first account created on an instance becomes the instance admin automatically. For the full install path, reverse proxy setup, provider OAuth guides, and operations docs, use the docs site.

## Supported Platforms

- X
- Mastodon
- Bluesky
- Threads
- LinkedIn

## Documentation

- [Landing and docs site](https://docs.openpost.social/)
- [Quickstart](https://docs.openpost.social/guide/quickstart)
- [Installation](https://docs.openpost.social/installation/docker-compose)
- [Android app](https://docs.openpost.social/installation/android)
- [Configuration](https://docs.openpost.social/configuration/environment-variables)
- [Providers](https://docs.openpost.social/providers/overview)
- [Operations](https://docs.openpost.social/operations/troubleshooting)
- [Development](https://docs.openpost.social/development/setup)

## Contributing

Use the development docs in the documentation site, the repo guidance in `AGENTS.md`, and the existing code patterns in `frontend/` and `backend/`.

## Security

Report security issues through [SECURITY.md](SECURITY.md).

## License

MIT
