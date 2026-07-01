# Environment Variables

This is the low-level quick reference version of the configuration docs.

| Variable | Purpose |
|---|---|
| `OPENPOST_PORT` | Backend port |
| `OPENPOST_EDITION` | Product edition: `selfhost` or `cloud` |
| `OPENPOST_DATABASE_DRIVER` | Database driver: `sqlite` or `postgres` |
| `OPENPOST_DATABASE_PATH` | SQLite path or DSN |
| `OPENPOST_DATABASE_URL` | Postgres URL when using the Postgres driver |
| `OPENPOST_APP_URL` | Public frontend URL |
| `OPENPOST_PUBLIC_URL` | Canonical browser origin used for WebAuthn/passkeys |
| `OPENPOST_EXTRA_CORS_ORIGINS` | Extra CORS allowlist |
| `OPENPOST_DISABLE_REGISTRATIONS` | Disable new signups after bootstrap |
| `OPENPOST_JWT_SECRET` | JWT signing secret |
| `OPENPOST_ENCRYPTION_KEY` | OAuth token encryption secret |
| `OPENPOST_STORAGE_DRIVER` | Media storage driver: `local` or `s3` |
| `OPENPOST_MEDIA_PATH` | Local media directory |
| `OPENPOST_MEDIA_URL` | Public media base URL |
| `OPENPOST_S3_ENDPOINT` | S3-compatible endpoint for R2 or non-AWS storage |
| `OPENPOST_S3_REGION` | S3 region |
| `OPENPOST_S3_BUCKET` | S3 bucket |
| `OPENPOST_S3_ACCESS_KEY_ID` | S3 access key ID |
| `OPENPOST_S3_SECRET_ACCESS_KEY` | S3 secret access key |
| `OPENPOST_S3_PUBLIC_BASE_URL` | Public media base URL for S3-backed media |
| `OPENPOST_S3_FORCE_PATH_STYLE` | Force path-style S3 addressing |
| `OPENPOST_POLAR_ACCESS_TOKEN` | Polar API access token |
| `OPENPOST_POLAR_API_BASE_URL` | Polar API base URL |
| `OPENPOST_POLAR_WEBHOOK_SECRET` | Polar webhook verification secret |
| `OPENPOST_POLAR_CHECKOUT_SUCCESS_URL` | Polar checkout success URL |
| `OPENPOST_POLAR_RETURN_URL` | Polar customer portal return URL |
| `OPENPOST_POLAR_STARTER_PRODUCT_ID` | Polar Starter product ID |
| `OPENPOST_POLAR_CREATOR_PRODUCT_ID` | Polar Creator product ID |
| `OPENPOST_POLAR_PRO_PRODUCT_ID` | Polar Pro product ID |
| `OPENPOST_PROVIDER_APPS` | Structured provider app registry JSON |
| `X_CLIENT_ID` | X client ID |
| `X_CLIENT_SECRET` | X client secret |
| `X_REDIRECT_URI` | X callback override |
| `MASTODON_REDIRECT_URI` | Mastodon callback override |
| `MASTODON_SERVERS` | Mastodon server JSON |
| `LINKEDIN_CLIENT_ID` | LinkedIn client ID |
| `LINKEDIN_CLIENT_SECRET` | LinkedIn client secret |
| `LINKEDIN_REDIRECT_URI` | LinkedIn callback override |
| `LINKEDIN_DISABLE_THREAD_REPLIES` | Disable LinkedIn thread replies |
| `THREADS_CLIENT_ID` | Threads client ID |
| `THREADS_CLIENT_SECRET` | Threads client secret |
| `THREADS_REDIRECT_URI` | Threads callback override |
| `META_GRAPH_API_VERSION` | Meta Graph API version for Facebook Pages |

Facebook and TikTok are configured through `OPENPOST_PROVIDER_APPS` with providers `facebook` and `tiktok`; no legacy env vars are required.

Legacy aliases still work for upgrades: `OPENPOST_DB_PATH`, `OPENPOST_FRONTEND_URL`, `OPENPOST_CORS_EXTRA_ORIGINS`, `JWT_SECRET`, `ENCRYPTION_KEY`, `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET`, `TWITTER_REDIRECT_URI`, and `OPENPOST_DISABLE_LINKEDIN_THREAD_REPLIES`.
