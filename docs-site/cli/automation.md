# CLI Automation

The CLI is designed for headless use in CI, cron, and deployment jobs. Use an API token created in OpenPost and pass it through environment variables or stdin.

## Environment

| Variable | Purpose |
| --- | --- |
| `OPENPOST_TOKEN` | Bearer token for non-interactive API access. |
| `OPENPOST_INSTANCE` | Default OpenPost instance URL. |
| `OPENPOST_WORKSPACE` | Default workspace ID or name. |
| `OPENPOST_OUTPUT=json` | Default output format for scripts. |
| `OPENPOST_PROFILE` | Selects a named CLI profile. |

Useful flags:

| Flag | Purpose |
| --- | --- |
| `--yes` | Skip confirmation prompts. |
| `--json` | Print machine-readable JSON for one command. |
| `--set <name-or-id>` | Publish through a managed social set instead of spelling out `--accounts`. |
| `--schedule next-slot` | Use the next available posting schedule slot instead of posting immediately or choosing a fixed time. |

For recurring jobs, prefer a workspace default social set and `--schedule next-slot`. After creating a default set with `openpost set create <name> --accounts ... --default`, automation can omit both `--accounts` and `--set`; the CLI resolves the default set at runtime.

## GitHub Actions Example

```yaml
name: Daily Build Summary

on:
  schedule:
    - cron: "0 17 * * 1-5"

jobs:
  post-summary:
    runs-on: ubuntu-latest
    env:
      OPENPOST_INSTANCE: ${{ secrets.OPENPOST_INSTANCE }}
      OPENPOST_TOKEN: ${{ secrets.OPENPOST_TOKEN }}
      OPENPOST_WORKSPACE: ${{ secrets.OPENPOST_WORKSPACE }}
      OPENPOST_OUTPUT: json
    steps:
      - name: Install OpenPost CLI
        run: curl -fsSL https://raw.githubusercontent.com/rodrgds/openpost/main/scripts/install-cli.sh | sh

      - name: Post summary
        run: |
          openpost post create \
            --content "Daily build completed for ${GITHUB_REPOSITORY}@${GITHUB_SHA}" \
            --schedule next-slot \
            --yes \
            --json
```

The complete command and flag reference is generated from the Cobra command tree at [CLI Reference](/reference/cli).
