package workcell

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	profiles map[string]Profile
	backends map[string]Backend
	jobs     map[string]Job
	mu       sync.RWMutex
}

func NewRunner(profiles map[string]Profile) *Runner {
	copied := make(map[string]Profile, len(profiles))
	for key, profile := range profiles {
		copied[key] = profile
	}

	// Initialize backends
	backends := make(map[string]Backend)
	backends["fake"] = &FakeBackend{}
	backends["podman"] = NewPodmanBackend("")

	return &Runner{
		profiles: copied,
		backends: backends,
		jobs:     make(map[string]Job),
	}
}

func (runner *Runner) Run(ctx context.Context, request SubmitJobRequest) (Job, error) {
	profileID := strings.TrimSpace(request.Profile)
	if profileID == "" {
		profileID = "fake"
	}
	profile, ok := runner.profiles[profileID]
	if !ok {
		return Job{}, fmt.Errorf("%w: %s", ErrInvalidProfile, profileID)
	}
	if len(request.Command) == 0 {
		return Job{}, fmt.Errorf("%w: command array is required", ErrInvalidCommand)
	}

	now := time.Now().UTC()
	job := Job{
		ID:        "job_" + randomSuffix(),
		State:     JobRunning,
		Profile:   profile.ID,
		Backend:   profile.Backend,
		Command:   append([]string(nil), request.Command...),
		CreatedAt: now,
		StartedAt: now,
		Cleanup: CleanupState{
			State: "pending",
		},
	}

	select {
	case <-ctx.Done():
		return Job{}, ctx.Err()
	default:
	}

	// Get the backend for this profile
	backend, ok := runner.backends[profile.Backend]
	if !ok {
		return Job{}, fmt.Errorf("%w: backend %s not found", ErrInvalidProfile, profile.Backend)
	}

	// Run the job
	exitCode, stdout, stderr, err := backend.Run(ctx, job)
	if err != nil {
		job.State = JobFailed
		job.ExitCode = 1
		job.FinishedAt = time.Now().UTC()
		job.Logs.StdoutBytes = len(stdout)
		job.Logs.StderrBytes = len(stderr)
		runner.mu.Lock()
		runner.jobs[job.ID] = job
		runner.mu.Unlock()
		return job, nil
	}

	job.FinishedAt = time.Now().UTC()
	job.ExitCode = exitCode
	if exitCode == 0 {
		job.State = JobSucceeded
	} else {
		job.State = JobFailed
	}
	job.Logs.StdoutBytes = len(stdout)
	job.Logs.StderrBytes = len(stderr)

	// Cleanup
	if err := backend.Cleanup(ctx, job); err != nil {
		job.Cleanup.State = "failed"
	} else {
		job.Cleanup.State = "complete"
	}

	runner.mu.Lock()
	runner.jobs[job.ID] = job
	runner.mu.Unlock()

	return job, nil
}

func (runner *Runner) Get(id string) (Job, bool) {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	job, ok := runner.jobs[id]
	return job, ok
}

func randomSuffix() string {
	var bytes [6]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}