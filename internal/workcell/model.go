package workcell

import "time"

type JobState string

const (
	JobQueued    JobState = "queued"
	JobRunning   JobState = "running"
	JobSucceeded JobState = "succeeded"
	JobFailed    JobState = "failed"
)

type SubmitJobRequest struct {
	Profile        string         `json:"profile"`
	Command        []string       `json:"command"`
	Workspace      WorkspaceSpec  `json:"workspace,omitempty"`
	TimeoutSeconds int            `json:"timeoutSeconds,omitempty"`
	Env            EnvPolicy      `json:"env,omitempty"`
	Artifacts      ArtifactPolicy `json:"artifacts,omitempty"`
}

type WorkspaceSpec struct {
	Type string `json:"type,omitempty"`
}

type EnvPolicy struct {
	Allow []string `json:"allow,omitempty"`
}

type ArtifactPolicy struct {
	Paths []string `json:"paths,omitempty"`
}

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
	Artifacts  ArtifactInfo `json:"artifacts"`
}

type CleanupState struct {
	State string `json:"state"`
}

type LogSummary struct {
	StdoutBytes int `json:"stdoutBytes"`
	StderrBytes int `json:"stderrBytes"`
}

type ArtifactInfo struct {
	Count int `json:"count"`
	Bytes int `json:"bytes"`
}

type Profile struct {
	ID      string
	Backend string
}

func DefaultProfiles() map[string]Profile {
	return map[string]Profile{
		"fake": {
			ID:      "fake",
			Backend: "fake",
		},
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
		},
	}
}
