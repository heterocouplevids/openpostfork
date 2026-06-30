# Production Checklist

- [ ] Copy the root `.env.example` to `.env` and remove placeholder values
- [ ] Generate fresh `OPENPOST_JWT_SECRET`
- [ ] Generate fresh `OPENPOST_ENCRYPTION_KEY`
- [ ] Use secrets that are at least 32 characters long
- [ ] Set `OPENPOST_APP_URL` to the public HTTPS app origin
- [ ] Set `OPENPOST_PUBLIC_URL` to the public HTTPS app origin
- [ ] Set `OPENPOST_MEDIA_URL` to the public HTTPS media base URL
- [ ] Decide whether to set `OPENPOST_DISABLE_REGISTRATIONS=true` after creating the first admin account
- [ ] Configure reverse proxy with HTTPS
- [ ] Update provider callback URLs for X, LinkedIn, and Threads
- [ ] Configure Mastodon servers in `MASTODON_SERVERS` if you need Mastodon
- [ ] Persist `/data`
- [ ] Back up database, media, and secrets together
- [ ] Check `GET /api/v1/ready`
- [ ] Create a test draft and scheduled post before relying on automation
