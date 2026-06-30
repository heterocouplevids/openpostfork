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
protected-resource and authorization-server metadata:

```txt
GET /.well-known/oauth-protected-resource
GET /.well-known/oauth-authorization-server
```

Tool descriptors include OAuth security schemes, mirrored `_meta.securitySchemes`,
and tool annotations so clients can distinguish read-only tools from actions that
write OpenPost state or reach external URLs.

OAuth-aware clients can start account linking at the browser authorization page,
then exchange the returned code for an MCP-scoped bearer token:

```txt
GET /oauth/authorize
POST /oauth/token
```

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
- `list_drafts`: returns editable draft posts for a workspace so an assistant can inspect existing work before creating more drafts.
- `update_draft`: updates a draft's source content and optionally replaces destination accounts.
- `set_post_renditions`: creates or updates destination-specific copy for draft and scheduled posts. Renditions can only target accounts already attached as post destinations.
- `schedule_post`: creates a scheduled post with destination accounts and queues the `publish_post` job.
- `schedule_draft`: schedules an existing draft and queues the `publish_post` job without duplicating the post.
- `get_post_status`: returns the post status, scheduled run time, and per-destination status.
- `list_scheduled_posts`: returns upcoming scheduled posts for queue inspection.
- `cancel_post`: cancels a queued scheduled post and returns it to drafts.
- `suggest_next_slot`: returns the next free configured posting slot for a workspace.
- `upload_media_from_url`: fetches a public HTTP(S) media URL and stores it in a workspace.

## Current prompts

- `plan_social_post`: guides an assistant from a rough idea to a workspace-aware draft.
- `adapt_platform_renditions`: guides destination-specific copywriting for an existing draft or scheduled post.
- `review_schedule`: guides queue inspection and next-action recommendations without mutating posts.

## Current scope

- Uses the same Bearer authentication path as the CLI and API tokens.
- Dedicated `mcp:full` tokens can be created in Settings for ChatGPT, Claude, and other MCP clients. Existing `cli:full` tokens also remain accepted by `/mcp` so `openpost-mcp` profiles continue to work.
- Publishes MCP protected-resource metadata and returns `WWW-Authenticate` plus `_meta["mcp/www_authenticate"]` challenges for unauthenticated MCP requests.
- Publishes OAuth authorization-server metadata for public PKCE clients, including `S256`, `mcp:full`, and client ID metadata document support.
- Provides a browser approval page at `/oauth/authorize` and a form-encoded `/oauth/token` code exchange that mints `mcp:full` API tokens.
- Validates client metadata redirect URIs for URL-based client IDs, accepts ChatGPT fallback redirects for predefined clients, and binds OAuth-issued MCP tokens to the `/mcp` resource audience.
- Advertises and enforces the `mcp:full` OAuth scope in every MCP tool descriptor. Fine-grained per-session scopes are still planned.
- Provides `openpost-mcp` for local stdio clients without duplicating server tool logic.
- Advertises MCP prompt templates for common agentic scheduling workflows: planning a post, adapting platform renditions, and reviewing the publishing queue.
- Validates workspace membership and account ownership before returning, creating, scheduling, canceling, or uploading data.
- Keeps draft iteration agent-friendly: assistants can list drafts, update draft copy/destinations, set per-destination renditions, and schedule the same draft when it is ready.
- Validates rendition targets against the post destination list so assistants do not create variants that would never publish.
- Rejects media URL fetches that resolve to private, loopback, link-local, multicast, or otherwise local addresses.
- Enforces the same scheduled-post and media-upload entitlement and usage accounting as the web/API paths.
- Records MCP tool calls in `mcp_tool_calls` with user, workspace, tool name, success/error status, error message, duration, and timestamp, and exposes recent calls in settings.
- Records API-token client ID, name, scope, and token prefix for MCP tool calls when a request uses a dedicated CLI/MCP token, so Settings can attribute activity to ChatGPT, Claude, CI, or another configured client.
- Returns structured content so assistants can inspect workspace, account, post, destination, media, and suggested slot IDs without parsing prose.
