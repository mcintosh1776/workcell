# Test Strategy

## Testing goals

- Job lifecycle is deterministic.
- Cleanup is reliable.
- Backends are replaceable.
- Secrets are not leaked by default.
- Logs and artifacts are bounded and inspectable.
- Failure modes are visible.

## Test layers

### Unit tests

- job request validation
- profile resolution
- state transitions
- timeout handling
- log truncation
- artifact path validation
- env allowlist filtering
- error response shape

### Fake backend integration tests

Use a fake backend to prove runner behavior without requiring Incus or Podman.

Scenarios:

- success
- command failure
- prepare failure
- timeout
- cancellation
- artifact collection
- cleanup success
- cleanup failure

### Real backend smoke tests

Podman:

```bash
workcell run --profile podman-smoke -- echo hello
```

Incus:

```bash
workcell run --profile incus-smoke -- echo hello
```

### Security tests

- env values are not printed in status
- disallowed env vars are not forwarded
- artifact paths cannot escape workspace
- logs are truncated at configured limit
- cleanup runs after command failure
- cleanup runs after timeout

## QA review questions

- Which cleanup failure modes are missing?
- Which race conditions should be tested before v0.1?
- Do we need concurrency tests in the first implementation slice?
- What is the minimum real Incus smoke for release confidence?
- What is the minimum real Podman smoke for release confidence?

