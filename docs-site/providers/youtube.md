# YouTube

YouTube support is available as an initial video upload slice. It uses Google OAuth, asks the user to choose a YouTube channel, and stores the Google refresh token for scheduled uploads.

## Requirements

- Google Cloud OAuth app with the YouTube Data API v3 enabled
- OAuth redirect URL:

```text
https://your-domain.com/api/v1/accounts/youtube/callback
```

- OAuth scopes:
  - `https://www.googleapis.com/auth/userinfo.profile`
  - `https://www.googleapis.com/auth/userinfo.email`
  - `https://www.googleapis.com/auth/youtube.readonly`
  - `https://www.googleapis.com/auth/youtube.upload`
- One video attachment on the OpenPost post or YouTube-specific variant

## Configuration

Configure YouTube through `OPENPOST_PROVIDER_APPS`:

```json
[
  {
    "provider": "youtube",
    "client_id": "your-google-oauth-client-id",
    "client_secret": "your-google-oauth-client-secret"
  }
]
```

If `redirect_uri` is omitted, OpenPost derives it from `OPENPOST_APP_URL`.

## Current Scope

- Connects a selected YouTube channel.
- Uploads one video through the YouTube Data API `videos.insert` endpoint.
- Uploads videos as private by default.
- Derives the video title from the first non-empty line of the post or platform variant.
- Uses the full post or variant content as the video description.
- Supports scheduling and platform variants through the normal OpenPost post flow.

## Current Limits

- No public/unlisted privacy selector yet.
- No tags, category, made-for-kids, thumbnail, playlist, or comment support yet.
- No resumable upload flow yet; keep the first slice focused on smaller video/Shorts uploads.
- Live-account verification is still recommended before relying on production YouTube publishing.

## Troubleshooting

- `google account has no YouTube channels` usually means the authenticated Google user has no YouTube channel available to the OAuth app.
- `invalidTitle` usually means the first line of the post or variant is empty or invalid after trimming.
- `mediaBodyRequired` usually means the video file could not be read from OpenPost media storage.
- Upload permission errors usually mean the Google Cloud project lacks YouTube Data API v3 access or the OAuth app has not been verified for the requested scopes.
