package workcell

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Backend executes jobs.
type Backend interface {
	// Run executes the job and returns the exit code, stdout, stderr, and any error.
	Run(ctx context.Context, job Job) (exitCode int, stdout, stderr string, err error)
	// Cleanup removes any resources created by the backend for the job.
	Cleanup(ctx context.Context, job Job) error
}

// FakeBackend is a fake backend for testing.
type FakeBackend struct{}

func (b *FakeBackend) Run(ctx context.Context, job Job) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", fmt.Errorf("no command")
	}
	stdout := strings.Join(job.Command, " ")
	if job.Command[0] == "false" {
		return 1, stdout, "", nil
	}
	return 0, stdout, "", nil
}

func (b *FakeBackend) Cleanup(ctx context.Context, job Job) error {
	return nil
}

// PodmanBackend runs commands in Podman containers.
type PodmanBackend struct {
	image string
}

func NewPodmanBackend(image string) *PodmanBackend {
	if image == "" {
		image = "docker.io/library/alpine:3.20"
	}
	return &PodmanBackend{image: image}
}

func (b *PodmanBackend) Run(ctx context.Context, job Job) (int, string, string, error) {
	if len(job.Command) == 0 {
		return 1, "", "", fmt.Errorf("no command")
	}

	// Create container name based on job ID
	containerName := "workcell-" + job.ID

	// Build podman run command
	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		b.image,
	}
	args = append(args, job.Command...)

	cmd := exec.CommandContext(ctx, "podman", args...)
	stdoutBytes, err := cmd.Output()

	var stderr string
	var exitCode int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		} else {
			return 1, "", "", fmt.Errorf("podman run failed: %w", err)
		}
	}

	return exitCode, string(stdoutBytes), stderr, nil
}

func (b *PodmanBackend) Cleanup(ctx context.Context, job Job) error {
	// Container is removed automatically via --rm flag
	// This is a no-op but kept for interface compliance
	return nil
}