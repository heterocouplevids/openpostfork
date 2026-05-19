# Provider Overview

OAuth and provider app setup are the most common source of deployment friction. Use this section when you are enabling networks one by one.

| Provider | Auth method | Server setup | Notes |
|---|---|---|---|
| X | OAuth 1.0a | Client ID + secret | Requires an X developer app with OAuth 1.0a user auth enabled. |
| Mastodon | OAuth 2.0 per instance | `MASTODON_SERVERS` JSON | One app per instance. |
| Bluesky | App password | None | Users connect with handle + app password. |
| LinkedIn | OAuth 2.0 | Client ID + secret | Replies may need extra approval. |
| Threads | Meta OAuth | Client ID + secret + redirect URI | Public media URL required. |

Start with one provider, confirm the callback works, then expand.

## Support matrix

This matrix reflects current OpenPost support, not the full theoretical capability of each provider API.

| Provider | Text posts | Image posts | Threads / replies | Scheduled posts | Video posts | Platform-specific variants | Analytics |
|---|---|---|---|---|---|---|---|
| X | Yes | Yes | Yes | Yes | Partial, needs real-account verification | Yes | No |
| Mastodon | Yes | Yes | Yes | Yes | Partial, needs real-account verification | Yes | No |
| Bluesky | Yes | Yes | Yes | Yes | Partial, one MP4 path implemented and needs real-account verification | Yes | No |
| LinkedIn | Yes | Yes | Partial, implemented as comments | Yes | Partial, implementation exists and needs re-verification | Yes | No |
| Threads | Yes | Yes | Yes | Yes | Partial, public-media deployment dependent | Yes | No |

## Provider-specific caveats

- **X:** Requires an X developer app with OAuth 1.0a user auth enabled and matching callback URLs.
- **Mastodon:** Setup is per instance. Each Mastodon server needs its own app credentials in `MASTODON_SERVERS`.
- **Bluesky:** Uses handle plus app password. No server-side OAuth app is required.
- **LinkedIn:** Permissions and app review can block some publishing or reply workflows even when the integration code is present.
- **Threads:** Media must be reachable at a public `OPENPOST_MEDIA_URL`, and Meta fetches those files server-side.

Provider API policies, scopes, rate limits, and review requirements can change. Re-check provider docs if a previously working flow starts failing.
