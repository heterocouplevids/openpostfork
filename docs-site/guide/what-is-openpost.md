# What Is OpenPost?

OpenPost is a self-hosted Buffer/Hootsuite alternative — a social media scheduler with a Typefully-like writing and scheduling workflow. It lets you connect social accounts, compose posts, customize per platform, attach media, schedule publishing, and manage threads from your own server.

It is built for operators who want a practical publishing tool without handing content, tokens, and schedules to another monthly SaaS subscription. OpenPost keeps the stack intentionally small: Go, SvelteKit, SQLite, local media storage, and a single deployable binary or container.

## Who is this for?

OpenPost is for people who want the core social media scheduling workflow without relying on another hosted SaaS.

- **Indie hackers and creators** who need a cheaper or free alternative to Typefully, Buffer, or Hootsuite
- **Small teams** that want control over credentials, data, and branding
- **Open-source maintainers** managing multiple platform presences
- **Self-hosters** who want a lightweight tool instead of a full marketing suite
- **Agencies** managing separate brand workspaces

## What OpenPost supports

- X
- Mastodon
- Bluesky
- LinkedIn
- Threads

## What OpenPost is not

OpenPost is not trying to be a full enterprise social media management suite. It focuses on the core writing and scheduling workflow for text and image-based posts.

**Current limitations:**

- **No video support yet** — video uploads are not implemented for any platform
- **No full feature parity guarantee** — each social network has different capabilities; OpenPost provides the core features but may not support every platform-specific option
- **Analytics not built yet** — engagement tracking and reporting are on the roadmap but not yet available
- **Provider APIs can be restrictive** — each platform has its own API limits, rate limits, and approval requirements that may affect publishing

## What it deliberately avoids

- Redis or external queue requirements for simple deployments
- Postgres as a mandatory dependency
- Hosted-account lock-in
- Splitting the app into multiple services before it is necessary
