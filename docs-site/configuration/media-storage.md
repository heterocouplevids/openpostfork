# Media Storage

OpenPost stores media through its `BlobStorage` abstraction. Local filesystem storage is the default; S3-compatible storage is the cloud-ready driver path.

## Key settings

- `OPENPOST_MEDIA_PATH` controls where files are stored on disk.
- `OPENPOST_MEDIA_URL` controls how those files are exposed publicly.
- `OPENPOST_STORAGE_DRIVER` chooses `local` or `s3`.

## Recommended production values

```sh
OPENPOST_MEDIA_PATH=/data/media
OPENPOST_MEDIA_URL=https://openpost.example.com/media
```

## Why public media URLs matter

Threads requires the backend to hand Meta a publicly reachable media URL. If OpenPost cannot expose the file publicly, Threads media publishing will fail.

## Backups

Back up the media directory together with the SQLite database. You need both for a complete restore.

## S3-compatible storage

Use these settings for S3/R2-style storage:

```sh
OPENPOST_STORAGE_DRIVER=s3
OPENPOST_S3_ENDPOINT=https://<account>.r2.cloudflarestorage.com
OPENPOST_S3_REGION=auto
OPENPOST_S3_BUCKET=openpost-media
OPENPOST_S3_ACCESS_KEY_ID=...
OPENPOST_S3_SECRET_ACCESS_KEY=...
OPENPOST_S3_PUBLIC_BASE_URL=https://media.openpost.example
OPENPOST_S3_FORCE_PATH_STYLE=false
```

The S3-compatible storage driver supports server-side uploads and direct browser-to-S3 upload sessions.

Direct upload flow:

1. Call `POST /api/v1/media/upload-session` with `workspace_id`, `filename`, `mime_type`, and `size`.
2. Upload the file to the returned presigned `PUT` target with the returned headers.
3. Call `POST /api/v1/media/upload-session/{media_id}/complete` with the same `workspace_id`.

OpenPost reserves a pending media record before issuing the presigned URL, then finalizes the upload by reading the stored object, computing metadata and the SHA-256 dedupe hash, creating thumbnails when possible, recording media-upload usage, and marking the media ready. Local filesystem deployments should keep using the existing multipart upload endpoint.

The web app uses the direct upload session flow automatically when the server supports it. Local filesystem deployments and older servers fall back to multipart uploads.

Provider media-state tracking is still part of the production-readiness roadmap.
