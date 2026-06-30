# Database

OpenPost uses SQLite by default. The backend now has an explicit database-driver setting so hosted deployments can opt into Postgres without changing self-hosted defaults.

## Default path

The backend code defaults to:

```txt
file:openpost.db?cache=shared&mode=rwc
```

For container deployments, prefer an explicit file path such as:

```txt
/data/db/openpost.db
```

## Operational notes

- Persist the database on durable storage.
- Back up the database together with the media directory.
- Do not keep the database inside ephemeral container layers.
- SQLite is configured for a simple single-node deployment model.

## Driver settings

```sh
OPENPOST_DATABASE_DRIVER=sqlite
OPENPOST_DATABASE_PATH=file:openpost.db?cache=shared&mode=rwc
```

For Postgres-backed deployments:

```sh
OPENPOST_DATABASE_DRIVER=postgres
OPENPOST_DATABASE_URL=postgres://openpost:secret@db.internal:5432/openpost?sslmode=require
```

Postgres is intended for OpenPost Cloud and larger hosted deployments. SQLite remains the recommended self-hosted path until the hosted migration and operational runbooks are complete.
