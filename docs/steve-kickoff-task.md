# Steve Kickoff Task: Workcell Slice 001

## Assignment

Create the first runnable Workcell MVP without implementing Podman or Incus yet.

## Scope

Use the current Go skeleton as the boundary:

- CLI command: `workcell run --profile fake -- echo hello`
- HTTP endpoint: `POST /v1/jobs`
- shared job request/response model
- fake backend only
- in-memory job store only
- tests for validation and fake lifecycle

## Do Not Build Yet

- no Podman backend
- no Incus backend
- no Hetzner provisioning
- no authentication
- no persistent job store
- no artifact collection
- no UI
- no ARX-specific code

## Deployment Assumption

The first real host target is a single Ubuntu VPS on Hetzner. Do not encode
Hetzner-specific behavior into the core runner. Keep the runner portable and put
provider-specific setup in docs or later provisioning scripts.

## Acceptance Criteria

- Go 1.22 or newer is available in the development environment.
- `go test ./...` passes.
- `scripts/dev-smoke.sh` passes.
- invalid profile returns `invalid_profile`.
- empty command returns `invalid_command`.
- fake backend records successful and failed jobs.

## Expected PR Shape

Small PR only. Prefer tightening the existing skeleton over expanding scope.
If a bigger design issue appears, update `docs/implementation-slices/001-skeleton.md`
instead of building past it.
