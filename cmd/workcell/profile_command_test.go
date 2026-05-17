package main

import (
  "os/exec"
  "strings"
  "testing"
)

func TestProfilesCommandStillListsDefaults(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profiles")
  output, err := cmd.CombinedOutput()
  if err != nil {
    t.Fatalf("profiles command failed: %v output=%s", err, output)
  }
  got := string(output)
  for _, want := range []string{"fake", "podman-smoke"} {
    if !strings.Contains(got, want) {
      t.Fatalf("profiles output = %q, missing %q", got, want)
    }
  }
}

func TestHelpIncludesProfileCommand(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "help")
  output, err := cmd.CombinedOutput()
  if err != nil {
    t.Fatalf("help command failed: %v output=%s", err, output)
  }
  if !strings.Contains(string(output), "workcell profile <id>") {
    t.Fatalf("help output = %q, missing profile command", output)
  }
}

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

func TestProfileCommandWithInsufficientArgs(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile")
  output, err := cmd.CombinedOutput()
  if err == nil {
    t.Fatalf("profile missing arg command succeeded unexpectedly: %s", output)
  }
  if !strings.Contains(string(output), "profile id required") {
    t.Fatalf("profile missing arg output = %q, want required message", output)
  }
}

func TestProfileCommandRejectsUnknownProfile(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile", "missing")
  output, err := cmd.CombinedOutput()
  if err == nil {
    t.Fatalf("profile missing command succeeded unexpectedly: %s", output)
  }
  if !strings.Contains(string(output), "profile \"missing\" not found") {
    t.Fatalf("profile missing output = %q, want not found message", output)
  }
}

func TestProfileCommandRejectsEmptyProfile(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile", "")
  output, err := cmd.CombinedOutput()
  if err == nil {
    t.Fatalf("profile empty command succeeded unexpectedly: %s", output)
  }
  if !strings.Contains(string(output), "profile id required") {
    t.Fatalf("profile empty output = %q, want required message", output)
  }
}

func TestProfileCommandRejectsWhitespaceProfile(t *testing.T) {
  cmd := exec.Command("go", "run", ".", "profile", "   ")
  output, err := cmd.CombinedOutput()
  if err == nil {
    t.Fatalf("profile whitespace command succeeded unexpectedly: %s", output)
  }
  if !strings.Contains(string(output), "profile id required") {
    t.Fatalf("profile whitespace output = %q, want required message", output)
  }
}

func TestProfileFunctionRejectsInvalidInput(t *testing.T) {
  for _, args := range [][]string{nil, {}, {""}, {"   "}, {"fake", "extra"}} {
    if err := profile(args); err == nil {
      t.Fatalf("profile(%q) error = nil, want error", args)
    }
  }
}