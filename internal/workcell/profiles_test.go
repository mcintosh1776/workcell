package workcell

import "testing"

func TestSortedProfileIDs(t *testing.T) {
  got := SortedProfileIDs(map[string]Profile{
    "podman-smoke": {ID: "podman-smoke"},
    "fake": {ID: "fake"},
  })
  want := []string{"fake", "podman-smoke"}
  if len(got) != len(want) {
    t.Fatalf("len(SortedProfileIDs()) = %d, want %d", len(got), len(want))
  }
  for i := range want {
    if got[i] != want[i] {
      t.Fatalf("SortedProfileIDs()[%d] = %q, want %q", i, got[i], want[i])
    }
  }
}

func TestSortedProfileIDsHandlesEmptyProfiles(t *testing.T) {
  got := SortedProfileIDs(map[string]Profile{})
  if len(got) != 0 {
    t.Fatalf("len(SortedProfileIDs(empty)) = %d, want 0", len(got))
  }
}

func TestSortedProfileIDsForDefaultProfiles(t *testing.T) {
  got := SortedProfileIDs(DefaultProfiles())
  want := []string{"fake", "podman-smoke"}
  if len(got) != len(want) {
    t.Fatalf("len(SortedProfileIDs(DefaultProfiles())) = %d, want %d", len(got), len(want))
  }
  for i := range want {
    if got[i] != want[i] {
      t.Fatalf("SortedProfileIDs(DefaultProfiles())[%d] = %q, want %q", i, got[i], want[i])
    }
  }
}