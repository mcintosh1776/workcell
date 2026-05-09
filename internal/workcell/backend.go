package workcell

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Backend executes jobs.
type Backend interface {
	// Run executes the job and returns the exit code, stdout, stderr, and any error.
	Run(ctx context.Context, job Job, profile Profile) (exitCode int, stdout, stderr string, err error)
	// Cleanup removes any resources created by the backend for the job.
	Cleanup(ctx context.Context, job Job, profile Profile) error
}

// FakeBackend is a fake backend for testing.
type FakeBackend struct{}

func (b *FakeBackend) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", fmt.Errorf("no command")
	}
	stdout := strings.Join(job.Command, " ")
	if job.Command[0] == "false" {
		return 1, stdout, "", nil
	}
	return 0, stdout, "", nil
}

func (b *FakeBackend) Cleanup(ctx context.Context, job Job, profile Profile) error {
	return nil
}

// PodmanBackend runs commands in Podman containers.
type PodmanBackend struct{}

func NewPodmanBackend() *PodmanBackend {
	return &PodmanBackend{}
}

// sanitizeContainerName validates and sanitizes a job ID for use as a container name.
// Container names must match [a-zA-Z0-9][a-zA-Z0-9_.-]* and be 1-128 characters.
func sanitizeContainerName(jobID string) (string, error) {
	if jobID == "" {
		return "", fmt.Errorf("job ID cannot be empty")
	}
	
	// Prefix with workcell- to namespace our containers
	prefix := "workcell-"
	
	// Sanitize the job ID: only allow alphanumeric, underscore, dot, hyphen
	sanitized := make([]byte, 0, len(jobID))
	for i, c := range jobID {
		if i == 0 && !isValidFirstChar(c) {
			// If first char is not valid, prefix with 'j'
			sanitized = append(sanitized, 'j')
		}
		if isValidContainerChar(c) {
			sanitized = append(sanitized, byte(c))
		} else {
			sanitized = append(sanitized, '-')
		}
	}
	
	name := prefix + string(sanitized)
	if len(name) > 128 {
		name = name[:128]
	}
	
	return name, nil
}

func isValidFirstChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isValidContainerChar(c rune) bool {
	return isValidFirstChar(c) || c == '_' || c == '.' || c == '-'
}

func (b *PodmanBackend) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", fmt.Errorf("no command")
	}

	// Get image from profile config, fallback to default
	image := profile.BackendConfig.Image
	if image == "" {
		image = "docker.io/library/alpine:3.20"
	}

	// Sanitize container name
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return 1, "", "", fmt.Errorf("invalid job ID for container name: %w", err)
	}

	// Apply timeout from profile if set and context doesn't already have a shorter deadline
	if profile.BackendConfig.Timeout > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(profile.BackendConfig.Timeout)*time.Second)
			defer cancel()
		}
	}

	// Build podman run command
	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		image,
	}
	args = append(args, job.Command...)

	cmd := exec.CommandContext(ctx, "podman", args...)
	
	// Capture stdout and stderr separately
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 1, "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 1, "", "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return 1, "", "", fmt.Errorf("failed to start podman: %w", err)
	}

	// Read stdout and stderr concurrently to avoid blocking
	var stdoutBuf, stderrBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(&stdoutBuf, stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		io.Copy(&stderrBuf, stderrPipe)
	}()

	// Wait for command completion or context cancellation
	done := make(chan error, 1)
	go func() {
		wg.Wait()
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context cancelled or timed out - kill the process
		cmd.Process.Kill()
		<-done // Wait for process to exit
		return 1, stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("context cancelled: %w", ctx.Err())
	case err := <-done:
		stdout := stdoutBuf.String()
		stderr := stderrBuf.String()
		
		var exitCode int
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				return 1, stdout, stderr, fmt.Errorf("podman run failed: %w", err)
			}
		}
		
		return exitCode, stdout, stderr, nil
	}
}

func (b *PodmanBackend) Cleanup(ctx context.Context, job Job, profile Profile) error {
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return fmt.Errorf("invalid job ID for container name: %w", err)
	}

	// Best-effort removal: try to stop and remove the container explicitly
	// Ignore errors since container may already be gone (via --rm) or never created
	
	// Try to stop first (graceful, then force)
	stopCtx, stopCancel := context.WithTimeout(ctx, 10*time.Second)
	defer stopCancel()
	exec.CommandContext(stopCtx, "podman", "stop", "-t", "2", containerName).Run()
	
	// Force kill if still running
	killCtx, killCancel := context.WithTimeout(ctx, 5*time.Second)
	defer killCancel()
	exec.CommandContext(killCtx, "podman", "kill", containerName).Run()
	
	// Remove the container
	rmCtx, rmCancel := context.WithTimeout(ctx, 10*time.Second)
	defer rmCancel()
	exec.CommandContext(rmCtx, "podman", "rm", containerName).Run()
	
	return nil
}