package main

import (
  "fmt"
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
  if !strings.Contains(string(output), fmt.Sprintf("profile %q not found", "missing")) {
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