# Steve Handoff: Podman Backend Implementation

## Role

Writer.

## Assignment

Implement the bounded Podman backend smoke described in
`docs/steve-podman-backend-task.md`.

## Repo

```text
https://github.com/mcintosh1776/workcell
```

Branch from:

```text
main
```

## Required Reading

- `BUILD.md`
- `docs/backend-interface.md`
- `docs/implementation-slices/002-podman-backend.md`
- `docs/steve-podman-backend-task.md`
- `docs/cost-guardrails.md`

## Hard Boundaries

- implement Podman only
- keep fake backend working
- do not implement Incus
- do not expose a public API port
- do not add ARX-specific code
- do not add cloud provisioning behavior to the runner
- do not use privileged containers
- do not mount the Docker or Podman socket into jobs

## Expected Proof

On the disposable lab host, not Gondor:

```bash
go test ./...
workcell run --profile fake -- echo hello
workcell run --profile podman-smoke -- echo hello
workcell run --profile podman-smoke -- false
```

The Podman container must be removed after each job.

## Output

Open a PR with:

- summary
- files changed
- commands run
- cleanup proof
- risks or deferred items
