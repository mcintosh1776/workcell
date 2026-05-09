# Steve Task: Podman Backend Smoke

## Assignment

Implement only the first Podman backend smoke path.

## Starting Point

Slice 001 provides:

- Go CLI skeleton
- fake backend behavior
- shared job model
- in-memory runner
- lab host preflight scripts

## Scope

- add a backend interface in Go if needed
- keep fake backend working
- add Podman backend implementation
- support `podman-smoke` profile
- run a command in `docker.io/library/alpine:3.20` or equivalent
- remove the container after completion
- preserve structured errors

## Do Not Build

- no Incus backend
- no auth
- no public API exposure
- no provider provisioning
- no UI
- no ARX-specific integration
- no privileged containers
- no Docker socket mounting

## Acceptance Criteria

- `go test ./...` passes.
- fake backend tests still pass.
- `workcell run --profile fake -- echo hello` still works.
- `workcell run --profile podman-smoke -- echo hello` succeeds.
- `workcell run --profile podman-smoke -- false` returns a failed job without
  crashing.
- the Podman container is removed after each job.

## Preferred Proof Host

Use the disposable Workcell lab host, not Gondor.
