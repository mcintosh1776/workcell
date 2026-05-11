package workcell

import (
	"context"
	"testing"
)

// TestQuickStartExamples validates the README.md quick start examples.
// These tests ensure the documented installation and usage patterns work correctly.
func TestQuickStartExamples(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	// Test the basic fake profile example from README
	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("README quick start example failed: %v", err)
	}
	if job.State != JobSucceeded {
		t.Fatalf("README quick start example did not succeed: got %s", job.State)
	}

	// Verify the fake profile behaves as expected for documented usage
	if job.Backend != "fake" {
		t.Errorf("Expected backend 'fake', got '%s'", job.Backend)
	}
	if job.Cleanup.State != "complete" {
		t.Errorf("Expected cleanup complete, got '%s'", job.Cleanup.State)
	}
	
	// Validate actual output content from the echo command
	expectedOutput := "hello
"
	if job.Output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, job.Output)
	}
}