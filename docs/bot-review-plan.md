# Bot Review Plan

## Purpose

Use ARX bots as reviewers without making Workcell depend on ARX.

## Roles

### Codex

Owns the initial product and architecture spec package.

### Iris

Reviews:

- CLI ergonomics
- first-run flow
- setup docs
- error messages
- whether Incus feels too intimidating
- whether Podman fallback is discoverable

### QA

Reviews:

- test strategy
- cleanup coverage
- backend failure modes
- concurrency risks
- log/artifact evidence quality
- release smoke requirements

### Security

Reviews:

- trust model
- isolation boundaries
- secret handling
- host mount policy
- API token model
- accepted v0.1 risks
- abuse cases that require tests

### Steve

Implements only approved slices from `docs/steve-build-plan.md`.

### Sentinel

Reviews Steve PRs and checks whether implementation evidence matches the spec.

## Review order

1. Codex drafts spec package.
2. Iris, QA, and Security review spec package.
3. Codex resolves spec feedback.
4. Steve implements slice 001.
5. Sentinel reviews Steve's PR.

## Reviewer output format

Each review should return:

- must-fix before implementation
- acceptable-for-v0.1 risks
- later hardening items
- concrete test or doc changes

