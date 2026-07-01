# CLI Reference

This page is generated from the Cobra command tree. Do not edit it by hand.

Regenerate with:

```sh
cd cli && go run ./cmd/openpost-docs ../docs-site/reference/cli.md
```

## `openpost`

OpenPost CLI — control a self-hosted OpenPost instance from the terminal

openpost is a command-line client for the OpenPost social media scheduler.  It talks to a running OpenPost instance over HTTPS, authenticates with a revocable API token, and exposes the most common posting, scheduling, account, and media workflows for use from scripts, CI, and power-user shells.

**Usage**

```text
openpost [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `-h, --help` | `false` | help for openpost |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `-v, --version` | `false` | version for openpost |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost account` | Manage connected social accounts |
| `openpost auth` | Authenticate with an OpenPost instance |
| `openpost completion` | Generate shell completion script |
| `openpost instance` | Manage OpenPost instance profiles |
| `openpost jobs` | List background jobs |
| `openpost media` | Upload and list media attachments |
| `openpost post` | Create, list, view, update, and delete posts |
| `openpost publication` | Create and manage source publications |
| `openpost set` | Manage workspace social sets |
| `openpost thread` | Create multi-post threads |
| `openpost version` | Print the openpost CLI version |
| `openpost workspace` | Manage the active OpenPost workspace |

### `openpost account`

Manage connected social accounts

List, rename, and disconnect social accounts. Account slugs are the preferred selector for --accounts. New accounts are connected in the OpenPost web UI at &lt;instance&gt;/accounts.

**Usage**

```text
openpost account
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost account disconnect` | Disconnect a social account |
| `openpost account list` | List connected social accounts |
| `openpost account rename` | Rename a social account slug |

### `openpost account disconnect`

Disconnect a social account

**Usage**

```text
openpost account disconnect &lt;account-id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost account list`

List connected social accounts

List connected social accounts for the active workspace.  Use the SLUG column as the preferred selector for --accounts and account rename.

**Usage**

```text
openpost account list [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--platform` | `-` | filter by platform |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost account rename`

Rename a social account slug

Rename a connected account's slug. The selector can be an account id, slug, platform:username value, bare platform when unambiguous, or mastodon host.

**Usage**

```text
openpost account rename &lt;selector&gt; --slug &lt;new-slug&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--slug` | `-` | new account slug |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost auth`

Authenticate with an OpenPost instance

**Usage**

```text
openpost auth
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost auth login` | Log in to an OpenPost instance |
| `openpost auth logout` | Delete the stored token for the active profile |
| `openpost auth status` | Show authentication status for the active profile |
| `openpost auth token` | Manage API tokens |

### `openpost auth login`

Log in to an OpenPost instance

**Usage**

```text
openpost auth login &lt;instance&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--device` | `false` | print the device code and poll without opening a browser |
| `--insecure-storage` | `false` | store the token in credentials.json instead of the OS keyring |
| `--no-browser` | `false` | skip automatically opening the browser |
| `--with-token` | `false` | read a raw API token from stdin |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost auth logout`

Delete the stored token for the active profile

**Usage**

```text
openpost auth logout
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost auth status`

Show authentication status for the active profile

**Usage**

```text
openpost auth status
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost auth token`

Manage API tokens

**Usage**

```text
openpost auth token
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost auth token list` | List API tokens |
| `openpost auth token revoke` | Revoke an API token |

### `openpost auth token list`

List API tokens

**Usage**

```text
openpost auth token list
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost auth token revoke`

Revoke an API token

**Usage**

```text
openpost auth token revoke &lt;id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost completion`

Generate shell completion script

Output a shell completion script for the given shell.  To load completions:  Bash:   $ source &lt;(openpost completion bash)    # To load completions for each session, execute once:   # Linux:   $ openpost completion bash &gt; /etc/bash_completion.d/openpost   # macOS:   $ openpost completion bash &gt; $(brew --prefix)/etc/bash_completion.d/openpost  Zsh:   # If shell completion is not already enabled in your environment,   # you will need to enable it. You can execute the following once:   $ echo "autoload -U compinit; compinit" &gt;&gt; ~/.zshrc    # To load completions for each session, execute once:   $ openpost completion zsh &gt; "${fpath[1]}/_openpost"    # You will need to start a new shell for this setup to take effect.  Fish:   $ openpost completion fish \| source    # To load completions for each session, execute once:   $ openpost completion fish &gt; ~/.config/fish/completions/openpost.fish  PowerShell:   PS&gt; openpost completion powershell \| Out-String \| Invoke-Expression    # To load completions for every new session, run:   PS&gt; openpost completion powershell &gt; openpost.ps1   # and source this file from your PowerShell profile.

**Usage**

```text
openpost completion &lt;bash\|zsh\|fish\|powershell&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance`

Manage OpenPost instance profiles

**Usage**

```text
openpost instance
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost instance add` | Add or update an instance profile |
| `openpost instance diagnostics` | Collect a safe support snapshot for an OpenPost instance |
| `openpost instance health` | Check the active instance liveness and readiness |
| `openpost instance list` | List configured instances |
| `openpost instance remove` | Remove an instance profile |
| `openpost instance use` | Set the active instance profile |

### `openpost instance add`

Add or update an instance profile

**Usage**

```text
openpost instance add &lt;name&gt; &lt;url&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance diagnostics`

Collect a safe support snapshot for an OpenPost instance

**Usage**

```text
openpost instance diagnostics
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance health`

Check the active instance liveness and readiness

**Usage**

```text
openpost instance health
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance list`

List configured instances

**Usage**

```text
openpost instance list
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance remove`

Remove an instance profile

**Usage**

```text
openpost instance remove &lt;name&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost instance use`

Set the active instance profile

**Usage**

```text
openpost instance use &lt;name&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost jobs`

List background jobs

**Usage**

```text
openpost jobs
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost jobs list` | List background jobs |

### `openpost jobs list`

List background jobs

**Usage**

```text
openpost jobs list [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `0` | maximum number of jobs to return |
| `--status` | `-` | filter by status: pending, failed, completed |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost media`

Upload and list media attachments

**Usage**

```text
openpost media
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost media list` | List media attachments |
| `openpost media upload` | Upload a media file |

### `openpost media list`

List media attachments

**Usage**

```text
openpost media list [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `0` | maximum number of media items to return |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost media upload`

Upload a media file

**Usage**

```text
openpost media upload &lt;file&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--alt` | `-` | alt text for the uploaded media |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost post`

Create, list, view, update, and delete posts

**Usage**

```text
openpost post
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost post create` | Create a draft or scheduled post |
| `openpost post delete` | Delete a draft or scheduled post |
| `openpost post list` | List posts |
| `openpost post update` | Update a draft or scheduled post |
| `openpost post view` | View a post |

### `openpost post create`

Create a draft or scheduled post

**Usage**

```text
openpost post create [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors |
| `--content` | `-` | post content |
| `--file` | `-` | read post content from a file |
| `--media` | `[]` | media id or local file path; repeatable |
| `--media-alt` | `[]` | alt text for the matching uploaded --media |
| `--publication` | `-` | source publication ID |
| `--random-delay` | `0` | random delay in minutes |
| `--schedule` | `-` | natural-language, RFC3339, next-slot, now, or draft |
| `--set` | `-` | social set name or ID to publish to |
| `--thread-draft` | `-` | encoded thread draft to attach |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost post delete`

Delete a draft or scheduled post

**Usage**

```text
openpost post delete &lt;post-id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost post list`

List posts

**Usage**

```text
openpost post list [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `0` | maximum number of posts to return |
| `--status` | `-` | filter by status: draft, scheduled, published, failed |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost post update`

Update a draft or scheduled post

**Usage**

```text
openpost post update &lt;post-id&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors |
| `--content` | `-` | post content |
| `--publication` | `-` | source publication ID; pass an empty string to clear |
| `--random-delay` | `0` | random delay in minutes |
| `--schedule` | `-` | natural-language, RFC3339, next-slot, now, or draft; empty string unschedules |
| `--set` | `-` | social set name or ID to publish to |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost post view`

View a post

**Usage**

```text
openpost post view &lt;post-id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost publication`

Create and manage source publications

Create and manage source publications: the canonical idea, brief, or source material that posts, threads, and assistant workflows can reference through --publication.

**Usage**

```text
openpost publication
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost publication create` | Create a source publication |
| `openpost publication list` | List source publications |
| `openpost publication update` | Update a source publication |
| `openpost publication view` | View a source publication |

### `openpost publication create`

Create a source publication

**Usage**

```text
openpost publication create [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--audience` | `-` | intended audience |
| `--content` | `-` | source idea, brief, announcement, notes, or canonical material |
| `--file` | `-` | read source content from a file |
| `--goal` | `-` | goal such as announce, explain, launch, ask for feedback, or promote article |
| `--media` | `[]` | media id or local file path to attach; repeatable |
| `--media-alt` | `[]` | alt text for the matching uploaded --media |
| `--source-url` | `-` | source URL related to the publication |
| `--title` | `-` | short internal title |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost publication list`

List source publications

**Usage**

```text
openpost publication list [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--limit` | `0` | maximum number of publications to return |
| `--offset` | `0` | pagination offset |
| `--status` | `-` | filter by status: draft, ready, scheduled, published, failed |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost publication update`

Update a source publication

**Usage**

```text
openpost publication update &lt;publication-id&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--audience` | `-` | intended audience |
| `--clear-media` | `false` | remove all source media attachments |
| `--content` | `-` | source idea, brief, announcement, notes, or canonical material |
| `--file` | `-` | read source content from a file |
| `--goal` | `-` | goal such as announce, explain, launch, ask for feedback, or promote article |
| `--media` | `[]` | media id or local file path to attach; repeatable |
| `--media-alt` | `[]` | alt text for the matching uploaded --media |
| `--source-url` | `-` | source URL related to the publication |
| `--status` | `-` | status: draft, ready, scheduled, published, failed |
| `--title` | `-` | short internal title |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost publication view`

View a source publication

**Usage**

```text
openpost publication view &lt;publication-id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set`

Manage workspace social sets

Manage workspace social sets: reusable groups of social accounts.  Posts and threads use the workspace default set when neither --accounts nor --set is passed.

**Usage**

```text
openpost set
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost set add` | Add accounts to a social set |
| `openpost set create` | Create a social set |
| `openpost set default` | Set or clear the workspace default social set |
| `openpost set delete` | Delete a social set |
| `openpost set list` | List social sets |
| `openpost set remove` | Remove accounts from a social set |
| `openpost set rename` | Rename a social set |

### `openpost set add`

Add accounts to a social set

**Usage**

```text
openpost set add &lt;set&gt; --accounts &lt;selectors&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors to add |
| `--main` | `false` | mark added accounts as main accounts |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set create`

Create a social set

**Usage**

```text
openpost set create &lt;name&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors to include |
| `--default` | `false` | make this the workspace default set |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set default`

Set or clear the workspace default social set

**Usage**

```text
openpost set default &lt;set&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--unset` | `false` | clear default status instead of making the set default |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set delete`

Delete a social set

**Usage**

```text
openpost set delete &lt;set&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set list`

List social sets

**Usage**

```text
openpost set list
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set remove`

Remove accounts from a social set

**Usage**

```text
openpost set remove &lt;set&gt; --accounts &lt;selectors&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors to remove |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost set rename`

Rename a social set

**Usage**

```text
openpost set rename &lt;set&gt; &lt;name&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost thread`

Create multi-post threads

**Usage**

```text
openpost thread
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost thread create` | Create a thread from a markdown file |

### `openpost thread create`

Create a thread from a markdown file

**Usage**

```text
openpost thread create &lt;file&gt; [flags]
```

**Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--accounts` | `-` | comma-separated account selectors |
| `--publication` | `-` | source publication ID |
| `--random-delay` | `0` | random delay in minutes |
| `--schedule` | `-` | natural-language, RFC3339, next-slot, now, or draft |
| `--set` | `-` | social set name or ID to publish to |

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost version`

Print the openpost CLI version

**Usage**

```text
openpost version
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost workspace`

Manage the active OpenPost workspace

**Usage**

```text
openpost workspace
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

**Subcommands**

| Command | Description |
| --- | --- |
| `openpost workspace create` | Create a workspace |
| `openpost workspace list` | List workspaces |
| `openpost workspace use` | Set the active workspace for the current profile |

### `openpost workspace create`

Create a workspace

**Usage**

```text
openpost workspace create &lt;name&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost workspace list`

List workspaces

**Usage**

```text
openpost workspace list
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

### `openpost workspace use`

Set the active workspace for the current profile

**Usage**

```text
openpost workspace use &lt;name-or-id&gt;
```

**Inherited Flags**

| Flag | Default | Description |
| --- | --- | --- |
| `--instance` | `-` | OpenPost instance URL (default: profile or $OPENPOST_INSTANCE) |
| `--json` | `false` | emit machine-readable JSON instead of tables/prose |
| `--no-color` | `false` | disable ANSI colors |
| `--profile` | `-` | profile name from config (default: $OPENPOST_PROFILE or 'default') |
| `--quiet` | `false` | suppress non-error output |
| `--token` | `-` | API token override (default: keyring or $OPENPOST_TOKEN) |
| `--workspace` | `-` | workspace name or ID (default: profile or $OPENPOST_WORKSPACE) |
| `--yes` | `false` | skip interactive confirmations |

