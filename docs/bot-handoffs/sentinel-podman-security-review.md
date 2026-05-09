# Sentinel Handoff: Podman Backend Security Review

## Role

Review-only.

## Assignment

Review the Podman backend smoke plan for host escape, secret exposure, and
dangerous defaults.

## Required Reading

- `docs/security-model.md`
- `docs/threat-model.md`
- `docs/cost-guardrails.md`
- `docs/backend-interface.md`
- `docs/implementation-slices/002-podman-backend.md`
- `docs/steve-podman-backend-task.md`

## Review Focus

- workspace mount policy
- default user inside the container
- whether network should be enabled for `podman-smoke`
- forbidden Podman flags
- secret/env allowlist behavior
- whether host paths can be escaped
- cleanup failure visibility
- risks accepted for v0.1

## Output Format

- must-fix security blockers
- acceptable v0.1 risks
- forbidden flags or patterns
- tests Sentinel expects before merge
- later hardening items
