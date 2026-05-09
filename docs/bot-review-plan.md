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

### Inspector

Reviews:

- test strategy
- cleanup coverage
- backend failure modes
- concurrency risks
- log/artifact evidence quality
- release smoke requirements

### Sentinel

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

### Linus

Implements bounded engineering slices when the spec package identifies a clear
non-overlapping write scope. Linus should not invent new subsystems or change the
public API without a spec update.

### Sentinel

Reviews Steve PRs and checks whether implementation evidence matches the spec.

## Review order

1. Codex drafts spec package.
2. Iris, Inspector, and Sentinel review spec package.
3. Codex resolves spec feedback.
4. Steve implements slice 001, optionally with Linus on a separate bounded
   write scope.
5. Sentinel reviews implementation PRs.

## Current handoffs

- Steve: `docs/bot-handoffs/steve-podman-implementation.md`
- Linus: `docs/bot-handoffs/linus-podman-review.md`
- Iris: `docs/bot-handoffs/iris-build-cli-review.md`
- Inspector: `docs/bot-handoffs/inspector-podman-test-review.md`
- Sentinel: `docs/bot-handoffs/sentinel-podman-security-review.md`

## Reviewer output format

Each review should return:

- must-fix before implementation
- acceptable-for-v0.1 risks
- later hardening items
- concrete test or doc changes
