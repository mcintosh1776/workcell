package workcell

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func requirePodman(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skipf("podman is not available in this test environment: %v", err)
	}
}

func TestPodmanBackendRunCommand(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")

	ctx := context.Background()
	job := Job{
		ID:      "test-job-123",
		Command: []string{"echo", "hello"},
	}

	exitCode, stdout, stderr, err := backend.Run(ctx, job)
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
}

func TestPodmanBackendRunFalseCommand(t *testing.T) {
	requirePodman(t)
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")

	ctx := context.Background()
	job := Job{
		ID:      "test-job-456",
		Command: []string{"false"},
	}

	exitCode, _, _, err := backend.Run(ctx, job)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", exitCode)
	}
	if err := backend.Cleanup(ctx, job); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}

func TestPodmanBackendCleanup(t *testing.T) {
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")

	ctx := context.Background()
	job := Job{
		ID:      "test-job-789",
		Command: []string{"echo", "test"},
	}

	if err := backend.Cleanup(ctx, job); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}
