# CLI Publications

Use `openpost publication` to store the source idea, notes, article, launch brief, or campaign context before creating platform posts.

## Create a Source

```sh
openpost publication create \
  --title "June launch" \
  --file launch-notes.md \
  --goal announce \
  --audience builders
```

Attach existing media IDs or local files:

```sh
openpost publication create \
  --title "Feature walkthrough" \
  --content "Show the new scheduler workflow." \
  --media ./screenshot.png \
  --media-alt "OpenPost scheduler screenshot"
```

## Link Posts

Pass the publication ID when creating or updating posts:

```sh
openpost post create --publication <publication-id> --set launch --content "Shipping today." --schedule next-slot
openpost post update <post-id> --publication <publication-id>
```

Clear a link with an empty value:

```sh
openpost post update <post-id> --publication ""
```

Thread files can also include the source:

```md
---
set: launch
schedule: next-slot
publication: <publication-id>
---

First post.

---

Second post.
```

## Review and Update

```sh
openpost publication list --status draft
openpost publication view <publication-id>
openpost publication update <publication-id> --status ready
openpost publication update <publication-id> --clear-media
```

For all flags, see the [generated CLI reference](/reference/cli).
