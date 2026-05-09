# Cost Guardrails

## Goal

Make Workcell cheap to test and hard to accidentally leave running.

## Default Lab Budget Posture

Use the smallest host that proves the current slice.

Current tested minimum:

```text
Hetzner cx23
2 vCPU
4 GB RAM
40 GB disk
```

Do not start with a large build host just because later workloads might need it.
Rebuild larger only when a real test proves the small host is insufficient.

## Required Labels

Every provider-created lab host should be labeled:

```text
project=workcell
purpose=lab
environment=dev
```

Provider-specific labels make it possible to find and clean up dangling hosts.

## TTL Rule

Every temporary lab host should have an intended lifetime before it is created.

Suggested defaults:

- `same-day`: exploratory testing
- `one-week`: active backend slice
- `manual`: only when a human explicitly wants the host retained

Workcell does not yet enforce TTL automatically. Until it does, the operator is
responsible for teardown.

## Network Cost Rule

Do not expose public runner APIs by default.

Avoid:

- public Workcell API listeners
- public Incus API listeners
- broad SSH rules
- unnecessary snapshots or backups
- large artifact uploads

## Cleanup Rule

When a lab host is no longer actively useful:

```bash
scripts/destroy-hetzner-lab-host.sh --yes
```

If in doubt, destroy and rebuild. Reproducibility is part of the product.

## Future Product Requirement

Before Workcell creates hosts automatically, it needs:

- max active host count
- max host age
- max provider spend estimate
- required labels
- cleanup on failed bootstrap
- cleanup on idle timeout
- visible cost estimate in status output
