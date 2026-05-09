# Implementation Slice 001: CLI, Daemon, Job Model, Fake Backend

Owner: Steve
Collaborator: Linus, only if assigned a non-overlapping write scope
Status: blocked until Iris, Inspector, and Sentinel reviews are captured

## Goal

Create a minimal, testable Workcell foundation without real Incus or Podman
execution.

This slice should prove the public shape of the tool before backend complexity
is introduced.

## In Scope

- project skeleton
- CLI command structure
- runner daemon skeleton
- job request/response model
- profile model
- fake backend
- in-memory job lifecycle
- structured error shape
- unit tests for validation and state transitions

## Out Of Scope

- real Incus execution
- real Podman execution
- web UI
- cloud provisioning
- multi-node scheduling
- ARX-specific integration
- secret manager integration
- persistent job store

## Required Behavior

The first runnable path should be:

```bash
workcell run --profile fake -- echo hello
```

Expected behavior:

- command is accepted
- fake backend records the command
- job transitions through a deterministic lifecycle
- exit code is reported
- logs are available through the same abstraction that real backends will use

## Acceptance Criteria

- Invalid profile returns structured `invalid_profile`.
- Empty command returns structured `invalid_command`.
- Fake backend success path is tested.
- Fake backend failure path is tested.
- Job lifecycle cannot skip directly from `queued` to `succeeded`.
- Cancellation of a non-running fake job is deterministic.
- Public API shape matches `docs/api.md` unless the spec is updated first.
- No ARX-specific names, paths, or assumptions appear in implementation code.

## Suggested Split

Steve should own:

- job model
- CLI skeleton
- fake backend success/failure path

Linus may own, if assigned:

- structured error module
- lifecycle transition tests
- profile validation tests

Do not let both bots edit the same files in the same slice unless explicitly
coordinated.

## Review Gate

Before implementation starts, incorporate must-fix feedback from:

- `docs/reviews/iris-cli-api-docs-review.md`
- `docs/reviews/inspector-test-strategy-review.md`
- `docs/reviews/sentinel-security-threat-model-review.md`

