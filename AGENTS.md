# AGENTS: Working rules for Workcell

## Golden rules

- Keep Workcell standalone and open-source friendly.
- Do not add ARX-specific code paths, bot names, tenant concepts, or private
  infrastructure assumptions.
- Prefer small, reviewable slices.
- Treat security and cleanup as product features, not afterthoughts.
- Do not forward secrets by default.
- Do not mount broad host paths into jobs.
- Do not add UI, cloud provisioning, or multi-node scheduling before the v0.1
  CLI/API/runner contract is stable.

## Product boundary

Workcell exposes a generic API and CLI. ARX may consume that API later, but the
project must remain useful without ARX.

## Backend stance

- Incus is the primary backend for machine-like workspaces.
- Podman is supported for lightweight command containers.
- The public API should talk about jobs, profiles, workspaces, logs, and
  artifacts, not backend internals.

## Required review loop

Before implementation:

- Iris reviews CLI/API/documentation ergonomics.
- Inspector reviews test strategy and failure modes.
- Sentinel reviews threat model and isolation boundaries.

Implementation should proceed in small Steve/Linus-owned slices after the spec
package is accepted. Steve should own the primary slice plan; Linus can take
bounded engineering slices when the write scope is clear and non-overlapping.

## Validation discipline

Each implementation slice should include tests for the behavior it adds.
If a backend integration cannot run in the local environment, provide a dry-run
or fake-backend test seam and document the manual command that proves the real
backend behavior.
