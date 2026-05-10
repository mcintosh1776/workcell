# Build And Lab Setup

## Purpose

This document explains how to build and validate Workcell without polluting your
daily workstation.

Workcell is a runner for disposable build environments. The safest development
pattern is:

```text
developer workstation -> SSH -> disposable lab host -> Podman/Incus jobs
```

Do not install the full runner stack on a workstation just to test Workcell if a
cheap disposable VPS will do.

## Local Development Requirements

For code editing and fake-backend tests:

- Go 1.22 or newer
- `gofmt`
- `jq`

The `fake` profile is the safest local CLI smoke path because it validates the
command model without starting Podman or Incus workloads.

Run:

```bash
go test ./...
scripts/dev-smoke.sh
```

Expected:

```text
workcell_dev_smoke=ok
```

## Disposable Lab Host Requirements

For Podman and Incus validation:

- Ubuntu LTS
- amd64
- 4 GB RAM minimum
- 40 GB disk minimum for initial smokes
- private SSH access
- no public Workcell API port

The first tested shape is a Hetzner `cx23`:

```text
2 vCPU, 4 GB RAM, 40 GB disk
```

If real build jobs need more room, rebuild the lab host larger instead of
turning your workstation into the runner.

## Lab Host Bootstrap

On the lab host:

```bash
git clone https://github.com/mcintosh1776/workcell.git /opt/workcell
cd /opt/workcell
scripts/bootstrap-ubuntu-lab-host.sh --yes
incus admin init --minimal
```

Then run the preflight:

```bash
scripts/lab-host-preflight.sh
```

For live Podman and Incus container smoke jobs:

```bash
WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 scripts/lab-host-preflight.sh
```

Expected:

```text
workcell_lab_preflight=ok
```

## Optional Hetzner Provisioning Helper

If you use Hetzner Cloud, the guarded helper can create a small lab host and a
firewall that allows SSH only from one source CIDR.

Prerequisites on the operator machine:

- `curl`
- `jq`
- `HCLOUD_TOKEN`
- an existing Hetzner SSH key id

Example:

```bash
HCLOUD_TOKEN=... WORKCELL_LAB_SSH_KEY_ID=123456 WORKCELL_LAB_SSH_SOURCE_CIDR="$(curl -sS https://api.ipify.org)/32" scripts/provision-hetzner-lab-host.sh --yes
```

The helper creates:

- a firewall named `workcell-lab-001-firewall`
- a server named `workcell-lab-001`
- labels identifying it as a disposable Workcell lab host

The helper does not:

- expose the Workcell API
- install packages
- upload secrets
- configure ARX-specific infrastructure

After provisioning, SSH to the host and follow the lab bootstrap above.

## Remote Preflight From Operator Machine

Once a lab host is bootstrapped:

```bash
WORKCELL_LAB_SSH_HOST=<host> WORKCELL_LAB_SSH_KEY=<path-to-key> scripts/run-remote-lab-preflight.sh
```

To include live Podman and Incus container smoke jobs:

```bash
WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 WORKCELL_LAB_SSH_HOST=<host> WORKCELL_LAB_SSH_KEY=<path-to-key> scripts/run-remote-lab-preflight.sh
```

## Teardown

The lab host is disposable. If it becomes messy, rebuild it.

For Hetzner:

```bash
HCLOUD_TOKEN=... scripts/destroy-hetzner-lab-host.sh --yes
```

By default the teardown helper deletes the server only. Set
`WORKCELL_LAB_DELETE_FIREWALL=1` if the SSH-only firewall should also be
removed.