package workcell

import (
	"testing"
)

func TestIsTerminalJobState(t *testing.T) {
	tests := []struct {
		name     string
		state    JobState
		expected bool
	}{
		{
			name:     "succeeded state is terminal",
			state:    JobSucceeded,
			expected: true,
		},
		{
			name:     "failed state is terminal",
			state:    JobFailed,
			expected: true,
		},
		{
			name:     "queued state is not terminal",
			state:    JobQueued,
			expected: false,
		},
		{
			name:     "running state is not terminal",
			state:    JobRunning,
			expected: false,
		},
		{
			name:     "unknown state is not terminal",
			state:    JobState("unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTerminalJobState(tt.state)
			if result != tt.expected {
				t.Errorf("IsTerminalJobState(%v) = %v, want %v", tt.state, result, tt.expected)
			}
		})
	}
}