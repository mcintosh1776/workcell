# Changelog

## Unreleased

- Add initial Workcell specification package.
- Retain in-memory job history and bounded stdout/stderr logs through the
  runner API.
- Add `GET /v1/jobs` and `GET /v1/jobs/{jobId}/logs` API surfaces.
- Add synchronous validation job API support for git-bundle source handoffs,
  command execution, dirty tracked file detection, and stored validation
  results.
