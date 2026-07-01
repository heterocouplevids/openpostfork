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
| Facebook | Meta OAuth             | Provider app registry                           | Configurable | Pages only first slice; public media URL required for media.   |
| Instagram | Meta OAuth            | Provider app registry                           | Configurable | Business accounts only; public media URL required for media.   |
| TikTok   | OAuth 2.0              | Provider app registry                           | Configurable | Video-only first slice; public media URL required.             |
| YouTube  | Google OAuth           | Provider app registry                           | Configurable | One-video private upload first slice.                           |

Start with one provider, confirm the callback works, then expand.

Provider app credentials can come from legacy env vars, `OPENPOST_PROVIDER_APPS` JSON, or active encrypted `provider_apps` database rows managed through the instance-admin provider app API. Instance admins can manage database rows from **Settings -> Admin -> Provider Apps**.

Mastodon is the most common reason to use Provider Apps because each instance can need its own app registration. Users can still connect public custom Mastodon instances from the Accounts screen when dynamic registration is available. For other OAuth providers, Provider Apps are mainly for operators who want to bring their own keys instead of relying on hosted/default credentials.

Database rows are loaded at startup and override matching env/JSON entries, so operator-managed changes require a restart until hot reload exists.

If connection or publishing fails, use [Provider Troubleshooting](/providers/troubleshooting) to collect diagnostics and map common OAuth, permission, media URL, and publishing errors to the right fix.

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
| Instagram | No        | Yes         | No                               | Yes             | Partial, one public HTTPS video URL path implemented as Reels         | Yes                        | No        |
| TikTok   | No         | No          | No                               | Yes             | Partial, one public HTTPS video URL path implemented                  | Yes                        | No        |
| YouTube  | No         | No          | No                               | Yes             | Partial, one private video upload path implemented                    | Yes                        | No        |

## Provider-specific caveats

- **X:** Requires an X developer app with OAuth 1.0a user auth enabled and matching callback URLs.
- **Mastodon:** Setup is per instance. Custom public instances can be entered from Accounts; operator-pinned instances can use `MASTODON_SERVERS` or **Settings -> Admin -> Provider Apps**.
- **Bluesky:** Uses handle plus app password. No server-side OAuth app is required.
- **LinkedIn:** Permissions and app review can block some publishing or reply workflows even when the integration code is present.
- **Threads:** Media must be reachable at a public `OPENPOST_MEDIA_URL`, and Meta fetches those files server-side.
- **Facebook:** Configure through the provider app registry with provider `facebook`. The initial adapter connects a selected Page and supports text, one image URL, or one video URL.
- **Instagram:** Configure through the provider app registry with provider `instagram`. The initial adapter connects a selected Instagram Business account behind a Facebook Page and supports one image URL or one Reel video URL.
- **TikTok:** Configure through the provider app registry with provider `tiktok`. The initial adapter supports one video attachment via a public HTTPS media URL and the direct-post video endpoint.
- **YouTube:** Configure through the provider app registry with provider `youtube`. The initial adapter connects a selected channel and uploads one video as private by default.

Provider API policies, scopes, rate limits, and review requirements can change. Re-check provider docs if a previously working flow starts failing.
