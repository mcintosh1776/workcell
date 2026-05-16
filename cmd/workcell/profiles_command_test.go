package main

import (
  "os/exec"
  "testing"
)

func TestProfilesCommandOutput(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profiles")
  output, err := cmd.CombinedOutput()
  if err != nil {
    t.Fatalf("profiles command failed: %v output=%s", err, output)
  }
  got := string(output)
  want := "fake
podman-smoke
"
  if got != want {
    t.Fatalf("profiles output = %q, want %q", got, want)
  }
}