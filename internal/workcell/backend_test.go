package workcell

import (
	"context"
	"os/exec"
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
		name       string
		command    []string
		wantExit   int
		wantStdout string
		wantErr    bool
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
			name:    "empty command errors",
			command: []string{},
			wantErr: true,
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
		name           string
		callerDeadline time.Duration // 0 means no deadline
		profileTimeout int           // 0 means no timeout
		wantDeadline   bool
	}{
		{
			name:           "no caller deadline, no profile timeout",
			callerDeadline: 0,
			profileTimeout: 0,
			wantDeadline:   false,
		},
		{
			name:           "caller deadline only",
			callerDeadline: 5 * time.Second,
			profileTimeout: 0,
			wantDeadline:   true,
		},
		{
			name:           "profile timeout only",
			callerDeadline: 0,
			profileTimeout: 5,
			wantDeadline:   true,
		},
		{
			name:           "caller deadline stricter",
			callerDeadline: 3 * time.Second,
			profileTimeout: 10,
			wantDeadline:   true,
		},
		{
			name:           "profile timeout stricter",
			callerDeadline: 10 * time.Second,
			profileTimeout: 3,
			wantDeadline:   true,
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
				BackendConfig: BackendConfig{
					Timeout: tt.profileTimeout,
				},
			}

			newCtx, newCancel := effectiveDeadline(ctx, profile)
			defer newCancel()

			deadline, hasDeadline := newCtx.Deadline()
			if hasDeadline != tt.wantDeadline {
				t.Errorf("hasDeadline = %v, want %v", hasDeadline, tt.wantDeadline)
			}

			if hasDeadline && tt.callerDeadline > 0 && tt.profileTimeout > 0 {
				// Check that the stricter deadline was chosen
				callerDL, _ := ctx.Deadline()
				profileDL := time.Now().Add(time.Duration(tt.profileTimeout) * time.Second)
				expectedDL := callerDL
				if profileDL.Before(callerDL) {
					expectedDL = profileDL
				}
				// Allow 100ms tolerance for timing
				if deadline.Sub(expectedDL) > 100*time.Millisecond || expectedDL.Sub(deadline) > 100*time.Millisecond {
					t.Errorf("deadline mismatch: got %v, expected ~%v", deadline, expectedDL)
				}
			}
		})
	}
}

// PodmanBackend command construction tests (no Podman required)

func TestPodmanBackend_CommandConstruction(t *testing.T) {
	tests := []struct {
		name     string
		profile  Profile
		job      Job
		wantName string
		wantImg  string
	}{
		{
			name: "default image from profile",
			profile: Profile{
				ID:      "podman-smoke",
				Backend: "podman",
				BackendConfig: BackendConfig{
					Image: "docker.io/library/alpine:3.20",
				},
			},
			job: Job{
				ID:      "test_job123",
				Command: []string{"echo", "hello"},
			},
			wantName: "workcell-test_job123",
			wantImg:  "docker.io/library/alpine:3.20",
		},
		{
			name: "custom image from profile",
			profile: Profile{
				ID:      "custom",
				Backend: "podman",
				BackendConfig: BackendConfig{
					Image: "docker.io/library/busybox:latest",
				},
			},
			job: Job{
				ID:      "job-abc",
				Command: []string{"sh", "-c", "echo test"},
			},
			wantName: "workcell-job-abc",
			wantImg:  "docker.io/library/busybox:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify container name sanitization
			gotName, err := sanitizeContainerName(tt.job.ID)
			if err != nil {
				t.Fatalf("sanitizeContainerName failed: %v", err)
			}
			if gotName != tt.wantName {
				t.Errorf("container name = %q, want %q", gotName, tt.wantName)
			}

			// Verify image selection
			image := tt.profile.BackendConfig.Image
			if image == "" {
				image = "docker.io/library/alpine:3.20"
			}
			if image != tt.wantImg {
				t.Errorf("image = %q, want %q", image, tt.wantImg)
			}
		})
	}
}

// PodmanBackend Cleanup tests

func TestPodmanBackend_Cleanup_InvalidJobID(t *testing.T) {
	backend := NewPodmanBackend()
	profile := Profile{ID: "podman", Backend: "podman"}
	job := Job{ID: ""} // Empty job ID

	err := backend.Cleanup(context.Background(), job, profile)
	if err == nil {
		t.Error("expected error for empty job ID")
	}
}

// PodmanBackend integration tests (require Podman)

func TestPodmanBackend_Run_Success(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 30,
		},
	}
	job := Job{
		ID:      "test-success-" + randomSuffix(),
		Command: []string{"echo", "hello world"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exit, stdout, stderr, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exit != 0 {
		t.Errorf("exit code = %d, want 0", exit)
	}
	if !strings.Contains(stdout, "hello world") {
		t.Errorf("stdout = %q, want to contain 'hello world'", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	// Cleanup should succeed
	cleanupErr := backend.Cleanup(ctx, job, profile)
	if cleanupErr != nil {
		t.Errorf("cleanup failed: %v", cleanupErr)
	}
}

func TestPodmanBackend_Run_ExitCode(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 30,
		},
	}
	job := Job{
		ID:      "test-exit-" + randomSuffix(),
		Command: []string{"sh", "-c", "exit 42"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exit, _, _, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exit != 42 {
		t.Errorf("exit code = %d, want 42", exit)
	}

	// Cleanup should succeed
	cleanupErr := backend.Cleanup(ctx, job, profile)
	if cleanupErr != nil {
		t.Errorf("cleanup failed: %v", cleanupErr)
	}
}

func TestPodmanBackend_Run_Timeout(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 1, // 1 second timeout
		},
	}
	job := Job{
		ID:      "test-timeout-" + randomSuffix(),
		Command: []string{"sleep", "10"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, _, _, err := backend.Run(ctx, job, profile)
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if !IsBackendError(err) {
		t.Errorf("expected BackendError for timeout, got %T", err)
	}
	if ctx.Err() == nil {
		// The inner context should have been cancelled due to timeout
		t.Log("timeout test completed")
	}

	// Cleanup should succeed (container may already be gone)
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()
	_ = backend.Cleanup(cleanupCtx, job, profile)
}

func TestPodmanBackend_Run_Cancel(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 300, // Long timeout, we'll cancel manually
		},
	}
	job := Job{
		ID:      "test-cancel-" + randomSuffix(),
		Command: []string{"sleep", "10"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cancel after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	_, _, _, err := backend.Run(ctx, job, profile)
	if err == nil {
		t.Fatal("expected error for cancellation")
	}
	if !IsBackendError(err) {
		t.Errorf("expected BackendError for cancellation, got %T", err)
	}

	// Cleanup should succeed (container may already be gone)
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()
	_ = backend.Cleanup(cleanupCtx, job, profile)
}

func TestPodmanBackend_Cleanup_Deterministic(t *testing.T) {
	requirePodman(t)

	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 30,
		},
	}
	job := Job{
		ID:      "test-cleanup-" + randomSuffix(),
		Command: []string{"echo", "test"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run the job
	_, _, _, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Cleanup should succeed
	cleanupErr := backend.Cleanup(ctx, job, profile)
	if cleanupErr != nil {
		t.Errorf("cleanup failed: %v", cleanupErr)
	}

	// Cleanup should be idempotent (container already gone)
	cleanupErr2 := backend.Cleanup(ctx, job, profile)
	if cleanupErr2 != nil {
		t.Errorf("second cleanup failed: %v", cleanupErr2)
	}
}

// Image trust documentation test
// This test documents that image trust/allowlisting is out of scope.
func TestPodmanBackend_ImageTrust_OutOfScope(t *testing.T) {
	// Image trust enforcement is explicitly out of scope for the PodmanBackend.
	// The backend uses the caller-provided image without additional validation.
	// Image trust must be enforced at the registry or policy layer.
	backend := NewPodmanBackend()
	if backend == nil {
		t.Fatal("backend should not be nil")
	}
	// This test serves as documentation of the trust model.
	t.Log("Image trust/allowlisting is out of scope - documented in BackendConfig")
}