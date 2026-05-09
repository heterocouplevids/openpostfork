# Why Self-Host OpenPost?

Self-hosting a social media scheduler might seem like extra work compared to using a hosted tool like Typefully, Buffer, or Hootsuite. Here is why you might consider it.

## Control Your Data

When you use a hosted SaaS, your drafts, schedules, media, and connected account tokens live on their servers. Self-hosting keeps everything on your infrastructure:

- Your post drafts stay on your server
- Your media library stays on your server
- Your schedules and history stay on your server
- Your OAuth tokens stay on your server

## Avoid Another Subscription

Hosted social media schedulers add up. Most charge per seat or per connected account. OpenPost runs on any small VPS, homelab, or even a Raspberry Pi — you already have the server, now you have the tool.

## Keep Tokens Secure

OpenPost encrypts OAuth tokens at rest with AES-256-GCM. Your tokens never leave your server, and you control who can access them.

## No Lock-In

If you decide to move to another tool or stop using a scheduler altogether, your data is yours. Export your posts, media, and schedules — nothing is trapped in someone else's platform.

## Run It Your Way

- Deploy with Docker or run the binary directly
- Back up with a single SQLite file
- Upgrade when you want, not when a SaaS forces an update
- Extend it if you need to — it is open source

## When to Consider a Hosted Tool Instead

OpenPost is intentionally focused on the core scheduling workflow. If you need:

- Built-in analytics and reporting
- Large-team approval workflows
- Enterprise compliance features
- Customer support from a vendor

…a hosted SaaS may be a better fit today. OpenPost aims to stay small, focused, and easy to run.

## TL;DR

OpenPost is for people who want the core Typefully/Buffer workflow — write, customize, schedule, publish — without handing their content and credentials to another monthly SaaS. If you already have a server and value control over your data, self-hosting OpenPost gives you a lightweight scheduling tool that stays out of your way.