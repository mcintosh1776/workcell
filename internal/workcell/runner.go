package workcell

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	profiles          map[string]Profile
	backends          map[string]Backend
	jobs              map[string]Job
	logs              map[string]JobLogs
	validationResults map[string]ValidationWorkerResult
	mu                sync.RWMutex
}

func NewRunner(profiles map[string]Profile) *Runner {
	copied := make(map[string]Profile, len(profiles))
	for key, profile := range profiles {
		copied[key] = profile
	}

	// Initialize backends
	backends := make(map[string]Backend)
	backends["fake"] = &FakeBackend{}
	backends["podman"] = NewPodmanBackend()

	return &Runner{
		profiles:          copied,
		backends:          backends,
		jobs:              make(map[string]Job),
		logs:              make(map[string]JobLogs),
		validationResults: make(map[string]ValidationWorkerResult),
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
	exitCode, stdout, stderr, err := backend.Run(ctx, job, profile)
	logs := JobLogs{
		Stdout:    stdout,
		Stderr:    stderr,
		Truncated: strings.Contains(stdout, "[output truncated]") || strings.Contains(stderr, "[output truncated]"),
	}
	if err != nil {
		// Distinguish backend infrastructure errors from command failures
		job.State = JobFailed
		if exitCode != 0 {
			job.ExitCode = exitCode
		} else {
			job.ExitCode = -1
		}
		job.FinishedAt = time.Now().UTC()
		job.Logs.StdoutBytes = len(stdout)
		job.Logs.StderrBytes = len(stderr)
		job.Logs.Truncated = logs.Truncated
		job.Stdout = stdout
		job.Stderr = stderr
		// Preserve backend error details
		job.Error = err.Error()
		runner.mu.Lock()
		runner.jobs[job.ID] = job
		runner.logs[job.ID] = logs
		runner.mu.Unlock()
		return job, nil
	}

	job.FinishedAt = time.Now().UTC()
	job.ExitCode = exitCode
	job.Stdout = stdout
	job.Stderr = stderr
	if exitCode == 0 {
		job.State = JobSucceeded
	} else {
		job.State = JobFailed
	}
	job.Logs.StdoutBytes = len(stdout)
	job.Logs.StderrBytes = len(stderr)
	job.Logs.Truncated = logs.Truncated

	// Cleanup - do not silently treat failed cleanup as complete
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()
	cleanupErr := backend.Cleanup(cleanupCtx, job, profile)
	if cleanupErr != nil {
		job.Cleanup.State = "failed"
		job.Cleanup.Error = cleanupErr.Error()
	} else {
		job.Cleanup.State = "complete"
	}

	runner.mu.Lock()
	runner.jobs[job.ID] = job
	runner.logs[job.ID] = logs
	runner.mu.Unlock()

	return job, nil
}

func (runner *Runner) Get(id string) (Job, bool) {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	job, ok := runner.jobs[id]
	return cloneJob(job), ok
}

func (runner *Runner) List() []Job {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	jobs := make([]Job, 0, len(runner.jobs))
	for _, job := range runner.jobs {
		jobs = append(jobs, cloneJob(job))
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})
	return jobs
}

func (runner *Runner) Logs(id string) (JobLogs, bool) {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	logs, ok := runner.logs[id]
	return logs, ok
}

func cloneJob(job Job) Job {
	job.Command = append([]string(nil), job.Command...)
	return job
}

func randomSuffix() string {
	var bytes [6]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
