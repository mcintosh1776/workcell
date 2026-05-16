package workcell

import (
  "fmt"
  "strings"
)

// ProfileDetailOutput returns one profile's details in deterministic display order.
func ProfileDetailOutput(profiles map[string]Profile, profileID string) (string, error) {
  if strings.TrimSpace(profileID) == "" {
    return "", fmt.Errorf("profile id required")
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