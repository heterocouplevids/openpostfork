# Health Checks

OpenPost exposes separate liveness and readiness endpoints.

## Liveness

```txt
GET /api/v1/health
```

Expected response:

```json
{"status":"ok"}
```

Use this endpoint when you only need to know whether the HTTP process is alive.

## Readiness

```txt
GET /api/v1/ready
```

Expected response:

```json
{"status":"ready","database":"ok"}
```

Use this endpoint for container health checks, deploy rollouts, and external uptime probes that should fail when the database is unavailable. It returns `503` when OpenPost cannot run a database probe.
