# Supported Platforms & Limitations

OpenPost supports publishing to X, Mastodon, Bluesky, Threads, and LinkedIn. Each platform has different API capabilities and limits.

## Current Platform Support

| Provider | Images | Video | Threading |
|---|---|---|---|
| X | 4 images | Not yet supported | Replies |
| Mastodon | 4 attachments | Not yet supported | Replies |
| Bluesky | 4 images | Not yet supported | AT Protocol reply refs |
| LinkedIn | 1 image | Not yet supported | Comments on first post |
| Threads | 1 image | Not yet supported | `reply_to_id` |

## Known Limitations

- **No video support yet** — video uploads are not implemented for any platform
- **No full feature parity guarantee** — OpenPost provides the core scheduling features but may not support every platform-specific option (e.g., polls, galleries, stories)
- **Provider APIs can change** — social platforms may change their APIs, rate limits, or app review requirements at any time
- **OAuth tokens require HTTPS** — callbacks need a valid domain with TLS for OAuth to work

These limits are a starting point, not a permanent contract. Providers can change them.
