package workcell

import (
  "fmt"
)

// JobStatusLine returns a compact deterministic status line for one job.
func JobStatusLine(job Job) string {
  return fmt.Sprintf(
    "job=%s state=%s profile=%s backend=%s exitCode=%d",
    job.ID,
    job.State,
    job.Profile,
    job.Backend,
    job.ExitCode,
  )
}