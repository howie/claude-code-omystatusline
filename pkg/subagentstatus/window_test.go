package subagentstatus

import (
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/modelwindow"
)

// TestResolveWindow covers the Q3 matrix: the two-part #35/#36 guard applied per task.
// A denominator is trusted only when it is a known capacity AND does not demote a
// confident inference; otherwise fall back to confident inference; otherwise token-only.
func TestResolveWindow(t *testing.T) {
	cases := []struct {
		name     string
		model    string
		reported int
		wantSize int
		wantOK   bool
	}{
		// known model, size omitted -> inferred.
		{"known-model-no-size", "claude-opus-4-8[1m]", 0, modelwindow.Window1M, true},
		{"known-200k-model-no-size", "claude-haiku-4-5", 0, modelwindow.Window200K, true},
		// model omitted, reported known capacity -> trust reported.
		{"no-model-reported-1m", "", modelwindow.Window1M, modelwindow.Window1M, true},
		{"no-model-reported-200k", "", modelwindow.Window200K, modelwindow.Window200K, true},
		// NON-DEMOTION (the core #35/#36 guard): confident 1M, reported 200K -> keep 1M.
		{"non-demoting-keeps-1m", "claude-opus-4-8[1m]", modelwindow.Window200K, modelwindow.Window1M, true},
		// reported >= inferred is honored (rescue an under-estimate).
		{"reported-rescues-to-1m", "claude-sonnet-4-5", modelwindow.Window1M, modelwindow.Window1M, true},
		// unknown model, non-capacity reported -> token-only (no guessed denominator).
		{"unknown-model-junk-size", "totally-unknown", 500000, 0, false},
		{"unknown-model-no-size", "totally-unknown", 0, 0, false},
		// both omitted -> token-only.
		{"both-omitted", "", 0, 0, false},
		// negative reported never used as a denominator.
		{"negative-reported-known-model", "claude-haiku-4-5", -5, modelwindow.Window200K, true},
		{"negative-reported-unknown-model", "totally-unknown", -5, 0, false},
		// unknown model but reported IS a known capacity -> trust the allow-set value.
		{"unknown-model-reported-200k", "totally-unknown", modelwindow.Window200K, modelwindow.Window200K, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			size, ok := resolveWindow(tc.model, tc.reported)
			if size != tc.wantSize || ok != tc.wantOK {
				t.Errorf("resolveWindow(%q, %d) = (%d, %v), want (%d, %v)",
					tc.model, tc.reported, size, ok, tc.wantSize, tc.wantOK)
			}
		})
	}
}
