package main

import (
  "os/exec"
  "testing"

  "github.com/mcintosh1776/workcell/internal/workcell"
)

func TestProfilesCommandOutput(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profiles")
  output, err := cmd.CombinedOutput()
  if err != nil {
    t.Fatalf("profiles command failed: %v output=%s", err, output)
  }
  got := string(output)
  want := workcell.ProfileListOutput(workcell.DefaultProfiles()) + "\n"
  if got != want {
    t.Fatalf("profiles output = %q, want %q", got, want)
  }
}