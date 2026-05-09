# Threat Model

## Assets

- build host integrity
- runner daemon token
- submitted repository contents
- forwarded environment values
- job logs
- job artifacts
- backend images/profiles
- other jobs' workspaces and artifacts

## Actors

- trusted operator
- trusted developer
- trusted internal agent
- buggy job code
- malicious job code submitted by an authorized user
- remote attacker without an API token

## Threats

### Job escapes workspace

Risk: job code reads or mutates host files.

Mitigations:

- no broad host mounts
- no privileged containers by default
- backend-specific isolation defaults
- per-job workspace directories
- cleanup and orphan scanning

### Secret leakage

Risk: secrets appear in logs, artifacts, process args, or status files.

Mitigations:

- no secrets by default
- env allowlist by name
- never pass secret values as CLI args
- bounded redaction filters
- status records list env names only

### Cross-job data access

Risk: one job reads another job's workspace or artifacts.

Mitigations:

- per-job directories
- restrictive filesystem permissions
- backend instance per job
- API ownership checks when implemented

### Resource exhaustion

Risk: jobs consume CPU, memory, disk, process count, network, or storage.

Mitigations:

- profile resource limits
- job timeout
- artifact size limits
- log size limits
- queue concurrency limit
- admin cleanup

### Runner API abuse

Risk: unauthorized user submits jobs or reads logs.

Mitigations:

- bearer token auth in v0.1
- no unauthenticated routes except health
- admin token separation when admin endpoints exist
- bind to private interface by default

### Cleanup failure

Risk: stopped jobs leave containers, instances, mounts, or files behind.

Mitigations:

- cleanup on every terminal path
- orphan scanner
- idempotent destroy
- visible `cleanup_failed` state

## v0.1 accepted risks

- authorized users can run arbitrary commands in job workspaces
- v0.1 is not safe for public untrusted workloads
- static bearer tokens are acceptable for early self-hosted operation
- Incus container isolation is not equivalent to a fresh VM

## Security bot review questions

- Are the v0.1 accepted risks stated clearly enough?
- Which backend defaults are unsafe?
- Which host mounts must be forbidden by policy?
- What must be tested before the first public release?
- Should Incus VM mode be required before any semi-untrusted workloads?

