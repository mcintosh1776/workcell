# Inspector Review Task: Test Strategy and Failure Modes

Owner: Inspector
Status: ready for review

## Context

Workcell will run user-supplied commands in disposable Incus or Podman
workspaces. The first implementation must be testable before real backend
execution exists.

Read these files before reviewing:

- `README.md`
- `AGENTS.md`
- `docs/product-brief.md`
- `docs/architecture.md`
- `docs/backend-interface.md`
- `docs/api.md`
- `docs/test-strategy.md`
- `docs/steve-build-plan.md`

## Review Focus

Evaluate whether the v0.1 plan has enough test coverage to avoid fragile runner
behavior.

Focus on:

- job lifecycle state transitions
- fake backend coverage
- real Incus smoke requirements
- real Podman smoke requirements
- timeout and cancellation behavior
- cleanup failure behavior
- log truncation
- artifact path validation
- env allowlist behavior
- concurrency risks

## Required Output

Write findings to:

```text
docs/reviews/inspector-test-strategy-review.md
```

Use this structure:

```md
# Inspector Review: Test Strategy and Failure Modes

## Must Fix Before Implementation

## Required v0.1 Tests

## Acceptable For v0.1

## Later Hardening

## Open Questions
```

Find concrete missing tests and failure modes. Do not ask for a broad CI system
or multi-node scheduler.

