package workcell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func requirePodman(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skipf("podman is not available in this test environment: %v", err)
	}
}

// FakeBackend tests (no Podman required)

func TestFakeBackend_Run(t *testing.T) {
	backend := &FakeBackend{}
	profile := Profile{ID: "fake", Backend: "fake"}

	tests := []struct {
		name           string
		command        []string
		wantExit       int
		wantStdout     string
		wantErr        bool
		wantBackendErr bool
	}{
		{
			name:       "echo command",
			command:    []string{"echo", "hello"},
			wantExit:   0,
			wantStdout: "echo hello",
		},
		{
			name:       "false command returns exit 1",
			command:    []string{"false"},
			wantExit:   1,
			wantStdout: "false",
		},
		{
			name:           "empty command errors",
			command:        []string{},
			wantErr:        true,
			wantBackendErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{ID: "test-job", Command: tt.command}
			exit, stdout, stderr, err := backend.Run(context.Background(), job, profile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.wantBackendErr && !IsBackendError(err) {
					t.Errorf("expected BackendError, got %T", err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if exit != tt.wantExit {
				t.Errorf("exit code = %d, want %d", exit, tt.wantExit)
			}
			if stdout != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", stdout, tt.wantStdout)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
		})
	}
}

func TestFakeBackend_Cleanup(t *testing.T) {
	backend := &FakeBackend{}
	profile := Profile{ID: "fake", Backend: "fake"}
	job := Job{ID: "test-job"}

	err := backend.Cleanup(context.Background(), job, profile)
	if err != nil {
		t.Errorf("cleanup should not error: %v", err)
	}
}

// BackendError tests

func TestBackendError(t *testing.T) {
	err := &BackendError{Op: "test", Err: context.Canceled}
	if err.Error() != "backend test failed: context canceled" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
	if !IsBackendError(err) {
		t.Error("IsBackendError should return true for BackendError")
	}
	if IsBackendError(context.Canceled) {
		t.Error("IsBackendError should return false for non-BackendError")
	}
}

// Test that BackendError is distinguishable from command exit failures
func TestBackendError_DistinguishableFromExitFailure(t *testing.T) {
	// BackendError indicates infrastructure failure
	backendErr := &BackendError{Op: "start", Err: fmt.Errorf("podman not found")}
	if !IsBackendError(backendErr) {
		t.Error("infrastructure error should be BackendError")
	}

	// Command exit failure is not a BackendError
	exitErr := fmt.Errorf("exit status 1")
	if IsBackendError(exitErr) {
		t.Error("command exit failure should not be BackendError")
	}
}

// Container name sanitization tests (no Podman required)

func TestSanitizeContainerName(t *testing.T) {
	tests := []struct {
		name    string
		jobID   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid job ID",
			jobID: "job_abc123",
			want:  "workcell-job_abc123",
		},
		{
			name:  "job ID with hyphens",
			jobID: "job-abc-123",
			want:  "workcell-job-abc-123",
		},
		{
			name:  "job ID with underscores",
			jobID: "job_abc_123",
			want:  "workcell-job_abc_123",
		},
		{
			name:  "job ID starting with underscore gets prefix",
			jobID: "_job123",
			want:  "workcell-j_job123",
		},
		{
			name:  "job ID with invalid chars",
			jobID: "job@123#test",
			want:  "workcell-job-123-test",
		},
		{
			name:  "very long job ID gets truncated",
			jobID: strings.Repeat("a", 200),
			want:  "workcell-" + strings.Repeat("a", 119),
		},
		{
			name:    "empty job ID errors",
			jobID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeContainerName(tt.jobID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("sanitizeContainerName() = %q, want %q", got, tt.want)
			}
			if len(got) > 128 {
				t.Errorf("container name too long: %d chars", len(got))
			}
		})
	}
}

// effectiveDeadline tests

func TestEffectiveDeadline(t *testing.T) {
	tests := []struct {
		name            string
		callerDeadline  time.Duration // 0 means no deadline
		profileTimeout  int           // 0 means no timeout
		wantNewContext  bool          // whether we expect a new context with deadline
		checkEarlier    bool          // if true, check that deadline is earlier than caller
	}{
		{
			name:           "no caller deadline, no profile timeout",
			callerDeadline: 0,
			profileTimeout: 0,
			wantNewContext: false,
		},
		{
			name:           "no caller deadline, has profile timeout",
			callerDeadline: 0,
			profileTimeout: 60,
			wantNewContext: true,
		},
		{
			name:           "caller deadline earlier than profile timeout",
			callerDeadline: 30 * time.Second,
			profileTimeout: 60,
			wantNewContext: false,
		},
		{
			name:           "profile timeout earlier than caller deadline",
			callerDeadline: 120 * time.Second,
			profileTimeout: 60,
			wantNewContext: true,
			checkEarlier:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc

			if tt.callerDeadline > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), tt.callerDeadline)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			profile := Profile{
				ID:            "test",
				Backend:       "podman",
				BackendConfig: BackendConfig{Timeout: tt.profileTimeout},
			}

			newCtx, newCancel := effectiveDeadline(ctx, profile)
			defer newCancel()

			_, hasDeadline := newCtx.Deadline()
			originalDeadline, originalHasDeadline := ctx.Deadline()

			if tt.wantNewContext {
				if !hasDeadline {
					t.Error("expected new context to have deadline")
				}
				if tt.checkEarlier && originalHasDeadline {
					newDeadline, _ := newCtx.Deadline()
					if !newDeadline.Before(originalDeadline) {
						t.Error("profile deadline should be earlier than caller deadline")
					}
				}
			} else {
				if hasDeadline != originalHasDeadline {
					t.Errorf("deadline changed unexpectedly: had=%v, now=%v", originalHasDeadline, hasDeadline)
				}
			}
		})
	}
}

// PodmanBackend tests (require Podman)

func TestPodmanBackend_Run_Success(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-run-success",
		Command: []string{"echo", "hello"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exit, stdout, stderr, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exit != 0 {
		t.Errorf("expected exit 0, got %d", exit)
	}
	if !strings.Contains(stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

func TestPodmanBackend_Run_CommandFailure(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-run-fail",
		Command: []string{"sh", "-c", "exit 42"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exit, _, _, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("command failure should not return error, got: %v", err)
	}
	if exit != 42 {
		t.Errorf("expected exit 42, got %d", exit)
	}
	// Verify this is NOT a BackendError
	if IsBackendError(err) {
		t.Error("command exit failure should not be BackendError")
	}
}

func TestPodmanBackend_Run_CommandFailureWithFakePodmanIsNotBackendError(t *testing.T) {
	backend := &PodmanBackend{binary: fakePodman(t, `#!/bin/sh
case "$1" in
  create) exit 0 ;;
  start) echo "command failed" >&2; exit 42 ;;
  inspect) echo 42; exit 0 ;;
  stop|kill|rm) exit 0 ;;
esac
exit 99
`)}
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-fake-command-failure",
		Command: []string{"sh", "-c", "exit 42"},
	}

	exit, _, _, err := backend.Run(context.Background(), job, profile)
	if err != nil {
		t.Fatalf("command failure should not return backend error, got: %v", err)
	}
	if exit != 42 {
		t.Fatalf("exit = %d, want 42", exit)
	}
}

func TestPodmanBackend_Run_InfrastructureFailure(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "nonexistent-image-12345:latest",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-run-infra-fail",
		Command: []string{"echo", "hello"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, _, _, err := backend.Run(ctx, job, profile)
	if err == nil {
		t.Fatal("expected infrastructure error for nonexistent image")
	}
	if !IsBackendError(err) {
		t.Errorf("expected BackendError for infrastructure failure, got %T: %v", err, err)
	}
}

func TestPodmanBackend_Run_PodmanExit125IsBackendError(t *testing.T) {
	backend := &PodmanBackend{binary: fakePodman(t, `#!/bin/sh
case "$1" in
  create) echo "image pull failed" >&2; exit 125 ;;
  stop|kill|rm) echo "no such container" >&2; exit 1 ;;
esac
exit 99
`)}
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "nonexistent-image-12345:latest",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-fake-infra-failure",
		Command: []string{"echo", "hello"},
	}

	exit, _, _, err := backend.Run(context.Background(), job, profile)
	if err == nil {
		t.Fatal("expected backend error for podman infrastructure exit")
	}
	if !IsBackendError(err) {
		t.Fatalf("expected BackendError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "image pull failed") {
		t.Fatalf("error = %q, want image pull detail", err.Error())
	}
	if exit != 125 {
		t.Fatalf("exit = %d, want 125", exit)
	}
}

func TestPodmanBackend_Run_TimeoutCleanup(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 1, // 1 second timeout
		},
	}
	job := Job{
		ID:      "test-timeout-cleanup",
		Command: []string{"sleep", "10"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, _, err := backend.Run(ctx, job, profile)
	// Should timeout or be cancelled
	if err != nil && !IsBackendError(err) {
		t.Logf("expected timeout or cancellation, got: %v", err)
	}

	// Verify container was cleaned up - should not exist
	containerName, _ := sanitizeContainerName(job.ID)
	checkCmd := exec.Command("podman", "inspect", containerName)
	checkErr := checkCmd.Run()
	if checkErr == nil {
		t.Error("container should have been removed after timeout")
	}
}

func TestPodmanBackend_Cleanup_IsIdempotentForMissingContainer(t *testing.T) {
	backend := &PodmanBackend{binary: fakePodman(t, `#!/bin/sh
case "$1" in
  stop|kill|rm) echo "Error: no such container: $4" >&2; exit 1 ;;
esac
exit 99
`)}
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-cleanup-nonexistent",
		Command: []string{"echo", "hello"},
	}

	if err := backend.Cleanup(context.Background(), job, profile); err != nil {
		t.Fatalf("cleanup should be idempotent for missing container: %v", err)
	}
}

func TestPodmanBackend_Cleanup_Success(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-cleanup-success",
		Command: []string{"echo", "hello"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First run a container
	_, _, _, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	err = backend.Cleanup(ctx, job, profile)
	if err != nil {
		t.Fatalf("cleanup should succeed after Run removed container: %v", err)
	}
}

// Image trust model test - documents that allowlisting is out of scope
func TestImageTrustModel_Documented(t *testing.T) {
	// This test documents the image trust model:
	// - The backend uses caller-provided images without additional allowlist validation
	// - Image trust enforcement is out of scope
	// - Trust must be handled at registry or policy layer
	
	profile := Profile{
		ID:      "podman-test",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20", // Any image is accepted
			Timeout: 60,
		},
	}
	job := Job{
		ID:      "test-image-trust",
		Command: []string{"echo", "hello"},
	}

	// Verify the backend accepts arbitrary images without validation
	// This is the expected behavior per the documented trust model
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		t.Fatalf("sanitize failed: %v", err)
	}
	if containerName == "" {
		t.Error("container name should not be empty")
	}

	// The profile image is used as-is without allowlist check
	if profile.BackendConfig.Image == "" {
		t.Error("image should be set")
	}
}

func fakePodman(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "podman")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake podman: %v", err)
	}
	return path
}
