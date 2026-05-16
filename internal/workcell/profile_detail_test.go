package workcell

import (
  "strings"
  "testing"
)

func TestProfileDetailOutputForFakeProfile(t *testing.T) {
  got, err := ProfileDetailOutput(DefaultProfiles(), "fake")
  if err != nil {
    t.Fatalf("ProfileDetailOutput(fake) error = %v", err)
  }
  for _, want := range []string{"id: fake", "backend: fake"} {
    if !strings.Contains(got, want) {
      t.Fatalf("ProfileDetailOutput(fake) = %q, missing %q", got, want)
    }
  }
}

func TestProfileDetailOutputForPodmanProfile(t *testing.T) {
  got, err := ProfileDetailOutput(DefaultProfiles(), "podman-smoke")
  if err != nil {
    t.Fatalf("ProfileDetailOutput(podman-smoke) error = %v", err)
  }
  for _, want := range []string{
    "id: podman-smoke",
    "backend: podman",
    "image: docker.io/library/alpine:3.20",
    "timeoutSeconds: 300",
  } {
    if !strings.Contains(got, want) {
      t.Fatalf("ProfileDetailOutput(podman-smoke) = %q, missing %q", got, want)
    }
  }
}

func TestProfileDetailOutputRejectsUnknownProfile(t *testing.T) {
  if _, err := ProfileDetailOutput(DefaultProfiles(), "missing"); err == nil {
    t.Fatal("ProfileDetailOutput(missing) error = nil, want error")
  }
}

func TestProfileDetailOutputRejectsEmptyProfile(t *testing.T) {
  if _, err := ProfileDetailOutput(DefaultProfiles(), ""); err == nil {
    t.Fatal("ProfileDetailOutput(empty) error = nil, want error")
  }
}

func TestProfileDetailOutputRejectsWhitespaceProfile(t *testing.T) {
  if _, err := ProfileDetailOutput(DefaultProfiles(), "   "); err == nil {
    t.Fatal("ProfileDetailOutput(whitespace) error = nil, want error")
  }
}

func TestProfileDetailOutputRejectsNegativeTimeout(t *testing.T) {
  profiles := map[string]Profile{
    "bad-timeout": {
      ID:      "bad-timeout",
      Backend: "fake",
      BackendConfig: BackendConfig{
        Timeout: -1,
      },
    },
  }
  if _, err := ProfileDetailOutput(profiles, "bad-timeout"); err == nil {
    t.Fatal("ProfileDetailOutput(bad-timeout) error = nil, want error")
  }
}