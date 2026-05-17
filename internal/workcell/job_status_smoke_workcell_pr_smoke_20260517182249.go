package workcell

import (
  "fmt"
)

// JobStatusSmokeLineworkcell_pr_smoke_20260517182249 returns a compact deterministic smoke status line for one job.
func JobStatusSmokeLineworkcell_pr_smoke_20260517182249(job Job) string {
  return fmt.Sprintf(
    "job=%s state=%s profile=%s backend=%s exitCode=%d",
    job.ID,
    job.State,
    job.Profile,
    job.Backend,
    job.ExitCode,
  )
}