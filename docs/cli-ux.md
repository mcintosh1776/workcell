# CLI UX

## Goals

- Make the common path obvious.
- Hide backend complexity behind profiles.
- Make failure states readable.
- Make setup and diagnostics less scary, especially for Incus.

## Commands

```bash
workcell doctor
workcell init
workcell run --profile ubuntu-systemd -- make test
workcell status job_123
workcell logs job_123
workcell artifacts job_123
workcell cancel job_123
workcell cleanup --dry-run
```

## First-run flow

```bash
workcell init --backend incus
workcell doctor
workcell run --profile default -- echo hello
```

Podman-friendly path:

```bash
workcell init --backend podman
workcell run --profile node-fast -- npm test
```

## Output style

Example:

```text
job job_01hxyz queued profile=ubuntu-systemd backend=incus
preparing workspace...
starting instance wc-job-01hxyz...
running: make test
...
exit=0 duration=68s cleanup=complete artifacts=2
```

## Error message rules

Errors should say:

- what failed
- what Workcell expected
- the next command to diagnose

Bad:

```text
backend failed
```

Good:

```text
Incus profile 'ubuntu-systemd' could not launch image images:ubuntu/26.04.
Run: workcell doctor --backend incus
```

## Iris review questions

- Is `workcell run --profile ... -- <command>` intuitive?
- Does Incus setup feel approachable?
- Are status/log/artifact commands discoverable?
- Should `run` stream logs by default?
- Should `artifacts` download or only print paths by default?

