# API

## Principles

- HTTP JSON API.
- Token authenticated.
- Backend-neutral job model.
- No ARX-specific concepts.
- Job request is immutable after submission except cancellation.

## Endpoints

```http
GET  /v1/health
POST /v1/jobs
GET  /v1/jobs
GET  /v1/jobs/{jobId}
GET  /v1/jobs/{jobId}/logs
GET  /v1/jobs/{jobId}/artifacts
POST /v1/jobs/{jobId}/cancel
POST /v1/admin/cleanup
```

Admin endpoints may be deferred until after the basic runner exists.

## Submit job

```http
POST /v1/jobs
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "profile": "ubuntu-systemd",
  "command": ["make", "test"],
  "workspace": {
    "type": "upload"
  },
  "timeoutSeconds": 1800,
  "env": {
    "allow": ["CI", "NODE_OPTIONS"]
  },
  "artifacts": {
    "paths": ["coverage", "test-results"]
  }
}
```

Response:

```json
{
  "id": "job_01hxyz",
  "state": "queued",
  "createdAt": "2026-05-09T16:00:00Z"
}
```

## Job status

```json
{
  "id": "job_01hxyz",
  "state": "succeeded",
  "profile": "ubuntu-systemd",
  "backend": "incus",
  "exitCode": 0,
  "createdAt": "2026-05-09T16:00:00Z",
  "startedAt": "2026-05-09T16:00:04Z",
  "finishedAt": "2026-05-09T16:01:12Z",
  "cleanup": {
    "state": "complete"
  },
  "logs": {
    "stdoutBytes": 12000,
    "stderrBytes": 900
  },
  "artifacts": {
    "count": 2,
    "bytes": 42000
  }
}
```

## Logs

`GET /v1/jobs/{jobId}/logs` returns bounded logs.

Open question for v0.1:

- return combined logs by default, or separate stdout/stderr fields?

Recommended default:

```json
{
  "stdout": "...",
  "stderr": "...",
  "truncated": false
}
```

## Error shape

```json
{
  "ok": false,
  "error": {
    "code": "invalid_profile",
    "message": "Profile does not exist: ubuntu-systemd"
  }
}
```

