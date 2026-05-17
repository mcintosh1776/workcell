package workcell

import (
	"testing"
)

func TestJobStatusLineIncludesStableFields(t *testing.T) {
	job := Job{
		ID:       "job-123",
		State:    JobSucceeded,
		Profile:  "fake",
		Backend:  "fake",
		ExitCode: 0,
	}
	got := JobStatusLine(job)
	want := "job=job-123 state=succeeded profile=fake backend=fake exitCode=0"
	if got != want {
		t.Fatalf("JobStatusLine() = %q, want %q", got, want)
	}
}

func TestJobStatusLineIncludesZeroValues(t *testing.T) {
	got := JobStatusLine(Job{})
	want := "job= state= profile= backend= exitCode=0"
	if got != want {
		t.Fatalf("JobStatusLine(empty) = %q, want %q", got, want)
	}
}
