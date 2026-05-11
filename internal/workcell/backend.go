package workcell

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
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
	switch job.Command[0] {
	case "echo":
		return 0, strings.Join(job.Command[1:], " ") + "\n", "", nil
	case "false":
		return 1, "", "", nil
	default:
		return 1, "", "", &BackendError{
			Op:  "validate",
			Err: fmt.Errorf("fake backend only supports deterministic echo and false commands"),
		}
	}
}

func (b *FakeBackend) Cleanup(ctx context.Context, job Job, profile Profile) error {
	return nil
}

// PodmanBackend runs commands in Podman containers.
// Image trust: This backend uses the caller-provided image without additional
// allowlist validation. Image trust enforcement is out of scope for this
// implementation and must be handled at the registry or policy layer.
type PodmanBackend struct {
	binary          string
	maxCaptureBytes int
}

var podmanBinary = "podman"
var errOutputLimitExceeded = errors.New("podman output limit exceeded")

func NewPodmanBackend() *PodmanBackend {
	return &PodmanBackend{binary: podmanBinary, maxCaptureBytes: 10 * 1024 * 1024}
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

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(jobID)))[:12]
	name := prefix + string(sanitized) + "-" + hash
	if len(name) > 128 {
		suffix := "-" + hash
		maxSanitized := 128 - len(prefix) - len(suffix)
		name = prefix + string(sanitized[:maxSanitized]) + suffix
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
		return 0, "", "", &BackendError{Op: "validate", Err: fmt.Errorf("no command")}
	}
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

	// Build podman create command. We use create/start/inspect instead of
	// podman run so command exit codes can be distinguished from Podman
	// infrastructure failures.
	args := []string{
		"create",
		"--name", containerName,
		"--pull", "missing",
		"--network", "none",
		"--security-opt", "no-new-privileges",
		"--user", "65532:65532",
		"--pids-limit", "256",
		"--memory", "512m",
		"--cpus", "1",
		"--read-only",
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=64m",
		"--tmpfs", "/var/tmp:rw,noexec,nosuid,size=64m",
	}
	if profile.BackendConfig.Timeout > 0 {
		// Add a podman timeout as well for extra safety
		args = append(args, "--stop-timeout", fmt.Sprintf("%d", profile.BackendConfig.Timeout))
	}
	args = append(args, profile.BackendConfig.Image)
	args = append(args, job.Command...)

	createStdout, createStderr, err := b.runPodmanCapture(ctx, args...)
	if err != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = b.forceRemoveContainer(cleanupCtx, containerName)
		return podmanProcessExitCode(err), createStdout, createStderr, &BackendError{
			Op:  "create",
			Err: errorWithOutput(err, createStderr),
		}
	}

	// Capture stdout and stderr
	stdout, stderr, startErr := b.runPodmanCapture(ctx, "start", "--attach", containerName)
	exitCode, inspected := b.inspectExitCode(ctx, containerName)

	// Deterministic cleanup: always try to remove container after run
	// This handles cases where the context was cancelled or process was killed
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()

	// Force remove the container to ensure no orphans
	_ = b.forceRemoveContainer(cleanupCtx, containerName)

	if startErr != nil {
		if errors.Is(startErr, errOutputLimitExceeded) {
			return exitCode, stdout, stderr, &BackendError{Op: "output", Err: errorWithOutput(startErr, stderr)}
		}
		if ctx.Err() != nil {
			return podmanProcessExitCode(startErr), stdout, stderr, &BackendError{Op: "start", Err: errorWithOutput(ctx.Err(), stderr)}
		}
		if inspected {
			return exitCode, stdout, stderr, nil
		}
		return podmanProcessExitCode(startErr), stdout, stderr, &BackendError{Op: "start", Err: errorWithOutput(startErr, stderr)}
	}

	if inspected {
		return exitCode, stdout, stderr, nil
	}
	return 0, stdout, stderr, nil
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
	_, stderr, err := b.runPodmanCapture(ctx, args...)
	if err != nil {
		detail := strings.TrimSpace(stderr)
		if detail == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, detail)
	}
	return nil
}

func (b *PodmanBackend) runPodmanCapture(ctx context.Context, args ...string) (string, string, error) {
	cmd := b.command(ctx, args...)
	captureLimit := b.maxCaptureBytes
	if captureLimit <= 0 {
		captureLimit = 10 * 1024 * 1024
	}
	stdout := &cappedBuffer{limit: captureLimit}
	stderr := &cappedBuffer{limit: captureLimit}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if stdout.truncated || stderr.truncated {
		if err == nil {
			err = errOutputLimitExceeded
		} else {
			err = fmt.Errorf("%w: %v", errOutputLimitExceeded, err)
		}
	}
	return stdout.String(), stderr.String(), err
}

type cappedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		b.truncated = true
		return len(p), nil
	}
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buffer.Write(p)
	return len(p), nil
}

func (b *cappedBuffer) String() string {
	if b.truncated {
		return b.buffer.String() + "\n[output truncated]"
	}
	return b.buffer.String()
}

func (b *PodmanBackend) inspectExitCode(ctx context.Context, containerName string) (int, bool) {
	stdout, _, err := b.runPodmanCapture(ctx, "inspect", "--format", "{{.State.ExitCode}}", containerName)
	if err != nil {
		return -1, false
	}
	exitCode, parseErr := strconv.Atoi(strings.TrimSpace(stdout))
	if parseErr != nil {
		return -1, false
	}
	return exitCode, true
}

func podmanProcessExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func errorWithOutput(err error, stderr string) error {
	detail := strings.TrimSpace(stderr)
	if detail == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, detail)
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
