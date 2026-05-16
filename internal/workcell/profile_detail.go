package workcell

import (
  "fmt"
  "strings"
)

// ProfileDetailOutput returns one profile's details in deterministic display order.
func ProfileDetailOutput(profiles map[string]Profile, profileID string) (string, error) {
  profile, ok := profiles[profileID]
  if !ok {
    return "", fmt.Errorf("profile %q not found", profileID)
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
  return strings.Join(lines, string([]byte{10})), nil
}