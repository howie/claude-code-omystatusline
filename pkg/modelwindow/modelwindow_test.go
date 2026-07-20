package modelwindow

import "testing"

func TestInfer(t *testing.T) {
	cases := []struct {
		name          string
		modelID       string
		wantSize      int
		wantConfident bool
	}{
		// Empty: fail-safe fallback, NOT confident (callers must not draw a % from a guess).
		{"empty", "", Window200K, false},
		// Explicit [1m] marker wins over everything, confident.
		{"opus-1m-marker", "claude-opus-4-8[1m]", Window1M, true},
		{"sonnet-1m-marker-uppercase", "claude-sonnet-4-5[1M]", Window1M, true},
		// Haiku is always 200K, confident.
		{"haiku", "claude-haiku-4-5", Window200K, true},
		{"haiku-dated", "claude-haiku-4-5-20251001", Window200K, true},
		// Known families by version, confident either way.
		{"sonnet-46-1m", "claude-sonnet-4-6", Window1M, true},
		{"sonnet-45-200k", "claude-sonnet-4-5", Window200K, true},
		{"opus-48-1m", "claude-opus-4-8", Window1M, true},
		{"opus-45-200k", "claude-opus-4-5", Window200K, true},
		{"fable-5-1m", "claude-fable-5", Window1M, true},
		{"sonnet-5-default-1m", "claude-sonnet-5", Window1M, true},
		// Known family with unparseable version defaults to 200K but stays confident
		// (a sonnet/opus/fable id is still a recognized family).
		{"sonnet-unparseable-version", "claude-sonnet-20250514", Window200K, true},
		// Unknown family resolved by version rule (major>=5 or 4.6+): confident 1M.
		{"unknown-family-major5", "claude-mythos-5", Window1M, true},
		{"unknown-family-4-6", "claude-nova-4-6", Window1M, true},
		// Unknown family that only reaches the 200K guess: NOT confident.
		{"unknown-family-guess-200k", "claude-future-1", Window200K, false},
		{"unknown-family-no-version", "totally-unknown", Window200K, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			size, confident := Infer(tc.modelID)
			if size != tc.wantSize || confident != tc.wantConfident {
				t.Errorf("Infer(%q) = (%d, %v), want (%d, %v)",
					tc.modelID, size, confident, tc.wantSize, tc.wantConfident)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	cases := []struct {
		id        string
		wantMajor int
		wantMinor int
	}{
		{"claude-sonnet-4-6", 4, 6},
		{"claude-opus-4-7-20251001", 4, 7},
		{"claude-sonnet-4-5-20250929", 4, 5},
		{"claude-sonnet-5-0", 5, 0},
		{"claude-opus-4-1-20250805", 4, 1},
		{"claude-haiku-4-5", 4, 5},
		// Single version number (no minor) falls back to {major}.0.
		{"claude-fable-5", 5, 0},
		{"claude-sonnet-4-20250514", 4, 0},
		{"claude-sonnet-4-", 4, 0},
		{"claude-future-1", 1, 0},
		// No parseable integer at all.
		{"claude-sonnet-20250514", -1, -1},
		{"", -1, -1},
		{"sonnet-4-10", 4, 10},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			maj, min := version(tc.id)
			if maj != tc.wantMajor || min != tc.wantMinor {
				t.Errorf("version(%q) = (%d, %d), want (%d, %d)", tc.id, maj, min, tc.wantMajor, tc.wantMinor)
			}
		})
	}
}
