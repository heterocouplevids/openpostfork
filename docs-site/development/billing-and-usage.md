# Billing And Usage Foundation

OpenPost Cloud billing is built around local entitlement snapshots and durable usage counters. The backend should not call Polar on every API request.

## Current pieces

- `entitlements.Service`: evaluates plan limits and keeps self-hosted defaults unlimited.
- `usage_counters`: monthly durable counters keyed by workspace, metric, and UTC month.
- `billing_subscriptions`: local Polar subscription snapshots keyed by workspace.
- `billing_webhook_events`: webhook event ledger for idempotent Polar processing.
- `POST /api/v1/billing/checkout`: creates a Polar checkout for the requested plan and workspace.
- `POST /api/v1/billing/portal`: creates a Polar customer portal session for the workspace.
- `POST /api/v1/billing/polar/webhook`: verifies Standard Webhooks signatures and upserts local subscription state.
- Cloud mode reads `billing_subscriptions.entitlement_snapshot` for workspace-scoped quota checks.
- Workspace creation checks `LimitWorkspaces` before inserting a new workspace.
- Provider connection flows check `social_accounts` before inserting a new active social account.
- Media uploads check `media_bytes_uploaded_monthly` and `media_bytes_stored`; successful new uploads increment monthly uploaded-byte usage.
- Scheduled single posts and threads check `scheduled_posts_monthly` before inserting posts or jobs; successful scheduled creates increment monthly scheduled-post usage.
- The publishing worker records successful published posts and provider publish write calls into monthly usage counters.

## Monthly metrics

Initial metrics match the production-readiness plan:

- `scheduled_posts_monthly`
- `published_posts_monthly`
- `media_bytes_uploaded_monthly`
- `media_bytes_stored`
- `provider_write_calls_monthly`
- `social_accounts`
- `workspaces`
- `team_members`

## Next enforcement points

- Publishing worker enforcement for `published_posts_monthly` and `provider_write_calls_monthly`
- Team invitations: `team_members`

## Polar configuration

Set these only on hosted/cloud deployments:

- `OPENPOST_POLAR_ACCESS_TOKEN`
- `OPENPOST_POLAR_WEBHOOK_SECRET`
- `OPENPOST_POLAR_CHECKOUT_SUCCESS_URL`
- `OPENPOST_POLAR_RETURN_URL`
- `OPENPOST_POLAR_STARTER_PRODUCT_ID`
- `OPENPOST_POLAR_CREATOR_PRODUCT_ID`
- `OPENPOST_POLAR_PRO_PRODUCT_ID`

`OPENPOST_POLAR_RETURN_URL` is the OpenPost app URL Polar should return users to after hosted checkout or customer portal flows, usually the billing settings page. `OPENPOST_POLAR_CUSTOMER_PORTAL_URL` is still accepted as a legacy alias.

Checkout metadata uses primitive `limit_<metric>` keys, for example `limit_scheduled_posts_monthly`, because Polar metadata values are primitive values. The webhook processor rebuilds those keys into the local entitlement snapshot. API handlers consume local snapshots only; they do not call Polar on quota checks.
