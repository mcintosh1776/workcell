# Sentinel Review Task: Security and Threat Model

Owner: Sentinel
Status: ready for review

## Context

Workcell is a self-hosted disposable workspace runner. It will execute commands
submitted by trusted developers and trusted internal agents. v0.1 is not a
public hostile-code sandbox.

Read these files before reviewing:

- `README.md`
- `AGENTS.md`
- `docs/product-brief.md`
- `docs/security-model.md`
- `docs/threat-model.md`
- `docs/backend-interface.md`
- `docs/api.md`
- `docs/test-strategy.md`

## Review Focus

Evaluate whether the security model is explicit enough before implementation.

Focus on:

- accepted v0.1 risks
- Incus container versus Incus VM boundary
- Podman rootless expectations
- host mount policy
- secret/env forwarding rules
- API token model
- artifact/log leakage
- cross-job access
- cleanup failure as a security event
- whether anything is unsafe to publish as OSS guidance

## Required Output

Write findings to:

```text
docs/reviews/sentinel-security-threat-model-review.md
```

Use this structure:

```md
# Sentinel Review: Security and Threat Model

## Must Fix Before Implementation

## Required Security Tests

## Accepted v0.1 Risks

## Later Hardening

## Open Questions
```

Be findings-first. If a risk is acceptable for v0.1, say why and state the
boundary clearly.

