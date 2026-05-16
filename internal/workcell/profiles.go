package workcell

import "sort"

// SortedProfileIDs returns profile IDs in deterministic display order.
func SortedProfileIDs(profiles map[string]Profile) []string {
  ids := make([]string, 0, len(profiles))
  for id := range profiles {
    ids = append(ids, id)
  }
  sort.Strings(ids)
  return ids
}