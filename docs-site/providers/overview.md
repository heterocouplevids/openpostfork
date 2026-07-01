# Provider Overview

OAuth and provider app setup are the most common source of deployment friction. Use this section when you are enabling networks one by one.

## Current provider apps

These providers have adapter code in OpenPost today. The Accounts page discovers them through `GET /api/v1/accounts/providers` and shows whether each one is ready to connect on the current server.

| Provider | Auth method            | Server setup                                    | Status       | Notes                                                          |
| -------- | ---------------------- | ----------------------------------------------- | ------------ | -------------------------------------------------------------- |
| Bluesky  | App password           | None                                            | Built-in     | Users connect with handle + app password.                      |
| X        | OAuth 1.0a             | Client ID + secret                              | Configurable | Requires an X developer app with OAuth 1.0a user auth enabled. |
| Mastodon | OAuth 2.0 per instance | Dynamic registration or `MASTODON_SERVERS` JSON | Configurable | One app per instance, unless dynamic registration is enabled.  |
| LinkedIn | OAuth 2.0              | Client ID + secret                              | Configurable | Replies may need extra approval.                               |
| Threads  | Meta OAuth             | Client ID + secret + redirect URI               | Configurable | Public media URL required.                                     |
| Facebook | Meta OAuth             | Structured provider app JSON                    | Configurable | Pages only first slice; public media URL required for media.   |
| TikTok   | OAuth 2.0              | Structured provider app JSON                    | Configurable | Video-only first slice; public media URL required.             |

Start with one provider, confirm the callback works, then expand.

## Planned provider apps

OpenPost now exposes planned providers in the provider discovery API so clients can render a truthful roadmap without enabling broken connect buttons.

| Provider  | Planned focus                                                                        | Status          |
| --------- | ------------------------------------------------------------------------------------ | --------------- |
| Instagram | Images, Reels, scheduling, platform variants, MCP workflows | Planned adapter |
| YouTube   | Shorts and video publishing workflows                       | Planned adapter |

Do not add planned providers to `OPENPOST_PROVIDER_APPS` yet. The backend intentionally rejects unsupported provider app entries until each adapter implements the shared `PlatformAdapter` contract.

## Support matrix

This matrix reflects current OpenPost support, not the full theoretical capability of each provider API.

| Provider | Text posts | Image posts | Threads / replies                | Scheduled posts | Video posts                                                           | Platform-specific variants | Analytics |
| -------- | ---------- | ----------- | -------------------------------- | --------------- | --------------------------------------------------------------------- | -------------------------- | --------- |
| X        | Yes        | Yes         | Yes                              | Yes             | Partial, needs real-account verification                              | Yes                        | No        |
| Mastodon | Yes        | Yes         | Yes                              | Yes             | Partial, needs real-account verification                              | Yes                        | No        |
| Bluesky  | Yes        | Yes         | Yes                              | Yes             | Partial, one MP4 path implemented and needs real-account verification | Yes                        | No        |
| LinkedIn | Yes        | Yes         | Partial, implemented as comments | Yes             | Partial, implementation exists and needs re-verification              | Yes                        | No        |
| Threads  | Yes        | Yes         | Yes                              | Yes             | Partial, public-media deployment dependent                            | Yes                        | No        |
| Facebook | Yes        | Yes         | No                               | Yes             | Partial, one public HTTPS video URL path implemented                  | Yes                        | No        |
| TikTok   | No         | No          | No                               | Yes             | Partial, one public HTTPS video URL path implemented                  | Yes                        | No        |

## Provider-specific caveats

- **X:** Requires an X developer app with OAuth 1.0a user auth enabled and matching callback URLs.
- **Mastodon:** Setup is per instance. Each Mastodon server needs its own app credentials in `MASTODON_SERVERS`.
- **Bluesky:** Uses handle plus app password. No server-side OAuth app is required.
- **LinkedIn:** Permissions and app review can block some publishing or reply workflows even when the integration code is present.
- **Threads:** Media must be reachable at a public `OPENPOST_MEDIA_URL`, and Meta fetches those files server-side.
- **Facebook:** Configure through `OPENPOST_PROVIDER_APPS` with provider `facebook`. The initial adapter connects a selected Page and supports text, one image URL, or one video URL.
- **TikTok:** Configure through `OPENPOST_PROVIDER_APPS` with provider `tiktok`. The initial adapter supports one video attachment via a public HTTPS media URL and the direct-post video endpoint.

Provider API policies, scopes, rate limits, and review requirements can change. Re-check provider docs if a previously working flow starts failing.
