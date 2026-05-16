package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/mcintosh1776/workcell/internal/workcell"
)

func TestVersionCommandOutput(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v output=%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != workcell.Version() {
		t.Fatalf("version output = %q, want %q", got, workcell.Version())
	}
}