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
POST /v1/validation-jobs
GET  /v1/validation-jobs/{validationJobId}
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

```json
{
  "stdout": "...",
  "stderr": "...",
  "truncated": false
}
```

## Validation jobs

`POST /v1/validation-jobs` accepts the generic validation-worker contract used
by callers that need a clean checkout, a profile name, a command list, and a
readiness result. The first implementation is synchronous and intended for
trusted local daemon use while the isolated backend matures.

Local command execution is disabled by default. Operators must set
`WORKCELL_ENABLE_LOCAL_VALIDATION_EXECUTION=1` only for trusted local daemon
deployments until validation jobs run under an isolated backend profile.

When `sourceTransport` is `git-bundle`, `sourceBundlePath` must point at a git
bundle visible to the Workcell daemon, and `sourceBundleSha256` is required.
Workcell verifies the digest before cloning the bundle into a job-local
workspace, checks out `headSha`, runs each command from `workingDirectory` or
the checkout root, and fails the validation if tracked files are left dirty.

```json
{
  "validationJobId": "validation_01hxyz",
  "repository": "owner/repo",
  "repoUrl": "https://github.com/owner/repo.git",
  "baseRef": "main",
  "headRef": "feature",
  "headSha": "abc123",
  "validationProfile": "unit",
  "commands": ["go test ./..."],
  "timeoutSeconds": 900,
  "networkPolicy": "disabled",
  "mutation": "none",
  "sourceTransport": "git-bundle",
  "sourceBundlePath": "/var/lib/workcell/requests/validation_01hxyz.bundle",
  "sourceBundleSha256": "..."
}
```

Response:

```json
{
  "validationJobId": "validation_01hxyz",
  "status": "succeeded",
  "headSha": "abc123",
  "validationProfile": "unit",
  "dirtyTrackedFiles": [],
  "exitCode": 0,
  "workerBackend": "workcell",
  "mutation": "none"
}
```

`GET /v1/validation-jobs/{validationJobId}` returns the stored result for a
completed validation job.

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
