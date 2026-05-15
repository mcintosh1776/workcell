# Workcell

Disposable workspaces for developers and AI agents.

## Quick Start

```bash
# Build and run from source with the deterministic fake profile
git clone https://github.com/mcintosh1776/workcell.git
cd workcell
go build -o workcell ./cmd/workcell
./workcell run --profile fake -- echo hello
```

> [!WARNING]
> The example above uses the `fake` profile for development smoke tests only.
> The fake profile is deterministic scaffolding. It does not spawn a process,
> shell, container, or VM. It returns simulated output derived from the submitted
> command text so the CLI, job model, output capture, and cleanup contract can be
> tested before Podman or Incus execution is enabled.

See [BUILD.md](BUILD.md) for local build and [docs/lab-host-bootstrap.md](docs/lab-host-bootstrap.md) for deployment.

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

Workcell is a Go CLI/daemon skeleton with a synchronous runner API. The `fake`
backend exists to prove the job model, command validation, API shape, output
capture, and test harness; it is deterministic scaffolding, not a real command
execution backend. `podman-smoke` is available for bounded container validation
on a lab host. Jobs are retained in memory with bounded stdout/stderr retrieval
through the API.

The repository also serves as a small end-to-end smoke target for the Steve,
QA, and Sentinel bot review workflow.

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
