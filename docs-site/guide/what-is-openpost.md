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

OpenPost is not trying to be a full enterprise social media management suite. It focuses on the core writing and scheduling workflow first, with media support varying by provider.

**Current limitations:**

- **Video support is provider-dependent** — some provider video paths exist in the codebase, but support is not consistent across every platform and not every path is verified end to end
- **No full feature parity guarantee** — each social network has different capabilities, and some provider-specific features may be unavailable in OpenPost
- **Advanced analytics are not the current focus** — engagement tracking and reporting are not a launch feature
- **Enterprise approval workflows are not the current focus** — OpenPost is not positioning itself as an enterprise review-and-approval suite
- **Provider APIs can be restrictive** — each platform has its own API limits, rate limits, and approval requirements that may affect publishing

## What it deliberately avoids

- Redis or external queue requirements for simple deployments
- Postgres as a mandatory dependency
- Hosted-account lock-in
- Splitting the app into multiple services before it is necessary
