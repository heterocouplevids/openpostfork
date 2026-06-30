# User Docs

Use these docs when you want to operate OpenPost as a product: create publications, connect accounts, draft posts, schedule publishing, automate from the CLI, or let an assistant help through MCP.

## Web app

The web app is the main editorial surface.

- [Workspaces](/usage/workspaces) keep brands, accounts, prompts, schedules, and media separate.
- [Accounts](/usage/accounts) explains how connected provider identities appear in a workspace.
- [Publications](/usage/publications) covers the source idea workflow: one launch, article, update, or asset set that can become multiple platform-native posts.
- [Composing Posts](/usage/composing-posts) covers destination selection, media, variants, and the composer.
- [Threads](/usage/threads) covers multi-post sequences.
- [Scheduling](/usage/scheduling) covers queued publishing and failure visibility.
- [Media Library](/usage/media-library) covers reusable uploaded assets.

## CLI

The CLI is for terminal, CI, cron, and scripted workflows against a running OpenPost instance.

- [CLI Overview](/cli/) explains the command model.
- [Installation](/cli/installation) covers release binaries and source builds.
- [Authentication](/cli/authentication) covers browser login, device flow, and API-token login.
- [Publications](/cli/publications) covers source records from terminal workflows.
- [Posting](/cli/posting) covers posts, threads, media, social sets, and `next-slot`.
- [Automation](/cli/automation) covers CI and recurring jobs.
- [Command Reference](/reference/cli) is generated from the Cobra command tree.

## MCP

MCP is for authenticated assistant workflows. Use it when a client such as ChatGPT, Claude, Cursor, Codex, or another agent should inspect context, create drafts, adapt renditions, or schedule posts with OpenPost permissions.

- [MCP Assistant Scheduling](/mcp/) covers the user-facing MCP workflow.
- [MCP and ChatGPT App Developer Notes](/development/mcp) cover implementation details for contributors.

## Where not to look

- Deployment, backups, provider credentials, and operational settings live in [Self-Hosting](/self-hosting/).
- Repository architecture, backend/frontend internals, API generation, and tests live in [Developer Docs](/development/).
