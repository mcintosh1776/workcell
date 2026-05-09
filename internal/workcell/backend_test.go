package workcell

import (
	"context"
	"testing"
)

func TestPodmanBackendRunCommand(t *testing.T) {
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
	if stdout != "hello
" {
		t.Fatalf("Stdout = %q, want "hello\n"", stdout)
	}
	if stderr != "" {
		t.Fatalf("Stderr = %q, want empty", stderr)
	}
}

func TestPodmanBackendRunFalseCommand(t *testing.T) {
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")
	
	ctx := context.Background()
	job := Job{
		ID:      "test-job-456",
		Command: []string{"false"},
	}
	
	exitCode, stdout, stderr, err := backend.Run(ctx, job)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", exitCode)
	}
	// Cleanup should succeed even for failed commands
	err = backend.Cleanup(ctx, job)
	if err != nil {
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
	
	// Cleanup should succeed even if container was already removed by --rm
	err := backend.Cleanup(ctx, job)
	if err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}