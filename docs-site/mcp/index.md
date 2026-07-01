# Assistant Scheduling With MCP

OpenPost's MCP support lets ChatGPT-style clients and local desktop assistants work with your scheduler through the same authenticated OpenPost instance you use in the web app and CLI.

Use it when you want an assistant to:

- inspect workspaces, connected accounts, media, drafts, publications, providers, and scheduled posts
- turn a rough idea into a source publication, then keep drafts and scheduled posts linked to that source
- adapt copy for each destination before scheduling
- attach existing workspace media or upload media from a public URL
- suggest the next posting slot, schedule approved posts, or cancel queued posts

## Ways to connect

### ChatGPT-style clients

Use the remote MCP endpoint from your OpenPost instance:

```txt
https://your-openpost-host.example/mcp
```

OAuth-aware clients can use OpenPost's browser account-linking flow. Clients that need a manual token can use a dedicated `mcp:full` token from **Settings -> CLI Devices & API Tokens**.

When approving OAuth or creating a manual token, prefer the current-workspace boundary unless the client truly needs every workspace you can access.

### Desktop MCP clients

Install and authenticate the OpenPost CLI, then run the local stdio proxy:

```sh
openpost-mcp --profile local
```

The proxy reads the selected CLI profile and forwards MCP frames to the remote `/mcp` endpoint. It does not open the database and does not need provider secrets on the client machine.

## Current assistant tools

- `list_workspaces`
- `list_provider_catalog`
- `list_accounts`
- `list_media`
- `list_publications`
- `create_publication`
- `create_draft`
- `list_drafts`
- `update_draft`
- `set_post_renditions`
- `schedule_post`
- `schedule_draft`
- `get_post_status`
- `list_scheduled_posts`
- `cancel_post`
- `suggest_next_slot`
- `upload_media_from_url`
- `render_scheduler_widget`

## Safe workflow

1. Ask the assistant to inspect the current workspace, provider catalog, accounts, and recent media.
2. Create or select a source publication.
3. Pass that `publication_id` when drafting or scheduling so the post stays linked to its source.
4. Draft the base post and destination-specific renditions.
5. Review the proposed schedule and destination list.
6. Let the assistant schedule only after the final content and accounts are correct.

MCP tools validate workspace membership, optional token workspace boundaries, and account ownership before reading or changing data. Scheduling and media uploads use the same quota and usage accounting as the web app and CLI.

## Activity and revocation

Recent MCP tool calls appear in **Settings -> CLI Devices & API Tokens** with client attribution when the request used a dedicated MCP or CLI token. Revoke the token there to disconnect a client.

For protocol details, Apps SDK metadata, OAuth discovery, and implementation notes, see [MCP And ChatGPT App](/development/mcp) in the developer docs.
