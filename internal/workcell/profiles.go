package workcell

import (
  "sort"
  "strings"
)

// SortedProfileIDs returns profile IDs in deterministic display order.
func SortedProfileIDs(profiles map[string]Profile) []string {
  ids := make([]string, 0, len(profiles))
  for id := range profiles {
    ids = append(ids, id)
  }
  sort.Strings(ids)
  return ids
}

// ProfileListOutput returns profile IDs formatted for CLI output.
func ProfileListOutput(profiles map[string]Profile) string {
  return strings.Join(SortedProfileIDs(profiles), "\n")
}