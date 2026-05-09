# Security Model

## Trust model

v0.1 is for trusted operators, trusted developers, and trusted internal agents.
It is not a public sandbox for arbitrary hostile users.

Assumptions:

- submitted commands can be destructive inside their workspace
- job code may be buggy or semi-hostile
- the build host is controlled by the operator
- users with API tokens are trusted to submit jobs within policy
- backend isolation reduces risk but does not eliminate it

## Isolation levels

### Incus container

Best for machine-like Linux workspaces with stronger lifecycle controls than a
single process container. Suitable for trusted internal validation and systemd
style tests.

### Incus VM

Future higher-isolation mode for riskier workloads. Not required for v0.1 but
the profile model should not block it.

### Podman container

Best for fast command checks. Rootless Podman should be preferred where
possible. Suitable for trusted repo tests and lightweight agent validation.

## Secret policy

Default: no secrets.

Secrets or environment values may only enter jobs through explicit allowlists.

Rules:

- never accept secret values on command-line flags
- never print secret values in logs
- never mount broad host secret directories
- redact known secret-looking strings where practical
- make env forwarding visible in the job request/status metadata without
  recording values

## Host filesystem policy

Jobs should receive a prepared workspace and explicit artifact directories only.

Forbidden by default:

- mounting `/`
- mounting `/home`
- mounting `/var/run/docker.sock`
- mounting host SSH keys
- privileged containers
- host networking unless a profile explicitly enables it

## Authentication

v0.1 should support static bearer tokens for the runner API.

Future auth may include:

- OIDC
- mTLS
- per-user tokens
- per-project tokens

## Authorization

v0.1 roles:

- user: submit jobs, read own jobs, cancel own jobs
- admin: inspect all jobs, force cleanup, manage profiles

If ownership is not implemented in v0.1, the API must document that tokens are
operator-level and not user-isolated.

## Cleanup requirements

Every backend must implement:

- timeout cleanup
- cancellation cleanup
- startup failure cleanup
- orphan detection
- explicit `cleanup` command or admin endpoint

Cleanup failure is a reportable security event.

