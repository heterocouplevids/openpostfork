# Instagram

Instagram support is available as an initial Instagram Business publishing slice. It uses Meta OAuth, asks the user to choose an Instagram Business account connected to a Facebook Page, and stores the selected Page access token.

## Requirements

- Meta developer app with Facebook Login configured
- Instagram Business or Creator account connected to a Facebook Page
- OAuth redirect URL:

```text
https://your-domain.com/api/v1/accounts/instagram/callback
```

- App permissions:
  - `instagram_basic`
  - `instagram_content_publish`
  - `pages_show_list`
  - `pages_read_engagement`
  - `business_management`
- Public `OPENPOST_MEDIA_URL` or S3/R2 public media URL for image and Reel video posts

## Configuration

Configure Instagram through `OPENPOST_PROVIDER_APPS`:

```json
[
  {
    "provider": "instagram",
    "client_id": "your-meta-app-id",
    "client_secret": "your-meta-app-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives it from `OPENPOST_APP_URL`.

## Current Scope

- Connects a selected Instagram Business account behind a Facebook Page.
- Publishes one image URL with a caption.
- Publishes one video URL as a Reel.
- Supports scheduling and platform variants through the normal OpenPost post flow.

## Current Limits

- No carousel, Story, comment, or insights support yet.
- No text-only Instagram posts.
- Media URLs must be public HTTPS URLs.
- Account discovery currently uses Pages returned by the authenticated Meta user.
- Live-account verification is still recommended before relying on production Instagram publishing.

## Troubleshooting

- `facebook account has no connected instagram business accounts` usually means the Meta user has no eligible Pages with connected Instagram Business accounts, or the app lacks the required scopes.
- Media publish failures usually mean the media URL is not public HTTPS or Meta cannot fetch it.
- Permission errors usually require Meta app review for the Instagram and Page permissions above.
