# Provider Troubleshooting

Use this page when a provider connection, OAuth callback, media upload, or scheduled publish fails. Start by capturing a support snapshot:

```bash
openpost instance diagnostics \
  --instance https://your-domain.com \
  --deployment docker-compose \
  --provider youtube \
  --logs-file ./openpost.log \
  --json
```

The snapshot includes health/readiness checks, token presence, workspace context, and a redacted last-100-line log tail. It does not print raw tokens or server secrets.

## First Checks

1. Confirm the provider appears as `available` in **Accounts** or through `GET /api/v1/accounts/providers`.
2. Compare the callback URL in the provider console with the OpenPost callback URL exactly.
3. Confirm `OPENPOST_APP_URL` is the public HTTPS app origin.
4. For media providers that fetch files server-side, confirm `OPENPOST_MEDIA_URL` or `OPENPOST_S3_PUBLIC_BASE_URL` is public HTTPS.
5. Open the failed post and inspect each destination error message.
6. Check logs around the callback or scheduled publish time.

## Common Symptoms

| Symptom                                | Likely cause                                                            | Fix                                                                        |
| -------------------------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Provider is unavailable on Accounts    | Missing provider app config                                             | Add the provider env vars or `OPENPOST_PROVIDER_APPS` entry, then restart. |
| OAuth redirects to the wrong host      | `OPENPOST_APP_URL` or provider callback mismatch                        | Set one public HTTPS origin and update the provider console.               |
| OAuth succeeds but no account is saved | Provider returned no usable profile, page, channel, or business account | Confirm scopes, account ownership, and provider app review state.          |
| Text publishes but media fails         | Media URL is private, local, or not HTTPS                               | Use public `OPENPOST_MEDIA_URL` or S3/R2 public media URLs.                |
| Scheduled post fails later             | Token expired, revoked, or provider rejected the payload                | Reconnect the account and retry with provider-compatible media.            |
| Provider returns permission errors     | App lacks product access, scopes, or review approval                    | Enable the product and request the listed scopes in the provider console.  |

## X

- Connection requires OAuth 1.0a user authentication. OAuth 2-only apps will not work with the current adapter.
- Callback must match `https://your-domain.com/api/v1/accounts/x/callback` unless `X_REDIRECT_URI` overrides it.
- Media support needs OAuth 1.0a token + secret pairs. Reconnect old accounts if media uploads fail with an OAuth 1.0a reconnect message.
- Video is safest as one MP4 attachment and cannot be mixed with images.

## Mastodon

- Custom instances must be public HTTPS and must allow app registration.
- OpenPost rejects private, loopback, link-local, multicast, and local-address instance hosts.
- Preconfigured instances must preserve the exact `instance_url`; the persisted provider key is `mastodon:<instance_url>`.
- Video support is instance-dependent. MP4 and WebM are the safest formats.

## Bluesky

- Bluesky uses handle + app password, not OAuth.
- Use an app password, not the account's main password.
- Image posts support up to four images.
- Video posts require one MP4 video under 100MB. Video cannot be mixed with images.

## LinkedIn

- Callback must match `https://your-domain.com/api/v1/accounts/linkedin/callback` unless `LINKEDIN_REDIRECT_URI` overrides it.
- LinkedIn permissions and app review can block publishing or social-action/comment workflows even when OAuth succeeds.
- If thread child posts fail, set `LINKEDIN_DISABLE_THREAD_REPLIES=true` until the app has the required comment permissions.
- Video upload uses LinkedIn's Videos API and still needs live-account re-verification before broad production claims.

## Threads

- Threads requires the Meta app's Threads product and scopes: `threads_basic`, `threads_content_publish`, and `threads_manage_replies`.
- Media URLs must be public HTTPS. Meta fetches media server-side and cannot use localhost, private DNS, or plain local paths.
- For local testing, expose both the app callback and `/media/...` paths through a tunnel.
- Threads supports one attachment in the current OpenPost flow. Videos must be MP4 or MOV.

## Facebook

- Facebook connects Pages, not personal profile timelines.
- `facebook account has no manageable pages` usually means the user has no eligible Pages or the app lacks `pages_show_list`.
- Page publishing requires `pages_show_list`, `pages_read_engagement`, and `pages_manage_posts`, often with Meta app review.
- Media posts require one public HTTPS image or video URL.

## Instagram

- Instagram requires an Instagram Business or Creator account connected to a Facebook Page.
- `facebook account has no connected instagram business accounts` means the authenticated Meta user has no eligible Page-backed Instagram account or the app lacks required scopes.
- Publishing requires `instagram_basic`, `instagram_content_publish`, Page scopes, and often Meta app review.
- Instagram does not publish text-only posts. Attach exactly one image or one Reel video.

## TikTok

- TikTok requires Login Kit plus Content Posting API access.
- Required scopes are `user.info.basic`, `user.info.profile`, `video.publish`, and `video.upload`.
- The redirect URI in the TikTok app must match `https://your-domain.com/api/v1/accounts/tiktok/callback` or the configured `redirect_uri`.
- The current adapter publishes one video through a public HTTPS media URL. Text-only and image posts are not enabled.

## YouTube

- Enable YouTube Data API v3 in the Google Cloud project.
- The OAuth app needs profile/email scopes plus `youtube.readonly` and `youtube.upload`.
- `google account has no YouTube channels` means the authenticated Google account has no eligible channel available to the OAuth app.
- `invalidTitle` usually means the first non-empty line of the post or YouTube variant is invalid for a video title.
- Uploads are private by default and support one video attachment in the current adapter.

## Escalation Checklist

Before filing an issue or escalating an operator incident, include:

- Output from `openpost instance diagnostics --provider <provider> --logs-file <path> --json`
- Provider name and account type being tested
- Deployment method and public app/media URLs
- Exact callback URL configured in the provider console
- Failed post destination error, if the account connection succeeded
- Whether the same account can publish text-only content
