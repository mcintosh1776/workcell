# Linus Handoff: Podman Backend Engineering Review

## Role

Review-only unless explicitly assigned a separate implementation slice.

## Assignment

Review the planned Podman backend implementation shape for Go/process-management
issues before Steve expands the runner.

## Required Reading

- `docs/backend-interface.md`
- `docs/implementation-slices/002-podman-backend.md`
- `docs/steve-podman-backend-task.md`
- `internal/workcell/runner.go`
- `cmd/workcell/main.go`

## Review Focus

- whether the backend interface is clean enough
- whether command execution and exit-code handling should be refactored before
  Podman is added
- whether `context.Context` cancellation is wired at the right layer
- whether stdout/stderr collection needs a concrete model before Podman
- whether cleanup state is expressive enough
- whether the fake backend should become a real backend implementation behind
  an interface

## Output Format

- must-fix before Podman implementation
- acceptable-for-v0.1 risks
- recommended small refactor, if any
- files Steve should avoid touching unnecessarily
