package workcell

import (
  "testing"
)

func TestHasBackendError(t *testing.T) {
  tests := []struct {
    name string
    job Job
    want bool
  }{
    {name: "failed with backend error", job: Job{State: JobFailed, Error: "podman unavailable"}, want: true},
    {name: "failed command without backend error", job: Job{State: JobFailed, ExitCode: 1}, want: false},
    {name: "successful job", job: Job{State: JobSucceeded}, want: false},
    {name: "whitespace only error", job: Job{State: JobFailed, Error: "  "}, want: false},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := HasBackendError(tt.job); got != tt.want {
        t.Fatalf("HasBackendError(%+v) = %v, want %v", tt.job, got, tt.want)
      }
    })
  }
}