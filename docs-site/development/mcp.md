# MCP And ChatGPT App

OpenPost exposes an authenticated MCP foundation at:

```txt
POST /mcp
```

The endpoint is JSON-RPC over HTTP and requires:

```http
Authorization: Bearer <jwt-or-api-token>
```

## Current tools

- `list_workspaces`: returns the workspaces available to the authenticated user.
- `list_accounts`: returns active social accounts for a workspace.
- `create_draft`: creates a draft post in a workspace, optionally assigned to destination accounts.
- `schedule_post`: creates a scheduled post with destination accounts and queues the `publish_post` job.
- `get_post_status`: returns the post status, scheduled run time, and per-destination status.
- `cancel_post`: cancels a queued scheduled post and returns it to drafts.
- `suggest_next_slot`: returns the next free configured posting slot for a workspace.

## Current scope

- Uses the same Bearer authentication path as the CLI and API tokens.
- Validates workspace membership and account ownership before returning, creating, scheduling, or canceling data.
- Enforces the same scheduled-post entitlement and usage accounting as the web/API post creation path.
- Returns structured content so assistants can inspect workspace, account, post, destination, and suggested slot IDs without parsing prose.

## Next tools

- `upload_media_from_url`
