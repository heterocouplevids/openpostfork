# Environment Variables

This page summarizes the env vars used by the backend. Some values in `.env.example` are recommended deployment examples; code defaults may differ.

## Core settings

| Variable | Required | Default | Description |
|---|---:|---|---|
| `OPENPOST_PORT` | No | `8080` | HTTP server port. |
| `OPENPOST_DATABASE_PATH` | No | `file:openpost.db?cache=shared&mode=rwc` | SQLite database path or DSN. |
| `OPENPOST_APP_URL` | No, but set it in real deployments | `http://localhost:8080` | Public frontend origin used for CORS and auth flow assumptions. |
| `OPENPOST_PUBLIC_URL` | No | falls back to `OPENPOST_APP_URL` | Canonical browser origin used when configuring WebAuthn/passkeys. Set this to your real app URL in production. |
| `OPENPOST_EXTRA_CORS_ORIGINS` | No | empty | Extra comma-separated origins to allow. |
| `OPENPOST_DISABLE_REGISTRATIONS` | No | `false` | Disables new self-service signups after setup. The first account on a fresh instance is still allowed and becomes the instance admin automatically. |
| `OPENPOST_JWT_SECRET` | Yes | none | Secret used to sign JWTs. Must be at least 32 characters. |
| `OPENPOST_ENCRYPTION_KEY` | Yes | none | Secret used to encrypt stored OAuth tokens. Must be at least 32 characters. |
| `OPENPOST_MEDIA_PATH` | No | `./media` | Local directory for uploaded media. |
| `OPENPOST_MEDIA_URL` | No, but required for Threads production use | `/media` | Public base URL for media files. |
| `OPENPOST_ENV` | No | empty | Optional deployment label. Secret validation is enforced regardless of environment mode. |

## X

| Variable | Required | Default | Description |
|---|---:|---|---|
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

## Notes

- The preferred names above are what new deployments should use.
- Backward-compatible aliases still work for existing installs: `OPENPOST_DB_PATH`, `OPENPOST_FRONTEND_URL`, `OPENPOST_CORS_EXTRA_ORIGINS`, `JWT_SECRET`, `ENCRYPTION_KEY`, `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET`, `TWITTER_REDIRECT_URI`, and `OPENPOST_DISABLE_LINKEDIN_THREAD_REPLIES`.
- The root `.env.example` is the best copy-paste starting point.
- Set explicit public URLs in production even when defaults exist.
- For Threads, treat `OPENPOST_MEDIA_URL` as mandatory.
