# TikTok

TikTok support is available as an initial video publishing slice. It uses OAuth plus the Content Posting API direct-post video endpoint.

## What you need

- TikTok developer app
- Login Kit and Content Posting API access
- Provider app registry entry with provider `tiktok`
- Callback URL: `https://your-domain.com/api/v1/accounts/tiktok/callback`
- Public `OPENPOST_MEDIA_URL` or S3/R2 public media URL
- Scopes: `user.info.basic`, `user.info.profile`, `video.publish`, `video.upload`

Example `OPENPOST_PROVIDER_APPS` entry:

```json
[
  {
    "provider": "tiktok",
    "client_id": "your-client-key",
    "client_secret": "your-client-secret",
    "redirect_uri": "https://your-domain.com/api/v1/accounts/tiktok/callback"
  }
]
```

## Current limits

- One video attachment per post.
- Public HTTPS media URL required.
- Text-only, image, carousel, inbox upload, and photo-post paths are not enabled yet.
- Live-account and app-review behavior still needs deployment verification.

## Common issues

- `OPENPOST_MEDIA_URL` points at localhost or a private host.
- TikTok app lacks Content Posting API access or required scopes.
- The TikTok app's redirect URI does not exactly match OpenPost's callback URL.
