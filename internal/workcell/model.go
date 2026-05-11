package workcell

import "time"

type JobState string

const (
	JobQueued    JobState = "queued"
	JobRunning   JobState = "running"
	JobSucceeded JobState = "succeeded"
	JobFailed    JobState = "failed"
)

// SubmitJobRequest represents a request to submit a job.
type SubmitJobRequest struct {
	Profile        string         `json:"profile"`
	Command        []string       `json:"command"`
	Workspace      WorkspaceSpec  `json:"workspace,omitempty"`
	TimeoutSeconds int            `json:"timeoutSeconds,omitempty"`
	Env            EnvPolicy      `json:"env,omitempty"`
	Artifacts      ArtifactPolicy `json:"artifacts,omitempty"`
}

// WorkspaceSpec defines the workspace configuration.
type WorkspaceSpec struct {
	Type string `json:"type,omitempty"`
}

// EnvPolicy defines environment variable policies.
type EnvPolicy struct {
	Allow []string `json:"allow,omitempty"`
}

// ArtifactPolicy defines artifact collection policies.
type ArtifactPolicy struct {
	Paths []string `json:"paths,omitempty"`
}

// Job represents a job execution.
type Job struct {
	ID         string       `json:"id"`
	State      JobState     `json:"state"`
	Profile    string       `json:"profile"`
	Backend    string       `json:"backend"`
	Command    []string     `json:"command"`
	ExitCode   int          `json:"exitCode"`
	CreatedAt  time.Time    `json:"createdAt"`
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt time.Time    `json:"finishedAt"`
	Cleanup    CleanupState `json:"cleanup"`
	Logs       LogSummary   `json:"logs"`
	Stdout     string       `json:"stdout,omitempty"`
	Stderr     string       `json:"stderr,omitempty"`
	Artifacts  ArtifactInfo `json:"artifacts"`
	// Error contains backend infrastructure error details, if any.
	// This is distinct from ExitCode which indicates command failure.
	Error string `json:"error,omitempty"`
}

// CleanupState tracks the cleanup status of a job.
type CleanupState struct {
	State string `json:"state"`
	// Error contains cleanup failure details, if any.
	// Failed cleanup is not treated as complete.
	Error string `json:"error,omitempty"`
}

// LogSummary provides a summary of job logs.
type LogSummary struct {
	StdoutBytes int `json:"stdoutBytes"`
	StderrBytes int `json:"stderrBytes"`
}

// ArtifactInfo provides information about collected artifacts.
type ArtifactInfo struct {
	Count int `json:"count"`
	Bytes int `json:"bytes"`
}

// Profile defines a job execution profile.
type Profile struct {
	ID            string
	Backend       string
	BackendConfig BackendConfig
}

// BackendConfig contains backend-specific configuration.
// Image trust: Image allowlisting is out of scope for this implementation.
// Image trust enforcement must be handled at the registry or policy layer.
type BackendConfig struct {
	Image   string
	Timeout int // seconds, 0 means no additional timeout beyond context
}

// DefaultProfiles returns the default set of profiles.
func DefaultProfiles() map[string]Profile {
	return map[string]Profile{
		"fake": {
			ID:      "fake",
			Backend: "fake",
		},
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
			BackendConfig: BackendConfig{
				Image:   "docker.io/library/alpine:3.20",
				Timeout: 300,
			},
		},
	}
}
