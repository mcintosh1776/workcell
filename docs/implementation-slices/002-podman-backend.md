# Implementation Slice 002: Podman Backend Smoke

Owner: Steve or Linus with non-overlapping write scope
Status: ready after slice 001 fake backend is stable

## Goal

Run one command in a disposable Podman container and always remove the
container afterward.

## In Scope

- profile field for Podman image
- `podman-smoke` example profile
- create container
- mount prepared workspace read/write
- run command
- capture exit code
- capture stdout/stderr
- remove container
- surface cleanup state

## Out Of Scope

- Incus
- pods or compose
- privileged containers
- Docker socket mounting
- host networking by default
- artifact collection beyond a minimal placeholder
- remote API authentication

## Required Command Shape

```bash
workcell run --profile podman-smoke -- echo hello
```

Expected:

- backend is `podman`
- state is `succeeded`
- exit code is `0`
- cleanup state is `complete`
- container no longer exists after the job finishes

Failure path:

```bash
workcell run --profile podman-smoke -- false
```

Expected:

- backend is `podman`
- state is `failed`
- exit code is non-zero
- cleanup state is `complete`
- daemon or CLI does not crash

## Cleanup Contract

Podman cleanup must be idempotent.

If command execution fails, cleanup still runs.

If cleanup fails, job status must make that visible instead of silently reporting
success.

## Lab Proof

Run on a disposable lab host, not on a workstation:

```bash
WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 scripts/lab-host-preflight.sh
```

Then run the Podman profile proof once implemented:

```bash
go test ./...
workcell run --profile podman-smoke -- echo hello
workcell run --profile podman-smoke -- false
```
