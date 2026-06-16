# Posting with the CLI

Use `openpost post` for single posts, `openpost media` for uploads, and `openpost thread` for multi-post markdown threads.

## Choose Accounts or Sets

Use `--accounts` for one-off account selection:

```sh
openpost post create --accounts x,linkedin --content 'Shipping today.'
```

Use social sets for reusable groups:

```sh
openpost set create launch --accounts x,linkedin --default
openpost post create --set launch --content 'Shipping today.'
```

If neither `--accounts` nor `--set` is passed, `post create` and `thread create`
use the workspace default social set when one exists. If no default set exists,
the post remains a draft with no destinations, matching the previous CLI
behavior.

Manage sets with:

```sh
openpost set list
openpost set create <name> --accounts x,linkedin --default
openpost set add <name> --accounts bluesky
openpost set remove <name> --accounts linkedin
openpost set default <name>
openpost set delete <name> --yes
```

## Create a Scheduled Post

```sh
openpost post create --accounts x --content 'Shipping the first CLI release today.' --schedule 'tomorrow 2pm'
```

You can also schedule with an RFC3339 timestamp:

```sh
openpost post create --accounts x --file launch.md --schedule '2026-06-15T09:00:00+01:00'
```

Use the next available workspace posting slot:

```sh
openpost post create --content 'Shipping today.' --schedule next-slot
```

When the command uses a social set, `next-slot` asks the server for the next
slot for that set:

```sh
openpost post create --set launch --content 'Shipping today.' --schedule next-slot
openpost thread create launch.md --set launch --schedule next-slot
```

## Attach Media

Upload media first, then pass the returned media ID to the post command:

```sh
openpost media upload ./image.png --alt 'Product screenshot showing the new queue view'
openpost post create --accounts x --content 'New queue view is live.' --media <id> --schedule 'next monday 9am'
```

## Create a Thread

Create `launch.md` with front matter and `---` separators:

```md
---
set: launch
schedule: tomorrow 2pm
---

We shipped the OpenPost CLI today.

---

It supports browser login, device mode for SSH hosts, and token-based automation.

---

Install it from the latest GitHub release and run:

openpost auth login <instance>
```

Then create the thread:

```sh
openpost thread create launch.md
```

## JSON Output

Automation can pass a token through the environment and request JSON output:

```sh
OPENPOST_TOKEN=op_cli_... openpost post list --json
```

## Relative Schedule

| Input | Resolution |
| --- | --- |
| `now` | The next one-minute boundary, so the publish worker can pick it up. |
| `draft` | No scheduled time; the post remains a draft. |
| `next-slot` / `next slot` / `slot` | The next available posting schedule slot from the server. |
| `2pm` | Today at 14:00 if still in the future, otherwise tomorrow at 14:00. |
| `tomorrow 2pm` | Tomorrow at 14:00 in the resolved workspace/profile/local timezone. |
| `in 3 hours` | Three hours after the command runs. |
| `next monday 9am` | The next Monday after today at 09:00. |
| `2026-06-15T09:00:00+01:00` | The exact RFC3339 instant with the supplied offset. |
| `2026-06-15 09:00` | The local date and time in the resolved workspace/profile/local timezone. |

`today` or `tomorrow` without a time is rejected so scheduled posts do not land at an accidental default time.
