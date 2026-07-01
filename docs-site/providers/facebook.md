# Facebook

Facebook support is available as an initial Pages publishing slice. It uses Meta OAuth, asks the user to choose a Page, and stores the selected Page access token.

## Requirements

- Meta developer app with Facebook Login configured
- OAuth redirect URL:

```text
https://your-domain.com/api/v1/accounts/facebook/callback
```

- App permissions:
  - `pages_show_list`
  - `pages_read_engagement`
  - `pages_manage_posts`
- Public `OPENPOST_MEDIA_URL` or S3/R2 public media URL for image and video posts

## Configuration

Configure Facebook through `OPENPOST_PROVIDER_APPS`:

```json
[
  {
    "provider": "facebook",
    "client_id": "your-meta-app-id",
    "client_secret": "your-meta-app-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives it from `OPENPOST_APP_URL`.

## Current Scope

- Connects a selected Facebook Page.
- Publishes text-only Page feed posts.
- Publishes one image URL through the Page photos endpoint.
- Publishes one video URL through the Page videos endpoint.
- Supports scheduling and platform variants through the normal OpenPost post flow.

## Current Limits

- No multi-image Page albums yet.
- No Facebook comments or thread-reply mapping yet.
- Media URLs must be public HTTPS URLs.
- Live-account verification is still recommended before relying on production Page publishing.

## Troubleshooting

- `facebook account has no manageable pages` usually means the authenticated user has no eligible Pages or the app lacks `pages_show_list`.
- Media publish failures usually mean the media URL is not public HTTPS or Meta cannot fetch it.
- Permission errors usually require Meta app review for the Pages permissions above.
