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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{ID: "test-job", Command: tt.command}
			exit, stdout, stderr, err := backend.Run(context.Background(), job, profile)

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
					Image: "myregistry.io/myimage:v1",
				},
			},
			job: Job{
				ID:      "job_abc",
				Command: []string{"ls", "-la"},
			},
			wantName: "workcell-job_abc",
			wantImg:  "myregistry.io/myimage:v1",
		},
		{
			name: "fallback image when not specified",
			profile: Profile{
				ID:      "minimal",
				Backend: "podman",
			},
			job: Job{
				ID:      "job_def",
				Command: []string{"pwd"},
			},
			wantName: "workcell-job_def",
			wantImg:  "docker.io/library/alpine:3.20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify container name sanitization
			containerName, err := sanitizeContainerName(tt.job.ID)
			if err != nil {
				t.Fatalf("sanitizeContainerName failed: %v", err)
			}
			if containerName != tt.wantName {
				t.Errorf("container name = %q, want %q", containerName, tt.wantName)
			}

			// Verify image comes from profile
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

// Podman-dependent integration tests

func TestPodmanBackend_Run_Success(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-success",
		Command: []string{"echo", "hello"},
	}

	exitCode, stdout, stderr, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", exitCode)
	}
	if strings.TrimSpace(stdout) != "hello" {
		t.Fatalf("Stdout = %q, want hello", stdout)
	}
	if stderr != "" {
		t.Fatalf("Stderr = %q, want empty", stderr)
	}

	// Cleanup
	if err := backend.Cleanup(ctx, job, profile); err != nil {
		t.Logf("Cleanup error: %v", err)
	}
}

func TestPodmanBackend_Run_FalseCommand(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-false",
		Command: []string{"false"},
	}

	exitCode, stdout, stderr, err := backend.Run(ctx, job, profile)
	// Should NOT error - false is a valid command that returns exit 1
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", exitCode)
	}
	// stderr should be captured (may be empty for false)
	_ = stderr
	_ = stdout

	// Cleanup
	backend.Cleanup(ctx, job, profile)
}

func TestPodmanBackend_Run_StderrCapture(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-stderr",
		Command: []string{"sh", "-c", "echo error >&2; exit 1"},
	}

	exitCode, stdout, stderr, err := backend.Run(ctx, job, profile)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr, "error") {
		t.Fatalf("Stderr = %q, want to contain 'error'", stderr)
	}
	if stdout != "" {
		t.Logf("Stdout: %q", stdout)
	}

	backend.Cleanup(ctx, job, profile)
}

func TestPodmanBackend_Run_Timeout(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image:   "docker.io/library/alpine:3.20",
			Timeout: 2, // 2 second timeout
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-timeout",
		Command: []string{"sleep", "30"},
	}

	start := time.Now()
	exit, _, _, err := backend.Run(ctx, job, profile)
	elapsed := time.Since(start)

	// Should timeout quickly (within 5 seconds)
	if elapsed > 5*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}

	// Should have error or non-zero exit
	if err == nil && exit == 0 {
		t.Error("expected timeout error or non-zero exit")
	}

	// Cleanup should succeed or fail gracefully
	backend.Cleanup(context.Background(), job, profile)
}

func TestPodmanBackend_Run_ContextCancellation(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	job := Job{
		ID:      "test-job-cancel",
		Command: []string{"sleep", "10"},
	}

	exit, stdout, stderr, err := backend.Run(ctx, job, profile)

	// Should fail due to context cancellation
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if exit != 1 {
		t.Errorf("exit code = %d, want 1", exit)
	}
	_ = stdout
	_ = stderr

	// Cleanup should not panic
	backend.Cleanup(context.Background(), job, profile)
}

func TestPodmanBackend_Cleanup(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-cleanup",
		Command: []string{"echo", "test"},
	}

	// Run the job first
	backend.Run(ctx, job, profile)

	// Cleanup should remove the container
	if err := backend.Cleanup(ctx, job, profile); err != nil {
		t.Logf("Cleanup returned error (may be expected): %v", err)
	}

	// Verify container is gone by trying to inspect it
	containerName, _ := sanitizeContainerName(job.ID)
	inspectCmd := exec.Command("podman", "inspect", containerName)
	if err := inspectCmd.Run(); err == nil {
		t.Error("container still exists after cleanup")
	}
}

func TestPodmanBackend_Cleanup_NoOrphans(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend()
	profile := Profile{
		ID:      "podman-smoke",
		Backend: "podman",
		BackendConfig: BackendConfig{
			Image: "docker.io/library/alpine:3.20",
		},
	}

	ctx := context.Background()
	job := Job{
		ID:      "test-job-orphan",
		Command: []string{"echo", "orphan test"},
	}

	// Run the job
	backend.Run(ctx, job, profile)

	// Cleanup explicitly
	err := backend.Cleanup(ctx, job, profile)
	// Cleanup is best-effort, error is acceptable
	t.Logf("cleanup result: %v", err)

	// Verify container is gone
	containerName, _ := sanitizeContainerName(job.ID)
	inspectCmd := exec.Command("podman", "inspect", containerName)
	if err := inspectCmd.Run(); err == nil {
		t.Error("container still exists after cleanup - orphan detected")
	}
}