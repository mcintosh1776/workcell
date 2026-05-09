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

- [Product brief](docs/product-brief.md)
- [Architecture](docs/architecture.md)
- [Security model](docs/security-model.md)
- [Threat model](docs/threat-model.md)
- [API](docs/api.md)
- [Backend interface](docs/backend-interface.md)
- [CLI UX](docs/cli-ux.md)
- [Test strategy](docs/test-strategy.md)
- [Bot review plan](docs/bot-review-plan.md)
- [Steve build plan](docs/steve-build-plan.md)

