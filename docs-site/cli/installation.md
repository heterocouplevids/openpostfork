# CLI Installation

Install the OpenPost CLI when you want to manage a running OpenPost instance from a terminal or automation job.

## Install from GitHub Releases

The install script downloads the latest matching `openpost-cli-*` release asset and installs it as `openpost`:

```sh
curl -fsSL https://raw.githubusercontent.com/rodrgds/openpost/main/scripts/install-cli.sh | sh
```

It supports Linux and macOS on `amd64` and `arm64`.

## Manual Install

Download the matching CLI binary from [GitHub Releases](https://github.com/rodrgds/openpost/releases/latest), then put it on your `PATH`.

```sh
chmod +x openpost-cli-linux-amd64
sudo mv openpost-cli-linux-amd64 /usr/local/bin/openpost
openpost --version
```

## Build from Source

From the repository root:

```sh
devenv shell -- bash -lc 'cd cli && go build -ldflags="-s -w" -o ../openpost ./cmd/openpost'
```

Or from inside `cli/`:

```sh
go build ./cmd/openpost
```

## First Run

```sh
openpost instance add local http://localhost:8080
openpost instance use local
openpost auth login http://localhost:8080
openpost workspace list
openpost workspace use personal
```

For every command and flag, see the [generated CLI reference](/reference/cli).
