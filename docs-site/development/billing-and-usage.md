# Billing And Usage Foundation

OpenPost Cloud billing is built around local entitlement snapshots and durable usage counters. The backend should not call Polar on every API request.

## Current pieces

- `entitlements.Service`: evaluates plan limits and keeps self-hosted defaults unlimited.
- `usage_counters`: monthly durable counters keyed by workspace, metric, and UTC month.
- Workspace creation checks `LimitWorkspaces` before inserting a new workspace.
- Provider connection flows check `social_accounts` before inserting a new active social account.
- Media uploads check `media_bytes_uploaded_monthly` and `media_bytes_stored`; successful new uploads increment monthly uploaded-byte usage.
- Scheduled single posts and threads check `scheduled_posts_monthly` before inserting posts or jobs; successful scheduled creates increment monthly scheduled-post usage.

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

- Publishing worker: `published_posts_monthly` and `provider_write_calls_monthly`
- Team invitations: `team_members`

Polar integration should update subscription state and entitlement snapshots through webhook handlers. API handlers should consume local snapshots only.
