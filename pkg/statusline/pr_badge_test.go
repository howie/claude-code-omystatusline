package statusline

import (
	"strings"
	"testing"
)

// TestFormatPRBadge 驗證 PR 徽章格式化：zero-value 隱藏、各 review state 的 glyph/色、
// OSC 8 超連結包裹（hyperlink）與 ASCII 降級。
func TestFormatPRBadge(t *testing.T) {
	t.Run("zero number hides badge", func(t *testing.T) {
		if got := FormatPRBadge(0, "https://x/pull/1", "approved", " | ", true); got != "" {
			t.Errorf("number=0 should hide badge, got %q", got)
		}
	})

	t.Run("approved shows check glyph in green", func(t *testing.T) {
		got := FormatPRBadge(1234, "", "approved", " | ", false)
		if !strings.Contains(got, "PR #1234") {
			t.Errorf("missing PR number, got %q", got)
		}
		if !strings.Contains(got, "✓") { // ✓
			t.Errorf("approved should show check, got %q", got)
		}
		if !strings.Contains(got, ColorGreen) {
			t.Errorf("approved glyph should be green, got %q", got)
		}
	})

	t.Run("changes_requested shows cross in red", func(t *testing.T) {
		got := FormatPRBadge(7, "", "changes_requested", " | ", false)
		if !strings.Contains(got, "✗") || !strings.Contains(got, ColorRed) { // ✗
			t.Errorf("changes_requested should show red cross, got %q", got)
		}
	})

	t.Run("review state is case-insensitive", func(t *testing.T) {
		got := FormatPRBadge(1, "", "APPROVED", " | ", false)
		if !strings.Contains(got, "✓") || !strings.Contains(got, ColorGreen) { // ✓
			t.Errorf("uppercase APPROVED should still show green check, got %q", got)
		}
	})

	t.Run("negative number hides badge", func(t *testing.T) {
		if got := FormatPRBadge(-1, "", "approved", " | ", false); got != "" {
			t.Errorf("negative number should hide badge, got %q", got)
		}
	})

	t.Run("unknown review state shows no glyph", func(t *testing.T) {
		got := FormatPRBadge(9, "", "", " | ", false)
		if strings.Contains(got, "✓") || strings.Contains(got, "✗") || strings.Contains(got, "\U0001F4AC") {
			t.Errorf("empty review state should have no glyph, got %q", got)
		}
		if !strings.Contains(got, "PR #9") {
			t.Errorf("missing PR number, got %q", got)
		}
	})

	t.Run("url with hyperlink wraps in OSC 8", func(t *testing.T) {
		url := "https://github.com/o/r/pull/5"
		got := FormatPRBadge(5, url, "approved", " | ", true)
		if !strings.Contains(got, "\033]8;;"+url+"\033\\") {
			t.Errorf("hyperlink should wrap text in OSC 8 open, got %q", got)
		}
		if !strings.Contains(got, "\033]8;;\033\\") {
			t.Errorf("hyperlink should include OSC 8 close, got %q", got)
		}
	})

	t.Run("ascii mode degrades to plain text", func(t *testing.T) {
		got := FormatPRBadge(5, "https://x/pull/5", "approved", " | ", false)
		if strings.Contains(got, "\033]8;;") {
			t.Errorf("ascii mode must not emit OSC 8, got %q", got)
		}
	})

	t.Run("OSC 8 wrapping does not change visible width", func(t *testing.T) {
		url := "https://github.com/anthropics/claude-code/pull/1234"
		linked := FormatPRBadge(1234, url, "approved", " | ", true)
		plain := FormatPRBadge(1234, "", "approved", " | ", false)
		if VisibleWidth(linked) != VisibleWidth(plain) {
			t.Errorf("OSC 8 link width %d != plain width %d (URL must be zero-width)",
				VisibleWidth(linked), VisibleWidth(plain))
		}
	})
}

// TestVisibleWidthOSC8 驗證 VisibleWidth 跳過 OSC 8 超連結：URL 與包裹序列貢獻 0 寬，
// 只算被包住的連結文字。
func TestVisibleWidthOSC8(t *testing.T) {
	linked := "\033]8;;https://example.com/very/long/url\033\\PR #1\033]8;;\033\\"
	if got, want := VisibleWidth(linked), VisibleWidth("PR #1"); got != want {
		t.Errorf("VisibleWidth(OSC8) = %d, want %d (= width of plain text)", got, want)
	}

	// BEL (0x07) 作為 OSC 終止符也應正確處理。
	belTerminated := "\033]8;;http://x\aPR\033]8;;\a"
	if got, want := VisibleWidth(belTerminated), VisibleWidth("PR"); got != want {
		t.Errorf("VisibleWidth(BEL-terminated OSC8) = %d, want %d", got, want)
	}
}
