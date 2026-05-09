# Architecture

## System overview

```text
developer/agent
    |
    | CLI or HTTP API
    v
workcell runner daemon
    |
    | backend adapter
    v
Incus or Podman workspace
    |
    | logs, exit code, artifacts
    v
local artifact store
```

## Components

### CLI

The CLI is the normal user entrypoint. It submits jobs, streams logs, checks
status, downloads artifacts, and runs doctor checks.

### Runner daemon

The daemon owns:

- HTTP API
- job validation
- queueing
- profile resolution
- backend adapter selection
- log capture
- artifact collection
- timeout/cancel handling
- cleanup

### Backend adapters

Backends provide isolated execution. v0.1 supports:

- Incus
- Podman

The API should not expose backend-specific lifecycle details except through
profile selection and diagnostic metadata.

### Profiles

Profiles are named execution environments. A profile selects:

- backend
- image
- mode
- resource limits
- workspace behavior
- artifact paths
- allowed env vars
- default timeout

Example:

```yaml
profiles:
  ubuntu-systemd:
    backend: incus
    mode: container
    image: images:ubuntu/26.04
    timeoutSeconds: 1800

  node-fast:
    backend: podman
    image: node:24-bookworm
    timeoutSeconds: 900
```

## Job lifecycle

```text
queued -> preparing -> running -> collecting -> succeeded
                                      |
                                      -> failed
                                      -> canceled
                                      -> timed_out
                                      -> cleanup_failed
```

Cleanup should run after every terminal state. If cleanup fails, the job result
must preserve the original command outcome and also record cleanup failure.

## Storage layout

Suggested local layout:

```text
/var/lib/workcell/
  config/
  jobs/
    job_abc123/
      request.json
      status.json
      stdout.log
      stderr.log
      artifacts/
      backend.json
  tmp/
```

## ARX integration boundary

ARX should call Workcell like any other client:

```text
POST /v1/jobs
GET /v1/jobs/{id}
GET /v1/jobs/{id}/logs
GET /v1/jobs/{id}/artifacts
```

Workcell must not depend on ARX tenants, bots, queues, OpenBao, Telegram, or
runtime paths.

