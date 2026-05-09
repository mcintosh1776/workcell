# Iris Review Task: CLI, API, and Documentation UX

Owner: Iris
Status: ready for review

## Context

Workcell is a standalone open-source project for disposable execution
workspaces. It is Incus-first and Podman-supported. ARX may consume it later
through the public API, but Workcell must remain useful without ARX.

Read these files before reviewing:

- `README.md`
- `AGENTS.md`
- `docs/product-brief.md`
- `docs/architecture.md`
- `docs/api.md`
- `docs/cli-ux.md`
- `docs/backend-interface.md`

## Review Focus

Evaluate whether a developer can understand and try Workcell without already
being an Incus expert.

Focus on:

- CLI command names and arguments
- first-run setup flow
- whether profiles hide enough backend complexity
- whether Incus and Podman positioning is clear
- whether error-message guidance is actionable
- whether README and docs explain what Workcell is not

## Required Output

Write findings to:

```text
docs/reviews/iris-cli-api-docs-review.md
```

Use this structure:

```md
# Iris Review: CLI, API, and Docs

## Must Fix Before Implementation

## Acceptable For v0.1

## Later UX Improvements

## Suggested Doc Edits

## Open Questions
```

Keep findings concrete. Prefer specific wording and command changes over broad
advice.

