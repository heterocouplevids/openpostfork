# Environment Variables

This page summarizes the env vars used by the backend. Some values in `.env.example` are recommended deployment examples; code defaults may differ.

## Core settings

| Variable | Required | Default | Description |
|---|---:|---|---|
| `OPENPOST_PORT` | No | `8080` | HTTP server port. |
| `OPENPOST_EDITION` | No | `selfhost` | Product edition. Valid values: `selfhost`, `cloud`. Cloud mode enforces hosted storage requirements at startup. |
| `OPENPOST_DATABASE_DRIVER` | Required as `postgres` in cloud mode | `sqlite` | Database driver. Valid values: `sqlite`, `postgres`. SQLite remains the self-hosted default. |
| `OPENPOST_DATABASE_PATH` | No | `file:openpost.db?cache=shared&mode=rwc` | SQLite database path or DSN. Also acts as a legacy fallback DSN for Postgres if `OPENPOST_DATABASE_URL` is unset. |
| `OPENPOST_DATABASE_URL` | Required for Postgres and cloud mode | empty | Postgres connection URL used when `OPENPOST_DATABASE_DRIVER=postgres`. |
| `OPENPOST_APP_URL` | No, but set it in real deployments | `http://localhost:8080` | Public frontend origin used for CORS and auth flow assumptions. |
| `OPENPOST_PUBLIC_URL` | No | falls back to `OPENPOST_APP_URL` | Canonical browser origin used when configuring WebAuthn/passkeys. Set this to your real app URL in production. |
| `OPENPOST_EXTRA_CORS_ORIGINS` | No | empty | Extra comma-separated origins to allow. Cloud mode allows only `OPENPOST_APP_URL` plus these explicit origins and rejects `*`. |
| `OPENPOST_DISABLE_REGISTRATIONS` | No | `false` | Disables new self-service signups after setup. The first account on a fresh instance is still allowed and becomes the instance admin automatically. |
| `OPENPOST_JWT_SECRET` | Yes | none | Secret used to sign JWTs. Must be at least 32 characters. |
| `OPENPOST_ENCRYPTION_KEY` | Yes | none | Secret used to encrypt stored OAuth tokens. Must be at least 32 characters. |
| `OPENPOST_STORAGE_DRIVER` | Required as `s3` in cloud mode | `local` | Media storage driver. Valid values: `local`, `s3`. |
| `OPENPOST_MEDIA_PATH` | No | `./media` | Local directory for uploaded media. |
| `OPENPOST_MEDIA_URL` | No, but required for public-URL provider media publishing | `/media` | Public base URL for media files. |
| `OPENPOST_S3_ENDPOINT` | Required for R2 or non-AWS S3-compatible storage | empty | S3-compatible API endpoint. Native AWS S3 can leave this empty. |
| `OPENPOST_S3_REGION` | Required for S3-compatible storage and cloud mode | empty | S3 region. R2 commonly uses `auto`. |
| `OPENPOST_S3_BUCKET` | Required for S3-compatible storage and cloud mode | empty | Bucket name for uploaded media. |
| `OPENPOST_S3_ACCESS_KEY_ID` | Required for S3-compatible storage and cloud mode | empty | S3 access key ID. |
| `OPENPOST_S3_SECRET_ACCESS_KEY` | Required for S3-compatible storage and cloud mode | empty | S3 secret access key. |
| `OPENPOST_S3_PUBLIC_BASE_URL` | Required in cloud mode | empty | Public media base URL for provider fetches and preview links. |
| `OPENPOST_S3_FORCE_PATH_STYLE` | No | `false` | Force path-style S3 addressing for compatible providers that require it. |
| `OPENPOST_POLAR_ACCESS_TOKEN` | Required in cloud mode | empty | Polar API access token for hosted checkout and customer portal sessions. |
| `OPENPOST_POLAR_API_BASE_URL` | No | `https://api.polar.sh/v1` | Polar API base URL. Use `https://sandbox-api.polar.sh/v1` for sandbox testing. |
| `OPENPOST_POLAR_WEBHOOK_SECRET` | Required in cloud mode | empty | Polar Standard Webhooks secret used to verify billing events. |
| `OPENPOST_POLAR_CHECKOUT_SUCCESS_URL` | Required in cloud mode | empty | Browser return URL after a successful checkout. |
| `OPENPOST_POLAR_RETURN_URL` | Required in cloud mode | empty | Browser return URL for customer portal sessions. |
| `OPENPOST_POLAR_STARTER_PRODUCT_ID` | Required in cloud mode | empty | Polar product ID for the Starter plan. |
| `OPENPOST_POLAR_CREATOR_PRODUCT_ID` | Required in cloud mode | empty | Polar product ID for the Creator plan. |
| `OPENPOST_POLAR_PRO_PRODUCT_ID` | Required in cloud mode | empty | Polar product ID for the Pro plan. |
| `OPENPOST_ENV` | No | empty | Optional deployment label. Secret validation is enforced regardless of environment mode. |

## Provider app registry

OpenPost builds provider adapters at startup from legacy provider env vars, optional `OPENPOST_PROVIDER_APPS` JSON, and active rows in the `provider_apps` database table.

Database rows are intended for hosted/operator-managed installs. They store `client_secret_encrypted` with the same `OPENPOST_ENCRYPTION_KEY` used for account tokens, override matching env/JSON entries, and require a server restart after changes until hot reload exists. Matching is by provider, except Mastodon uses provider plus `instance_url`.

Instance admins can manage encrypted database rows through `GET /api/v1/admin/provider-apps`, `POST /api/v1/admin/provider-apps`, and `DELETE /api/v1/admin/provider-apps/{id}`. API responses never return client secrets; send `client_secret` only when creating a row or rotating the existing secret.

The web app also exposes this registry in **Settings -> Provider Apps** for instance admins. Saves and deletes take effect after the next OpenPost server restart.

## X

| Variable | Required | Default | Description |
|---|---:|---|---|
| `OPENPOST_PROVIDER_APPS` | No | empty | Structured JSON provider app registry. Entries override matching legacy env providers; active `provider_apps` database rows override matching JSON entries. |
| `X_CLIENT_ID` | Yes for X | empty | X OAuth client ID. Leave empty to disable X. |
| `X_CLIENT_SECRET` | Yes for X | empty | X OAuth client secret. |
| `X_REDIRECT_URI` | No | derived from `OPENPOST_APP_URL` | X OAuth callback URL override. |

## Mastodon

| Variable | Required | Default | Description |
|---|---:|---|---|
| `MASTODON_REDIRECT_URI` | No | `urn:ietf:wg:oauth:2.0:oob` | Mastodon redirect URI. The default uses the OOB flow and does not need a public callback URL. |
| `MASTODON_SERVERS` | Yes for Mastodon | `[]` | JSON array of configured Mastodon apps and instance URLs. Leave empty to disable Mastodon. |

## LinkedIn

| Variable | Required | Default | Description |
|---|---:|---|---|
| `LINKEDIN_CLIENT_ID` | Yes for LinkedIn | empty | LinkedIn OAuth client ID. Leave empty to disable LinkedIn. |
| `LINKEDIN_CLIENT_SECRET` | Yes for LinkedIn | empty | LinkedIn OAuth client secret. |
| `LINKEDIN_REDIRECT_URI` | No | derived from `OPENPOST_APP_URL` | LinkedIn callback URL override. |
| `LINKEDIN_DISABLE_THREAD_REPLIES` | No | `false` | Disable LinkedIn comment-style child replies for thread posts. |

## Threads

| Variable | Required | Default | Description |
|---|---:|---|---|
| `THREADS_CLIENT_ID` | Yes for Threads | empty | Meta app ID. Leave empty to disable Threads. |
| `THREADS_CLIENT_SECRET` | Yes for Threads | empty | Meta app secret. |
| `THREADS_REDIRECT_URI` | No | derived from `OPENPOST_APP_URL` | Threads callback URL override. Threads production redirects must use HTTPS. |
| `META_GRAPH_API_VERSION` | No | `v25.0` | Meta Graph API version used by the Facebook Pages and Instagram adapters. |

## Facebook

Facebook Pages publishing is configured through the provider app registry instead of legacy provider-specific env vars. Use `OPENPOST_PROVIDER_APPS` for bootstrap/self-hosting or the instance-admin provider app API for hosted/operator-managed credentials.

Example:

```json
[
  {
    "provider": "facebook",
    "client_id": "your-meta-app-id",
    "client_secret": "your-meta-app-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives `https://your-domain.com/api/v1/accounts/facebook/callback` from `OPENPOST_APP_URL`. Facebook media publishing requires `OPENPOST_MEDIA_URL` or `OPENPOST_S3_PUBLIC_BASE_URL` to point at public HTTPS media URLs.

## Instagram

Instagram Business publishing is configured through the provider app registry instead of legacy provider-specific env vars. Use `OPENPOST_PROVIDER_APPS` for bootstrap/self-hosting or the instance-admin provider app API for hosted/operator-managed credentials.

Example:

```json
[
  {
    "provider": "instagram",
    "client_id": "your-meta-app-id",
    "client_secret": "your-meta-app-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives `https://your-domain.com/api/v1/accounts/instagram/callback` from `OPENPOST_APP_URL`. Instagram media publishing requires `OPENPOST_MEDIA_URL` or `OPENPOST_S3_PUBLIC_BASE_URL` to point at public HTTPS media URLs.

## TikTok

TikTok is configured through the provider app registry instead of legacy provider-specific env vars. Use `OPENPOST_PROVIDER_APPS` for bootstrap/self-hosting or the instance-admin provider app API for hosted/operator-managed credentials.

Example:

```json
[
  {
    "provider": "tiktok",
    "client_id": "your-client-key",
    "client_secret": "your-client-secret",
    "redirect_uri": "https://your-domain.com/api/v1/accounts/tiktok/callback"
  }
]
```

TikTok direct video publishing requires `OPENPOST_MEDIA_URL` or `OPENPOST_S3_PUBLIC_BASE_URL` to point at public HTTPS media URLs.

## YouTube

YouTube video uploads are configured through the provider app registry instead of legacy provider-specific env vars. Use `OPENPOST_PROVIDER_APPS` for bootstrap/self-hosting or the instance-admin provider app API for hosted/operator-managed credentials.

Example:

```json
[
  {
    "provider": "youtube",
    "client_id": "your-google-oauth-client-id",
    "client_secret": "your-google-oauth-client-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives `https://your-domain.com/api/v1/accounts/youtube/callback` from `OPENPOST_APP_URL`. The first adapter slice uploads one video as a private YouTube video and derives the video title from the first non-empty line of the post or platform variant.

## Notes

- The preferred names above are what new deployments should use.
- `OPENPOST_PROVIDER_APPS` accepts an array of objects with `provider`, `name`, `client_id`, `client_secret`, `redirect_uri`, and `instance_url`. The `provider_apps` table stores the same logical fields, with `client_secret` encrypted into `client_secret_encrypted`. Both currently support implemented adapters only: `x`, `mastodon`, `linkedin`, `threads`, `facebook`, `instagram`, `tiktok`, and `youtube`; Bluesky is enabled separately through app-password login.
- Backward-compatible aliases still work for existing installs: `OPENPOST_DB_PATH`, `OPENPOST_FRONTEND_URL`, `OPENPOST_CORS_EXTRA_ORIGINS`, `JWT_SECRET`, `ENCRYPTION_KEY`, `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET`, `TWITTER_REDIRECT_URI`, and `OPENPOST_DISABLE_LINKEDIN_THREAD_REPLIES`.
- The root `.env.example` is the best copy-paste starting point.
- Set explicit public URLs in production even when defaults exist.
- For Threads, Facebook, Instagram, and TikTok, treat `OPENPOST_MEDIA_URL` as mandatory unless S3/R2 public media URLs are configured.
