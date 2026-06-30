# MCP And ChatGPT App

OpenPost exposes an authenticated MCP foundation at:

```txt
POST /mcp
```

The endpoint is JSON-RPC over HTTP and requires a bearer token:

```http
Authorization: Bearer <jwt-or-api-token>
```

OpenPost accepts MCP `ping` requests and Streamable HTTP JSON-RPC
notifications. Notification POSTs such as `notifications/initialized` return
HTTP `202 Accepted` with no response body.

ChatGPT Apps-compatible clients can also discover and load the scheduler widget
resource:

```txt
resources/list
resources/read ui://widget/openpost-scheduler-v1.html
```

The widget is a self-contained `text/html;profile=mcp-app` resource. The
read-only `render_scheduler_widget` tool points at that resource through
`_meta.ui.resourceUri` and `_meta["openai/outputTemplate"]`, then passes
structured OpenPost data into the widget for rendering.

For ChatGPT Apps and other OAuth-aware MCP clients, OpenPost also publishes
protected-resource and authorization-server metadata:

```txt
GET /.well-known/oauth-protected-resource
GET /.well-known/oauth-authorization-server
```

Tool descriptors include OAuth security schemes, mirrored `_meta.securitySchemes`,
and tool annotations so clients can distinguish read-only tools from actions that
write OpenPost state or reach external URLs. They also include short Apps SDK
invocation status labels and output schemas for the returned `structuredContent`.

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
- `list_provider_catalog`: returns provider launch status so assistants know which platforms are available, need server configuration, or are still planned.
- `list_publications`: returns source publications for a workspace, optionally filtered by status.
- `list_accounts`: returns active social accounts for a workspace.
- `list_media`: returns recent workspace media attachments so assistants can reuse existing assets.
- `create_publication`: creates a source publication from an idea, link, goal, audience, and optional media before platform drafts are created.
- `create_draft`: creates a draft post in a workspace, optionally assigned to destination accounts and media attachments.
- `list_drafts`: returns editable draft posts for a workspace so an assistant can inspect existing work before creating more drafts.
- `update_draft`: updates a draft's source content and optionally replaces destination accounts or source media.
- `set_post_renditions`: creates or updates destination-specific copy for draft and scheduled posts. Renditions can only target accounts already attached as post destinations.
- `schedule_post`: creates a scheduled post with destination accounts and optional media, then queues the `publish_post` job.
- `schedule_draft`: schedules an existing draft and queues the `publish_post` job without duplicating the post. It can optionally replace source media before scheduling.
- `get_post_status`: returns the post status, scheduled run time, source media, and per-destination status.
- `list_scheduled_posts`: returns upcoming scheduled posts for queue inspection.
- `cancel_post`: cancels a queued scheduled post and returns it to drafts.
- `suggest_next_slot`: returns the next free configured posting slot for a workspace.
- `upload_media_from_url`: fetches a public HTTP(S) media URL and stores it in a workspace.
- `render_scheduler_widget`: renders structured OpenPost scheduler data in the ChatGPT Apps widget.

## Current prompts

- `plan_social_post`: guides an assistant from a rough idea to a workspace-aware draft.
- `adapt_platform_renditions`: guides destination-specific copywriting for an existing draft or scheduled post.
- `review_schedule`: guides queue inspection and next-action recommendations without mutating posts.

## Current scope

- Uses the same Bearer authentication path as the CLI and API tokens.
- Dedicated `mcp:full` tokens can be created in Settings for ChatGPT, Claude, and other MCP clients. Existing `cli:full` tokens also remain accepted by `/mcp` so `openpost-mcp` profiles continue to work.
- Publishes MCP protected-resource metadata and returns `WWW-Authenticate` plus `_meta["mcp/www_authenticate"]` challenges for unauthenticated MCP requests.
- Supports MCP `ping` and accepts `notifications/*` messages with HTTP `202 Accepted`, which keeps standard initialization handshakes quiet.
- Publishes OAuth authorization-server metadata for public PKCE clients, including `S256`, `mcp:full`, and client ID metadata document support.
- Provides a browser approval page at `/oauth/authorize` and a form-encoded `/oauth/token` code exchange that mints `mcp:full` API tokens.
- Validates client metadata redirect URIs for URL-based client IDs, accepts ChatGPT fallback redirects for predefined clients, and binds OAuth-issued MCP tokens to the `/mcp` resource audience.
- Advertises and enforces the `mcp:full` OAuth scope in every MCP tool descriptor. Fine-grained per-session scopes are still planned.
- Adds Apps SDK-friendly `_meta["openai/toolInvocation/invoking"]`, `_meta["openai/toolInvocation/invoked"]`, and `outputSchema` metadata to every tool descriptor.
- Exposes a ChatGPT Apps-compatible scheduler widget resource at `ui://widget/openpost-scheduler-v1.html`.
- Keeps data tools reusable across MCP clients and attaches widget UI metadata only to `render_scheduler_widget`.
- Provides `openpost-mcp` for local stdio clients without duplicating server tool logic.
- Advertises MCP prompt templates for common agentic scheduling workflows: planning a post, adapting platform renditions, and reviewing the publishing queue.
- Validates workspace membership and account ownership before returning, creating, scheduling, canceling, or uploading data.
- Lets assistants start from a publication source idea, attach workspace media, and later turn that publication into platform-specific drafts and renditions.
- Keeps draft iteration agent-friendly: assistants can list drafts, update draft copy/destinations, set per-destination renditions, and schedule the same draft when it is ready.
- Validates rendition targets against the post destination list so assistants do not create variants that would never publish.
- Rejects media URL fetches that resolve to private, loopback, link-local, multicast, or otherwise local addresses.
- Enforces the same scheduled-post and media-upload entitlement and usage accounting as the web/API paths.
- Records MCP tool calls in `mcp_tool_calls` with user, workspace, tool name, success/error status, error message, duration, and timestamp, and exposes recent calls in settings.
- Records API-token client ID, name, scope, and token prefix for MCP tool calls when a request uses a dedicated CLI/MCP token, so Settings can attribute activity to ChatGPT, Claude, CI, or another configured client.
- Returns structured content so assistants can inspect workspace, account, post, destination, media, and suggested slot IDs without parsing prose.
- Returns provider catalog structured content so assistants can avoid trying to connect or schedule to planned providers before adapters exist.
- Lets assistants attach workspace-owned source media to drafts and scheduled posts through `media_ids`, while preserving destination-specific media overrides through `set_post_renditions`.
