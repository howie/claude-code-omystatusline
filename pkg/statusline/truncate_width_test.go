package statusline

import "testing"

func TestTruncateToWidth(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		maxWidth int
		want     string
	}{
		{"fits", "abc", 5, "abc"},
		{"exact", "abc", 3, "abc"},
		{"truncate-ascii", "abcdef", 3, "abc"},
		{"zero-width", "abc", 0, ""},
		{"negative-width", "abc", -1, ""},
		{"empty-input", "", 5, ""},
		// Wide char must not be split: あ is width 2.
		{"wide-fits", "aあ", 3, "aあ"},
		{"wide-would-overflow", "aあb", 2, "a"},
		{"wide-first-too-big", "あ", 1, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TruncateToWidth(tc.in, tc.maxWidth)
			if got != tc.want {
				t.Errorf("TruncateToWidth(%q, %d) = %q, want %q", tc.in, tc.maxWidth, got, tc.want)
			}
			if VisibleWidth(got) > tc.maxWidth && tc.maxWidth > 0 {
				t.Errorf("TruncateToWidth(%q, %d) = %q has visible width %d > %d",
					tc.in, tc.maxWidth, got, VisibleWidth(got), tc.maxWidth)
			}
		})
	}
}

// ANSI color sequences contribute 0 width and must survive truncation, with a
// trailing reset appended so a cut inside colored text does not bleed.
func TestTruncateToWidthPreservesANSI(t *testing.T) {
	in := "\033[31mabcdef\033[0m"
	got := TruncateToWidth(in, 3)
	if VisibleWidth(got) != 3 {
		t.Errorf("visible width = %d, want 3 (got %q)", VisibleWidth(got), got)
	}
	if !containsRune(got, '\033') {
		t.Errorf("expected color escape preserved, got %q", got)
	}
	// Must end with a reset to avoid color bleed after a mid-color cut.
	if !hasSuffix(got, ColorReset) {
		t.Errorf("expected trailing reset %q, got %q", ColorReset, got)
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
