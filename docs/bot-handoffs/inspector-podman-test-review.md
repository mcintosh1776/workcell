# Inspector Handoff: Podman Backend Test Review

## Role

Review-only.

## Assignment

Define the minimum test strategy for the Podman backend smoke slice.

## Required Reading

- `docs/test-strategy.md`
- `docs/backend-interface.md`
- `docs/implementation-slices/002-podman-backend.md`
- `docs/steve-podman-backend-task.md`
- `scripts/lab-host-preflight.sh`
- `scripts/run-remote-lab-preflight.sh`

## Review Focus

- unit tests that do not require Podman
- integration tests that require a lab host
- how to prove container cleanup
- how to prove command failure does not crash the runner
- how to detect leaked containers
- how much should be in `scripts/dev-smoke.sh` versus lab-only scripts
- what artifact should be produced for a successful remote preflight

## Output Format

- must-fix test gaps
- recommended test names
- required smoke commands
- cleanup/leak proof strategy
- deferred test cases
