# MCP And ChatGPT App

OpenPost exposes an authenticated MCP foundation at:

```txt
POST /mcp
```

The endpoint is JSON-RPC over HTTP and requires a bearer token:

```http
Authorization: Bearer <jwt-or-api-token>
```

For ChatGPT Apps and other OAuth-aware MCP clients, OpenPost also publishes
protected-resource metadata:

```txt
GET /.well-known/oauth-protected-resource
```

Tool descriptors include OAuth security schemes, mirrored `_meta.securitySchemes`,
and tool annotations so clients can distinguish read-only tools from actions that
write OpenPost state or reach external URLs.

Desktop MCP clients can use the local stdio proxy from the CLI module:

```sh
openpost-mcp --profile local
```

The proxy loads the same OpenPost CLI profile and token, then forwards MCP
JSON-RPC frames to the remote `/mcp` endpoint.

Recent MCP tool calls are available in Settings under **CLI Devices & API
Tokens**. The same data is exposed to authenticated API clients at:

```txt
GET /api/v1/mcp/activity?limit=20
GET /api/v1/mcp/activity?workspace_id=<workspace-id>
```

## Current tools

- `list_workspaces`: returns the workspaces available to the authenticated user.
- `list_accounts`: returns active social accounts for a workspace.
- `create_draft`: creates a draft post in a workspace, optionally assigned to destination accounts.
- `schedule_post`: creates a scheduled post with destination accounts and queues the `publish_post` job.
- `get_post_status`: returns the post status, scheduled run time, and per-destination status.
- `cancel_post`: cancels a queued scheduled post and returns it to drafts.
- `suggest_next_slot`: returns the next free configured posting slot for a workspace.
- `upload_media_from_url`: fetches a public HTTP(S) media URL and stores it in a workspace.

## Current scope

- Uses the same Bearer authentication path as the CLI and API tokens.
- Dedicated `mcp:full` tokens can be created in Settings for ChatGPT, Claude, and other MCP clients. Existing `cli:full` tokens also remain accepted by `/mcp` so `openpost-mcp` profiles continue to work.
- Publishes MCP protected-resource metadata and returns `WWW-Authenticate` plus `_meta["mcp/www_authenticate"]` challenges for unauthenticated MCP requests.
- Advertises and enforces the `mcp:full` OAuth scope in every MCP tool descriptor. A dedicated OAuth authorization-server flow for ChatGPT account linking is still planned.
- Provides `openpost-mcp` for local stdio clients without duplicating server tool logic.
- Validates workspace membership and account ownership before returning, creating, scheduling, canceling, or uploading data.
- Rejects media URL fetches that resolve to private, loopback, link-local, multicast, or otherwise local addresses.
- Enforces the same scheduled-post and media-upload entitlement and usage accounting as the web/API paths.
- Records MCP tool calls in `mcp_tool_calls` with user, workspace, tool name, success/error status, error message, duration, and timestamp, and exposes recent calls in settings.
- Returns structured content so assistants can inspect workspace, account, post, destination, media, and suggested slot IDs without parsing prose.
