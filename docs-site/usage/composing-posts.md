# Composing Posts

OpenPost supports one post targeting multiple providers, with room for per-account variants.

## Typical workflow

1. Write the base post content.
2. Select target platforms.
3. Attach media if needed.
4. Add per-account variants where one copy does not fit all destinations.
5. Schedule or publish.

## Drafts and renditions

Drafts are the source of truth before publishing. Start with one base post,
then use per-account renditions when a platform needs different copy, hashtags,
or formatting. Unsynchronized renditions stay editable independently while the
base draft remains available for future changes.

## Platform previews

The preview panel renders each selected account separately, so multiple pages,
channels, or profiles on the same provider stay visible. Instagram, Facebook,
YouTube, and TikTok use provider-shaped cards instead of the generic preview,
and media warnings surface first-slice limits before publishing.

Drafts can still be saved while incomplete. Web, API, and MCP scheduling
validate destination media requirements server-side and return an error if a
provider cannot publish the selected attachments.

## Practical advice

- Keep one canonical message first, then customize only where a provider needs it.
- Validate media size and count before scheduling large batches.
