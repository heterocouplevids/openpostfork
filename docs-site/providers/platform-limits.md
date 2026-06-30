# Supported Platforms & Limitations

OpenPost currently supports publishing to X, Mastodon, Bluesky, Threads, and LinkedIn. Instagram, Facebook, YouTube, and TikTok are visible in provider discovery as planned adapters only.

Provider-native API capabilities are not the same as OpenPost-supported capabilities. The table below reflects the current OpenPost implementation state, including paths that still need real-account verification.

## Current Platform Support

| Provider | Text      | Images                                              | Video                                                                                          | Threading                              | Scheduling | Variants  |
| -------- | --------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------- | ---------- | --------- |
| X        | Supported | Up to 4 images                                      | Implemented, real-account verification still required                                          | Replies                                | Supported  | Supported |
| Mastodon | Supported | Up to 4 attachments                                 | Implemented through media upload + publish flow, real-account verification still required      | Replies                                | Supported  | Supported |
| Bluesky  | Supported | Up to 4 images                                      | Implemented for one MP4 video via `app.bsky.video.*`, real-account verification still required | AT Protocol reply refs                 | Supported  | Supported |
| LinkedIn | Supported | Single-image path supported                         | Implemented and recently fixed, still needs re-verification against the live API               | Thread children are posted as comments | Supported  | Supported |
| Threads  | Supported | Single-media path in current OpenPost composer flow | Implemented with MIME-aware media handling and public media URL requirement                    | `reply_to_id`                          | Supported  | Supported |

## Planned Platform Adapters

Planned providers are intentionally not connectable yet. They appear in the provider discovery API with `status: "planned"` so web, CLI, MCP, and ChatGPT App clients can expose the roadmap without attempting OAuth.

| Provider  | Planned focus                                                               | Connectable today |
| --------- | --------------------------------------------------------------------------- | ----------------- |
| Instagram | Image posts, Reels, scheduling, platform variants, agent workflows          | No                |
| Facebook  | Facebook Pages, media posts, scheduling, platform variants, agent workflows | No                |
| YouTube   | Shorts, video publishing, scheduling, agent workflows                       | No                |
| TikTok    | Short-form video publishing, scheduling, agent workflows                    | No                |

## Known Limitations

- **Video support is uneven** — implementation exists across multiple providers, but support is still provider-dependent and some paths need end-to-end verification with real accounts.
- **No full feature parity guarantee** — OpenPost provides the core scheduling features but may not support every platform-specific option (e.g., polls, galleries, stories)
- **Planned providers are discovery-only** — adding Instagram, Facebook, YouTube, or TikTok to provider app config fails until their adapters are implemented.
- **Provider APIs can change** — social platforms may change their APIs, rate limits, or app review requirements at any time
- **OAuth tokens require HTTPS** — callbacks need a valid domain with TLS for OAuth to work

## Reading this table correctly

- A provider can support a feature natively while OpenPost still marks it unsupported or unverified.
- "Implemented" means the code path exists in OpenPost.
- "Verified" means the implementation has been confirmed against a live provider account recently.
- Deployment details still matter. Threads in particular depends on a public media URL, and LinkedIn depends heavily on granted app permissions.

These limits are a starting point, not a permanent contract. Providers can change them.
