package workcell

import "testing"

func TestVersionReturnsDevelopmentVersion(t *testing.T) {
	if Version() != "workcell dev" {
		t.Fatalf("Version() = %q, want %q", Version(), "workcell dev")
	}
}