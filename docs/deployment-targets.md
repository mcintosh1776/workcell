# Deployment Targets

## Primary v0.1 Target: Hetzner VPS

Workcell v0.1 should target a single Ubuntu VPS on Hetzner.

Why:

- Arx Labs already operates on Hetzner.
- VPS provisioning is cheap and repeatable.
- Operators can create and destroy test hosts without owning spare bare metal.
- Podman works normally on a VPS.
- Incus containers work normally on a VPS.
- Incus virtual machines are not required for v0.1.

Recommended first host shape:

- Ubuntu LTS
- amd64
- at least 4 GB RAM
- at least 40 GB disk for fake and first backend smokes
- at least 80 GB disk if running larger real build jobs
- private SSH access only
- no public Workcell API until auth, rate limits, and network policy are real

The first disposable lab host is described in
[Lab host bootstrap](lab-host-bootstrap.md). Gondor should remain the operator
workstation, not the Workcell build/test runner.

## Backend Support Order

### 1. Fake Backend

Purpose:

- prove CLI shape
- prove API shape
- prove job lifecycle
- prove tests
- give Steve and Linus a safe first slice

### 2. Podman Backend

Purpose:

- fast container smoke jobs
- easy adoption for users who know Docker-like workflows
- lower setup fear than Incus

Expected use:

```bash
workcell run --profile podman-smoke -- echo hello
```

### 3. Incus Container Backend

Purpose:

- more machine-like workspaces
- cleaner fit for multi-step builds and service-like test environments
- stronger long-term fit for Arx Labs infrastructure work

Expected use:

```bash
workcell run --profile incus-smoke -- echo hello
```

## Explicit v0.1 Non-Targets

- bare-metal-only assumptions
- AWS-specific provisioning
- Kubernetes
- public multi-tenant untrusted execution
- Incus VM workloads requiring nested virtualization
- direct Arx Labs integration

## Later Targets

Other VPS providers should be possible once the single-host runner contract is
stable. The project should not bake in Hetzner-specific assumptions beyond
documentation and example provisioning notes.

Potential later providers:

- DigitalOcean
- Linode/Akamai
- Vultr
- AWS EC2
- local lab machines

The portable unit is the Workcell runner host, not the cloud provider.
