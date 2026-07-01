# Provider Roadmap

OpenPost is moving toward an agentic social scheduler: draft once, adapt per network, schedule through the web app, CLI, MCP, or ChatGPT-style clients, and keep the provider-specific details in one place.

The provider discovery API returns current and planned providers so clients can render a consistent account-connection surface.

| Status                | Meaning                                                                        |
| --------------------- | ------------------------------------------------------------------------------ |
| `available`           | Adapter is registered on this server and users can connect accounts.           |
| `needs_configuration` | Adapter exists, but the operator has not configured the provider app.          |
| `planned`             | Product roadmap item. The backend will not start a real OAuth flow for it yet. |

## Planned adapters

| Provider  | Initial product focus                                                                       |
| --------- | ------------------------------------------------------------------------------------------- |
| Instagram | Images, Reels, scheduling, per-platform variants, agent workflows.                          |
| Facebook  | Facebook Pages publishing, media posts, scheduling, per-platform variants, agent workflows. |
| YouTube   | Shorts, video publishing, scheduling, agent workflows.                                      |

## Implemented first slices

| Provider | Current product focus                                                                                   |
| -------- | ------------------------------------------------------------------------------------------------------- |
| TikTok   | One-video direct publishing through public HTTPS media URLs, scheduling, per-platform variants, MCP workflows. |

## Account-selection requirement

Some planned providers cannot be modeled as a single OAuth user profile:

- Facebook should connect a selected Page and save the Page token.
- Instagram should connect the selected Instagram Business account behind a Facebook Page.
- YouTube should connect the selected channel.

These adapters must implement the backend account-selection flow before they move from `planned` to connectable. TikTok uses a direct OAuth account flow and is connectable when configured, but its initial adapter is intentionally video-only.

## Implementation contract

Every provider still needs to implement the shared backend adapter before it becomes connectable:

- OAuth or app-password account connection.
- Token refresh behavior, when the provider supports refresh.
- Profile lookup for stable account identity, or account-selection support for page/channel providers.
- Media upload rules and validation.
- Publish behavior, including reply/thread semantics where available.
- Documentation for callbacks, app review requirements, media limits, and known API caveats.

Until an adapter lands, keep the provider in `status: "planned"` and do not accept it in `OPENPOST_PROVIDER_APPS`.
