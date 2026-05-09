# Lab Host Lifecycle

## Purpose

Keep Workcell lab hosts disposable, cheap, and auditable.

The lab host exists to validate Workcell without turning a workstation into the
runner. It should be easy to create, bootstrap, verify, destroy, and rebuild.

## Lifecycle

```text
create -> bootstrap -> preflight -> run experiments -> teardown or rebuild
```

## Create

Use a small VPS first.

Recommended first shape:

```text
Hetzner cx23
Ubuntu 24.04 LTS
2 vCPU
4 GB RAM
40 GB disk
```

Use labels:

```text
project=workcell
purpose=lab
environment=dev
```

Expose only SSH from a trusted source CIDR. Do not expose the Workcell API in
v0.1 lab hosts.

## Bootstrap

On the lab host:

```bash
git clone https://github.com/mcintosh1776/workcell.git /opt/workcell
cd /opt/workcell
scripts/bootstrap-ubuntu-lab-host.sh --yes
incus admin init --minimal
```

## Preflight

From the lab host:

```bash
cd /opt/workcell
scripts/lab-host-preflight.sh
```

For live container backend checks:

```bash
WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 scripts/lab-host-preflight.sh
```

From an operator workstation:

```bash
WORKCELL_LAB_SSH_HOST=<host-ip-or-name> \
WORKCELL_LAB_SSH_KEY=<path-to-key> \
scripts/run-remote-lab-preflight.sh
```

## Rebuild Rule

Prefer rebuild over repair when:

- package state gets messy
- Incus storage/network configuration becomes unclear
- test containers are stuck
- disk fills unexpectedly
- a backend experiment needs a clean baseline

The lab host is not precious infrastructure.

## Teardown

Use the provider console or the guarded helper:

```bash
HCLOUD_TOKEN=... \
WORKCELL_LAB_NAME=workcell-lab-001 \
scripts/destroy-hetzner-lab-host.sh --yes
```

The helper deletes the server by name. Firewall deletion is optional and must be
requested separately.

## Records

Do not commit live host IPs or provider tokens.

Operator-local records can live outside the repo, for example:

```text
~/.codex/memories/workcell-lab-001.env
```
