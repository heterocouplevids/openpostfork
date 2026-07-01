# Supported Platforms & Limitations

OpenPost currently supports publishing to X, Mastodon, Bluesky, Threads, LinkedIn, Facebook Pages, Instagram Business, TikTok, and YouTube.

Provider-native API capabilities are not the same as OpenPost-supported capabilities. The table below reflects the current OpenPost implementation state, including paths that still need real-account verification.

## Current Platform Support

| Provider | Text      | Images                                              | Video                                                                                          | Threading                              | Scheduling | Variants  |
| -------- | --------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------- | ---------- | --------- |
| X        | Supported | Up to 4 images                                      | Implemented, real-account verification still required                                          | Replies                                | Supported  | Supported |
| Mastodon | Supported | Up to 4 attachments                                 | Implemented through media upload + publish flow, real-account verification still required      | Replies                                | Supported  | Supported |
| Bluesky  | Supported | Up to 4 images                                      | Implemented for one MP4 video via `app.bsky.video.*`, real-account verification still required | AT Protocol reply refs                 | Supported  | Supported |
| LinkedIn | Supported | Single-image path supported                         | Implemented and recently fixed, still needs re-verification against the live API               | Thread children are posted as comments | Supported  | Supported |
| Threads  | Supported | Single-media path in current OpenPost composer flow | Implemented with MIME-aware media handling and public media URL requirement                    | `reply_to_id`                          | Supported  | Supported |
| Facebook | Supported | One public HTTPS image URL                          | Implemented for one public HTTPS video URL, needs live-account verification                    | No                                     | Supported  | Supported |
| Instagram | No        | One public HTTPS image URL                          | Implemented for one public HTTPS video URL as Reels, needs live-account verification           | No                                     | Supported  | Supported |
| TikTok   | No        | No                                                  | Implemented for one public HTTPS video URL through direct post, needs live-account verification | No                                     | Supported  | Supported |
| YouTube  | No        | No                                                  | Implemented for one private video upload, needs live-account verification                      | No                                     | Supported  | Supported |

## Planned Platform Adapters

No planned provider adapter is exposed as connectable today. Future provider roadmap items should stay `status: "planned"` until the backend adapter, UI, docs, and tests land together.

## Known Limitations

- **Video support is uneven** — implementation exists across multiple providers, but support is still provider-dependent and some paths need end-to-end verification with real accounts.
- **No full feature parity guarantee** — OpenPost provides the core scheduling features but may not support every platform-specific option (e.g., polls, galleries, stories)
- **Planned providers are discovery-only** — adding a future provider to provider app config fails until its adapter is implemented.
- **Provider APIs can change** — social platforms may change their APIs, rate limits, or app review requirements at any time
- **OAuth tokens require HTTPS** — callbacks need a valid domain with TLS for OAuth to work

## Reading this table correctly

- A provider can support a feature natively while OpenPost still marks it unsupported or unverified.
- "Implemented" means the code path exists in OpenPost.
- "Verified" means the implementation has been confirmed against a live provider account recently.
- Deployment details still matter. Threads, Facebook, Instagram, and TikTok depend on a public media URL, and LinkedIn depends heavily on granted app permissions.

These limits are a starting point, not a permanent contract. Providers can change them.
