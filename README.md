# Workcell

Disposable workspaces for developers and AI agents.

Workcell is a self-hosted execution runner for clean, repeatable command runs on
your own build machine. It is Incus-first for machine-like workspaces and also
supports Podman for fast container checks.

## Problem

Developers and coding agents need a place to run builds, tests, and validation
without polluting local machines, waiting on full CI, or giving every tool direct
cloud/provider credentials.

## Core promise

Bring a Linux build host. Workcell turns it into an API-driven disposable
workspace runner:

```text
CLI/API -> runner daemon -> Incus or Podman workspace -> logs/artifacts -> cleanup
```

## v0.1 scope

- one runner daemon
- one build host
- token-authenticated HTTP API
- CLI for submitting jobs and reading logs/status
- Incus backend
- Podman backend
- profile-based backend selection
- bounded logs
- filesystem artifacts
- mandatory cleanup

## Non-goals

- not a CI replacement
- not Kubernetes
- not an ARX-specific subsystem
- not a multi-cloud provisioner
- not a public untrusted-code platform in v0.1

## Documentation

- [Build and lab setup](BUILD.md)
- [Product brief](docs/product-brief.md)
- [Architecture](docs/architecture.md)
- [Deployment targets](docs/deployment-targets.md)
- [Lab host bootstrap](docs/lab-host-bootstrap.md)
- [Lab host lifecycle](docs/lab-host-lifecycle.md)
- [Cost guardrails](docs/cost-guardrails.md)
- [Security model](docs/security-model.md)
- [Threat model](docs/threat-model.md)
- [API](docs/api.md)
- [Backend interface](docs/backend-interface.md)
- [CLI UX](docs/cli-ux.md)
- [Test strategy](docs/test-strategy.md)
- [Bot review plan](docs/bot-review-plan.md)
- [Steve build plan](docs/steve-build-plan.md)
- [Steve kickoff task](docs/steve-kickoff-task.md)
- [Steve Podman backend task](docs/steve-podman-backend-task.md)

Current bot handoffs live under [docs/bot-handoffs](docs/bot-handoffs).

## Current development baseline

Workcell is a Go CLI/daemon skeleton. The only implemented backend is `fake`,
which exists to prove the job model, command validation, API shape, and test
harness before Podman or Incus are wired in.

Prerequisites:

- Go 1.22 or newer
- `jq` for `scripts/dev-smoke.sh`

```bash
go test ./...
go run ./cmd/workcell run --profile fake -- echo hello
```

For a local smoke:

```bash
scripts/dev-smoke.sh
```

For the disposable VPS lab-host path, see
[Lab host bootstrap](docs/lab-host-bootstrap.md). Do not install the full
Podman/Incus runner stack on Gondor just to validate Workcell.
