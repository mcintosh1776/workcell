package workcell

import (
  "fmt"
  "strings"
)

// ProfileDetailOutput returns one profile's details in deterministic display order.
func ProfileDetailOutput(profiles map[string]Profile, profileID string) (string, error) {
  profileID = strings.TrimSpace(profileID)
  if profileID == "" {
    return "", fmt.Errorf("profile id required")
  }
  if !validProfileID(profileID) {
    return "", fmt.Errorf("profile %q is invalid", profileID)
  }
  profile, ok := profiles[profileID]
  if !ok {
    return "", fmt.Errorf("profile %q not found", profileID)
  }
  if profile.BackendConfig.Timeout < 0 {
    return "", fmt.Errorf("profile %q has invalid timeout %d", profileID, profile.BackendConfig.Timeout)
  }

  lines := []string{
    "id: " + profile.ID,
    "backend: " + profile.Backend,
  }
  if profile.BackendConfig.Image != "" {
    lines = append(lines, "image: "+profile.BackendConfig.Image)
  }
  if profile.BackendConfig.Timeout > 0 {
    lines = append(lines, fmt.Sprintf("timeoutSeconds: %d", profile.BackendConfig.Timeout))
  }
  return strings.Join(lines, "\n"), nil
}

func validProfileID(profileID string) bool {
  for _, ch := range profileID {
    if ch >= 'a' && ch <= 'z' {
      continue
    }
    if ch >= 'A' && ch <= 'Z' {
      continue
    }
    if ch >= '0' && ch <= '9' {
      continue
    }
    if ch == '-' || ch == '_' {
      continue
    }
    return false
  }
  return true
}