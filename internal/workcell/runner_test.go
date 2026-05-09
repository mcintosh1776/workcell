package workcell

import (
	"context"
	"errors"
	"testing"
)

func TestRunnerFakeProfileSucceeds(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobSucceeded {
		t.Fatalf("State = %s, want %s", job.State, JobSucceeded)
	}
	if job.Backend != "fake" {
		t.Fatalf("Backend = %s, want fake", job.Backend)
	}
	if job.Cleanup.State != "complete" {
		t.Fatalf("Cleanup.State = %s, want complete", job.Cleanup.State)
	}
}

func TestRunnerRejectsInvalidProfile(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	_, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "missing",
		Command: []string{"echo", "hello"},
	})
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("error = %v, want ErrInvalidProfile", err)
	}
}

func TestRunnerRejectsEmptyCommand(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	_, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
	})
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("error = %v, want ErrInvalidCommand", err)
	}
}

func TestRunnerFakeProfileCanFail(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"false"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobFailed {
		t.Fatalf("State = %s, want %s", job.State, JobFailed)
	}
	if job.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", job.ExitCode)
	}
}
