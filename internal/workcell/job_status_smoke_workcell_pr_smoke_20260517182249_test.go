package workcell

import (
  "testing"
)

func TestJobStatusSmokeLineworkcell_pr_smoke_20260517182249IncludesStableFields(t *testing.T) {
  job := Job{
    ID:       "job-123",
    State:    JobSucceeded,
    Profile:  "fake",
    Backend:  "fake",
    ExitCode: 0,
  }
  got := JobStatusSmokeLineworkcell_pr_smoke_20260517182249(job)
  want := "job=job-123 state=succeeded profile=fake backend=fake exitCode=0"
  if got != want {
    t.Fatalf("JobStatusSmokeLineworkcell_pr_smoke_20260517182249() = %q, want %q", got, want)
  }
}

func TestJobStatusSmokeLineworkcell_pr_smoke_20260517182249IncludesZeroValues(t *testing.T) {
  got := JobStatusSmokeLineworkcell_pr_smoke_20260517182249(Job{})
  want := "job= state= profile= backend= exitCode=0"
  if got != want {
    t.Fatalf("JobStatusSmokeLineworkcell_pr_smoke_20260517182249(empty) = %q, want %q", got, want)
  }
}