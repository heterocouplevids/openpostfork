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
| YouTube   | Shorts, video publishing, scheduling, agent workflows.                                      |

## Implemented first slices

| Provider | Current product focus                                                                                   |
| -------- | ------------------------------------------------------------------------------------------------------- |
| Facebook | Selected Page publishing for text, one public HTTPS image URL, or one public HTTPS video URL. |
| Instagram | Selected Instagram Business account publishing for one public HTTPS image URL or Reel video URL. |
| TikTok   | One-video direct publishing through public HTTPS media URLs, scheduling, per-platform variants, MCP workflows. |

## Account-selection requirement

Some providers cannot be modeled as a single OAuth user profile:

- Instagram connects the selected Instagram Business account behind a Facebook Page.
- YouTube should connect the selected channel.

Instagram and Facebook use the backend account-selection flow today. YouTube must implement it before it moves from `planned` to connectable. TikTok uses a direct OAuth account flow and is connectable when configured, but its initial adapter is intentionally video-only.

## Implementation contract

Every provider still needs to implement the shared backend adapter before it becomes connectable:

- OAuth or app-password account connection.
- Token refresh behavior, when the provider supports refresh.
- Profile lookup for stable account identity, or account-selection support for page/channel providers.
- Media upload rules and validation.
- Publish behavior, including reply/thread semantics where available.
- Documentation for callbacks, app review requirements, media limits, and known API caveats.

Until an adapter lands, keep the provider in `status: "planned"` and do not accept it in `OPENPOST_PROVIDER_APPS`.
