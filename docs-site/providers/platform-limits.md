# Supported Platforms & Limitations

OpenPost supports publishing to X, Mastodon, Bluesky, Threads, and LinkedIn. Each platform has different API capabilities and limits.

Provider-native API capabilities are not the same as OpenPost-supported capabilities. The table below reflects the current OpenPost implementation state, including paths that still need real-account verification.

## Current Platform Support

| Provider | Images | Video | Threading |
|---|---|---|---|
| X | 4 images | Implementation exists, manual verification still needed | Replies |
| Mastodon | 4 attachments | Implementation likely works, manual verification still needed | Replies |
| Bluesky | 4 images | Implementation exists for one MP4 video, manual verification still needed | AT Protocol reply refs |
| LinkedIn | 1 image | Implementation exists, manual verification still needed | Comments on first post |
| Threads | 1 media item in OpenPost | Implementation uses stored MIME type and public media URL, manual verification still needed | `reply_to_id` |

## Known Limitations

- **Video support is uneven** — X, Mastodon, and Threads have working video implementations (videos with chunked X upload, Mastodon async media, Threads MIME-based video detection). Bluesky video is newly implemented via `app.bsky.video.uploadVideo`. LinkedIn video had a `fileSizeBytes` type issue (fixed). Provider limits and app-review permissions still apply.
- **No full feature parity guarantee** — OpenPost provides the core scheduling features but may not support every platform-specific option (e.g., polls, galleries, stories)
- **Provider APIs can change** — social platforms may change their APIs, rate limits, or app review requirements at any time
- **OAuth tokens require HTTPS** — callbacks need a valid domain with TLS for OAuth to work

These limits are a starting point, not a permanent contract. Providers can change them.
