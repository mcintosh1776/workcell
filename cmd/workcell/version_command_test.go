package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionCommandOutput(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v output=%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != "workcell dev" {
		t.Fatalf("version output = %q, want %q", got, "workcell dev")
	}
}