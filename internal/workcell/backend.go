package workcell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// BackendError represents an infrastructure error from the backend itself,
// distinct from a command that ran but exited with a non-zero code.
type BackendError struct {
	Op  string
	Err error
}

func (e *BackendError) Error() string {
	return fmt.Sprintf("backend %s failed: %v", e.Op, e.Err)
}

func (e *BackendError) Unwrap() error {
	return e.Err
}

// IsBackendError reports whether err is a BackendError.
func IsBackendError(err error) bool {
	var be *BackendError
	return errors.As(err, &be)
}

// Backend executes jobs.
type Backend interface {
	// Run executes the job and returns the exit code, stdout, stderr, and any error.
	// A non-nil error indicates a backend infrastructure failure (e.g., cannot start container).
	// The caller should check IsBackendError(err) to distinguish from command failures.
	Run(ctx context.Context, job Job, profile Profile) (exitCode int, stdout, stderr string, err error)
	// Cleanup removes any resources created by the backend for the job.
	// Returns error if cleanup fails; callers should not treat failed cleanup as complete.
	Cleanup(ctx context.Context, job Job, profile Profile) error
}

// FakeBackend is a fake backend for testing.
type FakeBackend struct{}

func (b *FakeBackend) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", &BackendError{Op: "validate", Err: fmt.Errorf("no command")}
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
// Image trust: This backend uses the caller-provided image without additional
// allowlist validation. Image trust enforcement is out of scope for this
// implementation and must be handled at the registry or policy layer.
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

// effectiveDeadline returns the stricter deadline between the caller's context
// and the profile timeout. If both have deadlines, the earlier one wins.
func effectiveDeadline(ctx context.Context, profile Profile) (context.Context, context.CancelFunc) {
	if profile.BackendConfig.Timeout <= 0 {
		return ctx, func() {}
	}

	profileTimeout := time.Duration(profile.BackendConfig.Timeout) * time.Second
	callerDeadline, callerHasDeadline := ctx.Deadline()
	profileDeadline := time.Now().Add(profileTimeout)

	if !callerHasDeadline || profileDeadline.Before(callerDeadline) {
		return context.WithTimeout(ctx, profileTimeout)
	}
	return ctx, func() {}
}

func (b *PodmanBackend) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", &BackendError{Op: "validate", Err: fmt.Errorf("no command")}
	}

	// Get image from profile config, fallback to default
	image := profile.BackendConfig.Image
	if image == "" {
		image = "docker.io/library/alpine:3.20"
	}

	// Sanitize container name
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return 1, "", "", &BackendError{Op: "sanitize", Err: fmt.Errorf("invalid job ID for container name: %w", err)}
	}

	// Apply the stricter effective deadline between caller context and profile timeout
	ctx, cancel := effectiveDeadline(ctx, profile)
	defer cancel()

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
		return 1, "", "", &BackendError{Op: "stdout pipe", Err: err}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 1, "", "", &BackendError{Op: "stderr pipe", Err: err}
	}

	if err := cmd.Start(); err != nil {
		return 1, "", "", &BackendError{Op: "start", Err: err}
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
	waitErr := cmd.Wait()
	wg.Wait()

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if waitErr != nil {
		// Check if this was a context cancellation/timeout
		if ctx.Err() != nil {
			// Ensure cleanup runs deterministically even on timeout/cancel
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cleanupCancel()
			_ = b.Cleanup(cleanupCtx, job, profile)
			return 1, stdout, stderr, &BackendError{Op: "run", Err: fmt.Errorf("context cancelled: %w", ctx.Err())}
		}

		// Check for exit code
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			return exitErr.ExitCode(), stdout, stderr, nil
		}
		return 1, stdout, stderr, &BackendError{Op: "run", Err: waitErr}
	}

	return 0, stdout, stderr, nil
}

// Cleanup removes the container created for the job.
// This method runs podman rm -f to ensure deterministic cleanup.
// Returns an error if cleanup fails; callers should not treat failed cleanup as complete.
func (b *PodmanBackend) Cleanup(ctx context.Context, job Job, profile Profile) error {
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return fmt.Errorf("cleanup: invalid job ID: %w", err)
	}

	// Use -f to force removal and -v to remove volumes
	cmd := exec.CommandContext(ctx, "podman", "rm", "-f", "-v", containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Check if container doesn't exist (already cleaned up or never created)
		if strings.Contains(string(out), "no such container") ||
			strings.Contains(string(out), "does not exist") {
			return nil
		}
		return fmt.Errorf("cleanup failed: %w (output: %s)", err, string(out))
	}
	return nil
}