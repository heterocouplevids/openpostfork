# Single Binary

OpenPost can run as a single Go binary with the frontend embedded into the executable.

## 1. Download a release

Download the binary for your platform from [GitHub Releases](https://github.com/rodrgds/openpost/releases).

Expected release assets:

- Linux x86_64: `openpost-server-linux-amd64`
- macOS Apple Silicon: `openpost-server-darwin-arm64`

## 2. Create `.env`

Create a working directory and write a complete `.env` file:

```dotenv
OPENPOST_PORT=8080
OPENPOST_DATABASE_PATH=/var/lib/openpost/openpost.db
OPENPOST_MEDIA_PATH=/var/lib/openpost/media
OPENPOST_MEDIA_URL=https://social.example.com/media

OPENPOST_JWT_SECRET=replace-with-a-random-secret-at-least-32-characters-long
OPENPOST_ENCRYPTION_KEY=replace-with-a-random-secret-at-least-32-characters-long

# Optional but commonly useful
OPENPOST_DISABLE_REGISTRATIONS=false

# Example provider config
# X_CLIENT_ID=
# X_CLIENT_SECRET=
# MASTODON_SERVERS='[{"name":"Personal","client_id":"...","client_secret":"...","instance_url":"https://mastodon.social"}]'
# LINKEDIN_CLIENT_ID=
# LINKEDIN_CLIENT_SECRET=
# THREADS_CLIENT_ID=
# THREADS_CLIENT_SECRET=
```

## 3. Prepare production paths

```bash
sudo mkdir -p /var/lib/openpost/media
sudo chown -R $(whoami) /var/lib/openpost
```

Recommended production locations:

- Database: `/var/lib/openpost/openpost.db`
- Media: `/var/lib/openpost/media`

## 4. Make it executable

```bash
chmod +x ./openpost
```

## 5. Run it

```bash
./openpost
```

By default, OpenPost listens on `http://localhost:8080`.

## 6. Run it with systemd

Example unit:

```ini
[Unit]
Description=OpenPost
After=network.target

[Service]
Type=simple
User=openpost
Group=openpost
WorkingDirectory=/opt/openpost
EnvironmentFile=/opt/openpost/.env
ExecStart=/opt/openpost/openpost
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Example install layout:

- Binary: `/opt/openpost/openpost`
- Environment file: `/opt/openpost/.env`
- Database: `/var/lib/openpost/openpost.db`
- Media: `/var/lib/openpost/media`

After creating the unit:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now openpost
sudo systemctl status openpost
```

## 7. Upgrade safely

1. Back up the database, media directory, and `.env` file first.
2. Stop the service: `sudo systemctl stop openpost`
3. Replace the binary with the new release asset.
4. Confirm ownership and execute permissions.
5. Start the service: `sudo systemctl start openpost`
6. Check logs and the health endpoint before considering the upgrade complete.

## Backup reminder

Do not upgrade without a restorable backup. See [Backups](/operations/backups).

## Notes

- Put the service behind HTTPS before enabling production OAuth callbacks.
- Protect the `.env` file because `OPENPOST_ENCRYPTION_KEY` is required to decrypt stored provider tokens.
