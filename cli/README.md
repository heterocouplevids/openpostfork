# OpenPost CLI

Command-line client for a running OpenPost instance.

## Build

From the repository root:

```sh
devenv shell -- bash -lc 'cd cli && go build ./cmd/openpost'
```

The binary is written to `cli/openpost`.

## Install

During development, install into your Go bin directory:

```sh
devenv shell -- bash -lc 'cd cli && go install ./cmd/openpost'
devenv shell -- bash -lc 'cd cli && go install ./cmd/openpost-mcp'
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Quickstart

Add an instance profile:

```sh
openpost instance add local http://localhost:8080
openpost instance use local
openpost instance health
```

Log in with the browser device flow:

```sh
openpost auth login http://localhost:8080
```

For headless shells, print the verification URL and code without opening a browser:

```sh
openpost auth login http://localhost:8080 --device
```

For automation, pass an existing API token through stdin:

```sh
printf '%s\n' "$OPENPOST_TOKEN" | openpost auth login http://localhost:8080 --with-token
```

Select a workspace:

```sh
openpost workspace list
openpost workspace use personal
```

## Account and media commands

List, rename, and disconnect connected social accounts:

```sh
openpost account list
openpost account list --platform x
openpost account rename x --slug main-x
openpost account disconnect <account-id> --yes
```

New accounts are connected in the OpenPost web UI at `<instance>/accounts`.
The CLI does not have a `connect` subcommand by design: provider credentials
live on the server, and the web UI is the only place to authorize a new social
account. Running `account list` against a workspace with no accounts prints
the instance's `/accounts` URL so the path is discoverable.

Upload and list workspace media:

```sh
openpost media upload ./image.png --alt "Product screenshot"
openpost media list --limit 25
```

## Social sets

Social sets are reusable groups of social accounts in a workspace. Create a
default set once, then `post create` and `thread create` use it automatically
when neither `--accounts` nor `--set` is passed:

```sh
openpost set create launch --accounts main-x,linkedin --default
openpost set list
openpost set add launch --accounts bluesky
openpost set remove launch --accounts linkedin
```

You can also target a specific set per command:

```sh
openpost post create --content "Launch note" --set launch
openpost thread create ./thread.md --set launch --schedule "next monday 9am"
```

## Posting

Create a draft:

```sh
openpost post create --content "Hello from OpenPost" --accounts x --workspace personal
openpost post create --content "Hello from the default social set"
```

Schedule a post with natural language or RFC3339:

```sh
openpost post create --content "Launch note" --accounts x,linkedin --schedule "tomorrow 2pm"
openpost post create --content "Launch note" --accounts x --schedule 2026-06-20T14:00:00Z
```

Use the next available posting slot from the workspace schedule:

```sh
openpost post create --content "Launch note" --schedule next-slot
openpost thread create ./thread.md --set launch --schedule next-slot
```

List and inspect posts:

```sh
openpost post list --status scheduled --limit 20
openpost post view <post-id>
```

Create a thread from markdown segments separated by `---` lines:

```sh
openpost thread create ./thread.md --accounts x --schedule "next monday 9am"
```

Useful diagnostics:

```sh
openpost auth status
openpost instance diagnostics --json
openpost auth token list
openpost completion bash
openpost --version
```

## MCP stdio proxy

`openpost-mcp` lets desktop MCP clients talk to the same authenticated remote
`/mcp` endpoint using the active CLI profile and token. Configure an instance
and log in with `openpost` first, then point the MCP client at:

```sh
openpost-mcp --profile local
```

You can also pass `--instance` and `--token` directly for automation.
