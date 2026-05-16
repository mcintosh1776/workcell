package main

import (
  "os/exec"
  "strings"
  "testing"
)

func TestProfileCommandOutput(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile", "fake")
  output, err := cmd.CombinedOutput()
  if err != nil {
    t.Fatalf("profile command failed: %v output=%s", err, output)
  }
  got := string(output)
  for _, want := range []string{"id: fake", "backend: fake"} {
    if !strings.Contains(got, want) {
      t.Fatalf("profile output = %q, missing %q", got, want)
    }
  }
}

func TestProfileCommandRejectsUnknownProfile(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile", "missing")
  output, err := cmd.CombinedOutput()
  if err == nil {
    t.Fatalf("profile missing command succeeded unexpectedly: %s", output)
  }
  if !strings.Contains(string(output), "profile "missing" not found") {
    t.Fatalf("profile missing output = %q, want not found message", output)
  }
}