# Backend Interface

## Goal

Keep the runner daemon backend-neutral. Incus and Podman should implement the
same lifecycle contract.

## Concepts

### Job

User-submitted command plus profile, workspace, timeout, env policy, and artifact
request.

### Workspace

Prepared source tree or uploaded directory visible inside the backend instance.

### Instance

Backend-specific runtime object:

- Incus container or VM
- Podman container

### Artifact

File or directory collected after command completion.

## Interface

Pseudo-interface:

```text
Prepare(job, profile) -> handle
Start(handle) -> stream
Wait(handle) -> result
CollectArtifacts(handle) -> artifactSummary
Cancel(handle) -> cancelResult
Destroy(handle) -> cleanupResult
Inspect(handle) -> backendStatus
```

The fake backend, Podman backend, and Incus backend must all satisfy this same
shape. Backend-specific code should not leak into the CLI command parser or API
route handlers.

## Required behavior

Every backend must:

- create an isolated instance per job
- attach the prepared workspace
- run the command as a non-root user where practical
- stream stdout/stderr
- enforce timeout
- collect configured artifacts
- destroy the instance
- make destroy idempotent

## Incus backend v0.1

Required:

- launch container from image/profile
- copy or mount workspace
- execute command
- collect logs
- pull artifacts
- delete instance

Deferred:

- Incus VM mode
- snapshots
- custom networks
- cluster support
- image publishing
- GUI/desktop sessions

## Podman backend v0.1

Required:

- create container from image
- mount workspace
- execute command
- collect logs
- copy artifacts
- remove container
- prove command success and failure with `podman-smoke`
- report cleanup failure instead of hiding it

Deferred:

- compose/pod orchestration
- privileged mode
- Docker socket mounting
- host networking by default

See [Implementation Slice 002](implementation-slices/002-podman-backend.md) for
the first bounded Podman implementation task.
