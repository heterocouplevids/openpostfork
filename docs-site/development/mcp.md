# MCP And ChatGPT App

OpenPost exposes an authenticated MCP foundation at:

```txt
POST /mcp
```

The endpoint is JSON-RPC over HTTP and requires:

```http
Authorization: Bearer <jwt-or-api-token>
```

## Initial tools

- `list_workspaces`: returns the workspaces available to the authenticated user.

## Current scope

- Uses the same Bearer authentication path as the CLI and API tokens.
- Keeps tools read-only while the MCP surface is being hardened.
- Returns structured content so assistants can inspect workspace IDs without parsing prose.

## Next tools

- `list_accounts`
- `create_draft`
- `upload_media_from_url`
- `schedule_post`
- `cancel_post`
- `get_post_status`
- `suggest_next_slot`
