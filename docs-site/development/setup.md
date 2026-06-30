# Development Setup

## Clone the repository

```bash
git clone https://github.com/rodrgds/openpost.git
cd openpost
```

## Frontend

```bash
pnpm install
pnpm --filter @openpost/web dev
```

Frontend dev server: `http://localhost:5173`

## Backend

```bash
cd ../backend
cp .env.example .env
go run ./cmd/openpost
```

Backend server: `http://localhost:8080`

## Docs site

From the repo root:

```bash
pnpm run sync:assets
pnpm --filter @openpost/docs docs:dev
```

Docs site: `http://localhost:4174`
