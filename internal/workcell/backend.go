package workcell

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
type PodmanBackend struct {
	binary string
}

var podmanBinary = "podman"

func NewPodmanBackend() *PodmanBackend {
	return &PodmanBackend{binary: podmanBinary}
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
	if strings.TrimSpace(profile.BackendConfig.Image) == "" {
		return 0, "", "", &BackendError{Op: "validate", Err: fmt.Errorf("podman image is required")}
	}

	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return 0, "", "", &BackendError{Op: "sanitize", Err: err}
	}

	// Apply effective deadline combining caller context and profile timeout
	ctx, cancel := effectiveDeadline(ctx, profile)
	defer cancel()

	// Build podman run command
	args := []string{"run", "--rm", "--name", containerName}
	if profile.BackendConfig.Timeout > 0 {
		// Add a podman timeout as well for extra safety
		args = append(args, "--stop-timeout", fmt.Sprintf("%d", profile.BackendConfig.Timeout))
	}
	args = append(args, profile.BackendConfig.Image)
	args = append(args, job.Command...)

	cmd := b.command(ctx, args...)

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	// Deterministic cleanup: always try to remove container after run
	// This handles cases where the context was cancelled or process was killed
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()

	// Force remove the container to ensure no orphans
	_ = b.forceRemoveContainer(cleanupCtx, containerName)

	if err != nil {
		// Check if it's an exit error (command failed) or infrastructure error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if ctx.Err() != nil || isPodmanInfrastructureExitCode(exitErr.ExitCode()) {
				return exitErr.ExitCode(), stdoutBuf.String(), stderrBuf.String(), &BackendError{Op: "run", Err: err}
			}
			// Command ran but exited non-zero - this is not a backend error.
			return exitErr.ExitCode(), stdoutBuf.String(), stderrBuf.String(), nil
		}
		// Infrastructure error (e.g., podman not found, image pull failed)
		return 0, stdoutBuf.String(), stderrBuf.String(), &BackendError{Op: "run", Err: err}
	}

	return 0, stdoutBuf.String(), stderrBuf.String(), nil
}

// forceRemoveContainer forcefully removes a container, ignoring most errors.
// Used for deterministic cleanup after run.
func (b *PodmanBackend) forceRemoveContainer(ctx context.Context, containerName string) error {
	// Try to stop first (ignore errors - container might not exist or already stopped)
	_ = b.containerStop(ctx, containerName)
	// Then remove (ignore errors - container might not exist)
	_ = b.containerRm(ctx, containerName)
	return nil
}

// Cleanup performs thorough cleanup and returns error if any step fails.
// Callers should not treat failed cleanup as complete.
func (b *PodmanBackend) Cleanup(ctx context.Context, job Job, profile Profile) error {
	containerName, err := sanitizeContainerName(job.ID)
	if err != nil {
		return fmt.Errorf("cleanup sanitize failed: %w", err)
	}

	var cleanupErrors []error

	// Try graceful stop first
	if err := b.containerStop(ctx, containerName); err != nil {
		if isMissingContainerError(err) {
			return nil
		}
		// If stop fails, try kill
		if killErr := b.containerKill(ctx, containerName); killErr != nil {
			if !isMissingContainerError(killErr) {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("stop failed (%v) and kill failed (%v)", err, killErr))
			}
		}
	}

	// Remove the container
	if err := b.containerRm(ctx, containerName); err != nil {
		if !isMissingContainerError(err) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("rm failed: %w", err))
		}
	}

	if len(cleanupErrors) > 0 {
		return fmt.Errorf("cleanup incomplete: %v", cleanupErrors)
	}
	return nil
}

func (b *PodmanBackend) command(ctx context.Context, args ...string) *exec.Cmd {
	binary := b.binary
	if strings.TrimSpace(binary) == "" {
		binary = podmanBinary
	}
	return exec.CommandContext(ctx, binary, args...)
}

func (b *PodmanBackend) containerStop(ctx context.Context, containerName string) error {
	return b.runPodman(ctx, "stop", "-t", "10", containerName)
}

func (b *PodmanBackend) containerKill(ctx context.Context, containerName string) error {
	return b.runPodman(ctx, "kill", containerName)
}

func (b *PodmanBackend) containerRm(ctx context.Context, containerName string) error {
	return b.runPodman(ctx, "rm", "-f", containerName)
}

func (b *PodmanBackend) runPodman(ctx context.Context, args ...string) error {
	cmd := b.command(ctx, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, detail)
	}
	return nil
}

func isPodmanInfrastructureExitCode(exitCode int) bool {
	return exitCode == 125 || exitCode == 126 || exitCode == 127
}

func isMissingContainerError(err error) bool {
	if err == nil {
		return false
	}
	return isMissingContainerMessage(err.Error())
}

func isMissingContainerMessage(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	return strings.Contains(normalized, "no such container") ||
		strings.Contains(normalized, "container does not exist") ||
		strings.Contains(normalized, "no container with name")
}
