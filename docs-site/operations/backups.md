# Backups

You need both the database and the media directory for a usable backup.

## What to back up

- SQLite database
- Media directory
- Your `.env` file or secret-management equivalent
- File ownership and permissions for the runtime directories

## Backup example

Stop OpenPost first if you want the simplest backup path:

```bash
sudo systemctl stop openpost
```

If your deployment keeps SQLite in WAL mode, copy the database together with any `-wal` and `-shm` files that exist. Those extra files can contain committed data that has not yet been checkpointed into the main `.db` file.

### Database

```bash
cp /var/lib/openpost/openpost.db openpost-backup-$(date +%Y%m%d).db
cp /var/lib/openpost/openpost.db-wal openpost-backup-$(date +%Y%m%d).db-wal 2>/dev/null || true
cp /var/lib/openpost/openpost.db-shm openpost-backup-$(date +%Y%m%d).db-shm 2>/dev/null || true
```

### Media

```bash
tar -czf media-backup-$(date +%Y%m%d).tar.gz /var/lib/openpost/media/
```

### Secrets

```bash
cp /opt/openpost/.env openpost-env-backup-$(date +%Y%m%d)
```

Restart when the backup finishes:

```bash
sudo systemctl start openpost
```

## Restore process

1. Stop OpenPost.
2. Restore the SQLite database files.
3. Restore the media directory.
4. Restore `.env` or the equivalent secrets source.
5. Fix ownership and permissions.
6. Start OpenPost.
7. Confirm login, media access, and scheduled-post visibility.

### Example restore

```bash
sudo systemctl stop openpost
sudo mkdir -p /var/lib/openpost/media /opt/openpost
sudo cp openpost-backup-20260518.db /var/lib/openpost/openpost.db
sudo cp openpost-backup-20260518.db-wal /var/lib/openpost/openpost.db-wal 2>/dev/null || true
sudo cp openpost-backup-20260518.db-shm /var/lib/openpost/openpost.db-shm 2>/dev/null || true
sudo tar -xzf media-backup-20260518.tar.gz -C /
sudo cp openpost-env-backup-20260518 /opt/openpost/.env
sudo chown -R openpost:openpost /var/lib/openpost /opt/openpost
sudo chmod 600 /opt/openpost/.env
sudo systemctl start openpost
```

## Migrate to another server

1. Install the new OpenPost binary or container deployment first.
2. Stop OpenPost on both the old and new server.
3. Copy the database, any `-wal` and `-shm` files, the media directory, and `.env`.
4. Restore ownership and permissions on the new server.
5. Start OpenPost on the new server.
6. Verify provider callbacks, media URLs, and scheduled posts before switching traffic.

If the hostname changes, update your reverse proxy, provider callback URLs, and `OPENPOST_MEDIA_URL` before making the new server live.

## Test restore checklist

- Can you log in with an existing account?
- Do previously uploaded media items load?
- Are connected accounts still listed?
- Are drafts and scheduled posts present?
- Does the readiness endpoint respond?
- If the server hostname changed: do provider callbacks and public media URLs still point at the new host?

## Notes

- Test restores, not just backups.
- Keep database and media snapshots reasonably aligned in time.
- Protect backup copies of `.env`: encrypted provider tokens still depend on `OPENPOST_ENCRYPTION_KEY`.
