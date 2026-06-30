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

The storage driver setting is available now; direct browser-to-S3 upload sessions and provider media-state tracking are part of the production-readiness roadmap.
