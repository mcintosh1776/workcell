# Steve And Linus Build Plan

## Rule for Steve and Linus

Do not build beyond the current slice. Do not add UI, cloud provisioning,
ARX-specific integration, or broad scheduling unless the spec is updated first.

Steve owns the overall implementation plan. Linus may implement bounded
engineering slices when the write set is clear and does not overlap Steve's
active slice.

## Slice 001: project skeleton and job model

Goal: create a compilable/testable foundation without real backend execution.

Deliver:

- CLI skeleton
- runner daemon skeleton
- job request/response model
- profile model
- in-memory job lifecycle
- fake backend
- unit tests for validation and state transitions

Acceptance criteria:

- `workcell run --profile fake -- echo hello` creates a fake job and returns a
  successful result.
- Invalid profile returns structured error.
- Command array is required and cannot be empty.
- Fake backend success/failure paths are tested.

## Slice 002: filesystem job store and logs

Deliver:

- local job directory layout
- persisted request/status
- stdout/stderr log files
- bounded log reads
- cleanup metadata

Acceptance criteria:

- completed job can be inspected after daemon restart.
- logs are bounded.
- status includes cleanup state.

## Slice 003: Podman backend smoke

Deliver:

- Podman adapter
- profile config for a smoke image
- run/destroy lifecycle
- smoke script

Acceptance criteria:

- `workcell run --profile podman-smoke -- echo hello` runs in Podman.
- container is removed after job completion.
- command failure returns non-zero exit code without daemon failure.

## Slice 004: Incus backend smoke

Deliver:

- Incus adapter
- profile config for Ubuntu image
- launch/exec/delete lifecycle
- smoke script

Acceptance criteria:

- `workcell run --profile incus-smoke -- echo hello` runs in Incus.
- instance is deleted after job completion.
- startup failure is recorded as job failure with cleanup attempted.

## Slice 005: artifacts and env allowlist

Deliver:

- artifact path collection
- env allowlist enforcement
- status metadata listing env names only
- path traversal protection

Acceptance criteria:

- configured artifact files are collected.
- artifact paths cannot escape workspace.
- disallowed env vars are not forwarded.
- secret values do not appear in status output.

## Slice 006: timeout, cancel, and cleanup

Deliver:

- timeout enforcement
- cancel endpoint/CLI
- idempotent backend destroy
- cleanup command

Acceptance criteria:

- long-running job times out.
- canceled job is cleaned up.
- cleanup can be run twice safely.
- cleanup failure is visible.
