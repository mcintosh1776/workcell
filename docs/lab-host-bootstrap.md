# Lab Host Bootstrap

## Purpose

Create the first disposable Workcell lab host without loading Gondor with build
runner dependencies.

Gondor should remain the operator workstation:

- edit code
- commit and push repos
- connect to lab hosts
- run lightweight orchestration commands

Gondor should not become the Workcell build/test runner:

- no local Incus lab
- no local Podman runner lab
- no long-lived Workcell daemon
- no polluted workstation dependency stack

## First Lab Host

Name:

```text
workcell-lab-001
```

Provider:

```text
Hetzner Cloud
```

Recommended shape:

- Ubuntu LTS
- amd64
- at least 4 GB RAM
- at least 40 GB disk for fake and first backend smokes
- private SSH access
- public Workcell API disabled

## Host Responsibilities

The lab host is allowed to contain:

- Go toolchain
- Podman
- Incus
- Workcell repo checkout
- temporary build/test workspaces
- disposable smoke artifacts

The lab host is not allowed to contain:

- ARX production secrets
- OpenBao root/operator tokens
- podcast/audio setup
- operator workstation state
- long-lived customer data

## Manual Bootstrap Steps

These are intentionally manual until Workcell proves its first runner contract.

### 1. Create VPS

Create a Hetzner Cloud server named `workcell-lab-001`.

Suggested starting size:

```text
cx23
```

The first lab host used `cx23`: 2 vCPU, 4 GB RAM, and 40 GB disk. If Incus
or real build workloads outgrow that shape, rebuild the disposable host as a
larger instance instead of tuning Gondor.

Use the normal operator SSH key. Keep firewall exposure minimal:

- SSH from trusted operator IPs or private access path
- no public Workcell API port
- no broad `0.0.0.0/0` SSH rule unless it is an explicit temporary recovery
  exception

### 2. Install Base Packages

On the lab host:

```bash
sudo apt update
sudo apt install -y git curl jq build-essential ca-certificates
```

Or, after cloning Workcell, run the guarded helper:

```bash
scripts/bootstrap-ubuntu-lab-host.sh --yes
```

Do not run this helper on Gondor.

### 3. Install Go

Install Go from the OS package or official tarball. The project target is:

```text
Go 1.22 or newer
```

Verify:

```bash
go version
command -v gofmt
```

The `gofmt` command only needs to exist for the first lab readiness check.

### 4. Install Podman

```bash
sudo apt install -y podman
podman run --rm docker.io/library/alpine:3.20 echo podman-ok
```

### 5. Install Incus

Prefer the official Incus package path for the chosen Ubuntu release.

After install:

```bash
incus version
incus admin init
incus launch images:ubuntu/24.04 workcell-smoke
incus exec workcell-smoke -- echo incus-ok
incus delete --force workcell-smoke
```

For v0.1, Incus containers are enough. Do not require Incus VMs or nested
virtualization.

### 6. Clone Workcell

```bash
git clone https://github.com/mcintosh1776/workcell.git
cd workcell
```

### 7. Run Slice 001 Proof

```bash
go test ./...
scripts/dev-smoke.sh
```

Expected:

```text
workcell_dev_smoke=ok
```

### 8. Run Lab Preflight

From the Workcell checkout:

```bash
scripts/lab-host-preflight.sh
```

To include live Podman and Incus container smoke jobs:

```bash
WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 scripts/lab-host-preflight.sh
```

Expected:

```text
workcell_lab_preflight=ok
```

## Success Criteria

The lab host is ready when:

- SSH access works
- Go and `gofmt` exist
- Podman smoke passes
- Incus container smoke passes
- Workcell `go test ./...` passes
- Workcell `scripts/dev-smoke.sh` passes
- Workcell `scripts/lab-host-preflight.sh` passes

## Teardown Rule

The lab host should be considered disposable. If the host gets messy, rebuild it
instead of nursing it back to health.

Workcell should eventually automate enough of this lifecycle that the lab host
itself becomes a product proof.
