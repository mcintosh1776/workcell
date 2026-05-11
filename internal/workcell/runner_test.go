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
	if job.Error != "" {
		t.Fatalf("Error = %q, want empty for successful fake job", job.Error)
	}
	if job.Stdout != "hello\n" {
		t.Fatalf("Stdout = %q, want %q", job.Stdout, "hello\n")
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
	if job.Error != "" {
		t.Fatalf("Error = %q, want empty for command exit failure", job.Error)
	}
}

func TestRunnerFakeProfileRejectsUnsupportedCommand(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"date"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobFailed {
		t.Fatalf("State = %s, want %s", job.State, JobFailed)
	}
	if job.Error == "" {
		t.Fatal("Error is empty, want fake backend validation detail")
	}
}

func TestRunnerBackendFailurePreservesBackendExitAndError(t *testing.T) {
	runner := NewRunner(map[string]Profile{
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
			BackendConfig: BackendConfig{
				Image:   "example.invalid/missing:latest",
				Timeout: 60,
			},
		},
	})
	runner.backends["podman"] = backendFunc{
		run: func(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
			return 125, "", "image pull failed", &BackendError{Op: "create", Err: errors.New("image pull failed")}
		},
		cleanup: func(ctx context.Context, job Job, profile Profile) error {
			return nil
		},
	}

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "podman-smoke",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobFailed {
		t.Fatalf("State = %s, want %s", job.State, JobFailed)
	}
	if job.ExitCode != 125 {
		t.Fatalf("ExitCode = %d, want 125", job.ExitCode)
	}
	if job.Error == "" {
		t.Fatal("Error is empty, want backend failure detail")
	}
	if job.Logs.StderrBytes == 0 {
		t.Fatal("StderrBytes = 0, want captured backend stderr")
	}
	if job.Stderr != "image pull failed" {
		t.Fatalf("Stderr = %q, want backend stderr detail", job.Stderr)
	}
}

func TestRunnerCleanupUsesFreshContextAfterRunContextExpires(t *testing.T) {
	runner := NewRunner(map[string]Profile{
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
			BackendConfig: BackendConfig{
				Image:   "docker.io/library/alpine:3.20",
				Timeout: 60,
			},
		},
	})
	runner.backends["podman"] = backendFunc{
		run: func(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
			if cancel, ok := ctx.Value(cancelContextKey{}).(context.CancelFunc); ok {
				cancel()
			}
			return 0, "ok", "", nil
		},
		cleanup: func(ctx context.Context, job Job, profile Profile) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, cancelContextKey{}, context.CancelFunc(cancel))

	job, err := runner.Run(ctx, SubmitJobRequest{
		Profile: "podman-smoke",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobSucceeded {
		t.Fatalf("State = %s, want %s", job.State, JobSucceeded)
	}
	if job.Cleanup.State != "complete" {
		t.Fatalf("Cleanup.State = %s, want complete", job.Cleanup.State)
	}
}

type backendFunc struct {
	run     func(context.Context, Job, Profile) (int, string, string, error)
	cleanup func(context.Context, Job, Profile) error
}

func (backend backendFunc) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	return backend.run(ctx, job, profile)
}

func (backend backendFunc) Cleanup(ctx context.Context, job Job, profile Profile) error {
	return backend.cleanup(ctx, job, profile)
}

type cancelContextKey struct{}
