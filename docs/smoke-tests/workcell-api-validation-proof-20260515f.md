# Workcell Validation API Proof

This proof documents the Workcell validation API behavior that ARX uses when a repository target selects the Workcell validation worker backend.

## Product behavior assertion

Workcell exposes a backend-neutral validation job contract at `POST /v1/validation-jobs` and stores completed results for `GET /v1/validation-jobs/{validationJobId}`. The request accepts repository identity, base/head refs, `headSha`, a validation profile, command list, timeout, network policy, mutation policy, working directory, and git-bundle source metadata. The response returns the validation job id, status, head SHA, validation profile, dirty tracked files, exit code, worker backend, mutation mode, and output artifact paths.

## Validation guardrails checked by review

- Local validation execution is disabled unless the operator explicitly sets `WORKCELL_ENABLE_LOCAL_VALIDATION_EXECUTION=1` for a trusted daemon deployment.
- The HTTP validation API requires `WORKCELL_VALIDATION_API_TOKEN_FILE` and bearer-token authentication.
- Token files must be regular files outside repositories and job workspaces, readable only by the daemon user.
- Git-bundle validation requests require `sourceBundleSha256`; Workcell verifies the digest before cloning and checking out `headSha`.
- Successful validation fails closed if tracked files are left dirty after commands complete.

## Review evidence

Relevant implementation and documentation files on `main`:

- `docs/api.md`
- `internal/workcell/validation.go`
- `internal/workcell/runner_test.go`
- `cmd/workcell/main.go`
- `cmd/workcell/main_test.go`

Repository validation for this PR is expected to run through the Workcell-backed `workcell-go-test` profile with `go test ./...`.