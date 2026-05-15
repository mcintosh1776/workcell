# Changelog

## Unreleased

- Add initial Workcell specification package.
- Retain in-memory job history and bounded stdout/stderr logs through the
  runner API.
- Add `GET /v1/jobs` and `GET /v1/jobs/{jobId}/logs` API surfaces.
- Add synchronous validation job API support for git-bundle source handoffs,
  command execution, dirty tracked file detection, and stored validation
  results.
- Require explicit trusted-local opt-in for synchronous validation command
  execution and require `sourceBundleSha256` for git-bundle validation jobs.
- Require a configured bearer token for validation job HTTP endpoints, preferring
  `WORKCELL_VALIDATION_API_TOKEN_FILE` instead of direct environment-token
  storage.
- Document validation API token-file ownership, placement, provisioning,
  workspace isolation, and rotation requirements.
- Enforce validation API token files as regular owner-only permission files
  before reading their contents.
