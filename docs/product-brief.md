# Product Brief

## One-sentence pitch

Workcell runs commands in clean, disposable workspaces on your own build host.

## Target users

- developers who want a faster, cleaner loop than full CI
- maintainers with cheap local or rented build machines
- AI coding agents that need a safe place to run tests
- teams that want repeatable validation without giving every tool cloud
  credentials

## Primary job

Given a repository checkout and a command, run that command in a clean workspace,
capture logs and artifacts, report the result, then clean up.

## Why now

AI coding agents make execution isolation more important. CI is often too slow
for tight iteration. Local machines become polluted and stateful. Cheap build
boxes are underused. Workcell sits between local execution and full CI.

## Positioning

Workcell is not a CI platform. It is a disposable workspace runner.

It should feel like:

```bash
workcell run --profile arx-systemd -- make test
workcell logs job_123
workcell artifacts job_123
```

## v0.1 promise

One host can accept jobs, run each job in an Incus or Podman workspace, capture
evidence, and clean up reliably.

## v0.1 non-goals

- no hosted service
- no web dashboard
- no multi-node scheduler
- no arbitrary tenant model
- no cloud VM provisioning
- no GitHub Actions replacement
- no secret forwarding except explicit env allowlists
- no public untrusted-code guarantee

## Design principles

- Profiles hide backend complexity.
- Cleanup is mandatory.
- Secrets are opt-in and allowlisted.
- Logs are bounded and redacted where practical.
- Jobs are immutable after submission except cancellation.
- Backend adapters are replaceable.
- ARX integration happens through the public API only.

