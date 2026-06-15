# Posting with the CLI

Use `openpost post` for single posts, `openpost media` for uploads, and `openpost thread` for multi-post markdown threads.

## Create a Scheduled Post

```sh
openpost post create --accounts x --content 'Shipping the first CLI release today.' --schedule 'tomorrow 2pm'
```

You can also schedule with an RFC3339 timestamp:

```sh
openpost post create --accounts x --file launch.md --schedule '2026-06-15T09:00:00+01:00'
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
accounts: x
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
| `2pm` | Today at 14:00 if still in the future, otherwise tomorrow at 14:00. |
| `tomorrow 2pm` | Tomorrow at 14:00 in the resolved workspace/profile/local timezone. |
| `in 3 hours` | Three hours after the command runs. |
| `next monday 9am` | The next Monday after today at 09:00. |
| `2026-06-15T09:00:00+01:00` | The exact RFC3339 instant with the supplied offset. |
| `2026-06-15 09:00` | The local date and time in the resolved workspace/profile/local timezone. |

`today` or `tomorrow` without a time is rejected so scheduled posts do not land at an accidental default time.
