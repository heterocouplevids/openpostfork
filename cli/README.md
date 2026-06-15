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
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

## Quickstart

Add an instance profile:

```sh
openpost instance add local http://localhost:8080
openpost instance use local
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

List and disconnect connected social accounts:

```sh
openpost account list
openpost account list --platform x
openpost account disconnect <account-id> --yes
```

Connecting accounts is still handled in the web UI:

```sh
openpost account connect x
```

Upload and list workspace media:

```sh
openpost media upload ./image.png --alt "Product screenshot"
openpost media list --limit 25
```

## Posting

Create a draft:

```sh
openpost post create --content "Hello from OpenPost" --accounts x --workspace personal
```

Schedule a post with natural language or RFC3339:

```sh
openpost post create --content "Launch note" --accounts x,linkedin --schedule "tomorrow 2pm"
openpost post create --content "Launch note" --accounts x --schedule 2026-06-20T14:00:00Z
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
openpost auth token list
openpost completion bash
openpost --version
```
