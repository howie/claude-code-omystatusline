package main

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
)

func TestFormatSegments(t *testing.T) {
	// 建立超出寬度的 segments（兩段各 60 字元，共 120 > 80）
	segs := []statusline.Segment{
		{Content: strings.Repeat("A", 60), Priority: 1},
		{Content: " | " + strings.Repeat("B", 60), Priority: 3},
	}

	t.Run("truncate mode produces one line with ellipsis", func(t *testing.T) {
		got := formatSegments(segs, 80, "truncate")
		if strings.Contains(got, "\n") {
			t.Error("overflow_mode=truncate should not wrap to second line")
		}
		if !strings.Contains(got, "…") {
			t.Error("overflow_mode=truncate should append ellipsis when truncating")
		}
	})

	t.Run("wrap mode produces two lines", func(t *testing.T) {
		got := formatSegments(segs, 80, "wrap")
		if !strings.Contains(got, "\n") {
			t.Error("overflow_mode=wrap should produce two lines when content exceeds width")
		}
	})

	t.Run("unknown mode falls back to wrap", func(t *testing.T) {
		got := formatSegments(segs, 80, "unknown_value")
		if !strings.Contains(got, "\n") {
			t.Error("unknown overflow_mode should fall back to wrap behavior")
		}
	})

	t.Run("empty string mode falls back to wrap", func(t *testing.T) {
		got := formatSegments(segs, 80, "")
		if !strings.Contains(got, "\n") {
			t.Error("empty overflow_mode should fall back to wrap behavior")
		}
	})
}

// TestContextWindowMaxTokens 驗證 maxTokens 永遠使用 contextWindowForModel（非 ContextWindowSize）。
// ContextWindowSize from Claude Code is the current token count, not the model's max capacity.
// Using it as the denominator causes percentage ≈ 100% for all sessions (the bug this fixes).
func TestContextWindowMaxTokens(t *testing.T) {
	orig := context.RenderMode
	context.RenderMode = terminal.ModeTrueColor
	defer func() { context.RenderMode = orig }()

	t.Run("contextWindowForModel is the denominator for sonnet-4-6", func(t *testing.T) {
		// Bug scenario: ContextWindowSize (207k) equals token usage → old code showed 100%.
		// New code always uses contextWindowForModel (1M) → ~20%.
		tokens := 207_000
		maxTokens := contextWindowForModel("claude-sonnet-4-6") // must be 1_000_000
		ctxData := context.BuildFromTokens(tokens, maxTokens)
		if ctxData.NoUsageData {
			t.Error("BuildFromTokens must not set NoUsageData")
		}
		if ctxData.Tokens != tokens {
			t.Errorf("Tokens = %d, want %d", ctxData.Tokens, tokens)
		}
		wantPct := tokens * 100 / 1_000_000 // ~20, not 100
		if ctxData.Percentage != wantPct {
			t.Errorf("Percentage = %d, want %d (bug: ContextWindowSize as denominator gives 100%%)", ctxData.Percentage, wantPct)
		}
		if ctxData.Percentage == 100 {
			t.Error("Percentage must not be 100 when tokens=207k and model max=1M (regression: ContextWindowSize bug)")
		}
	})

	t.Run("contextWindowForModel returns 1M for sonnet-4-6", func(t *testing.T) {
		got := contextWindowForModel("claude-sonnet-4-6")
		if got != 1_000_000 {
			t.Errorf("contextWindowForModel(sonnet-4-6) = %d, want 1000000", got)
		}
	})

	t.Run("ContextWindow with zero tokens shows 0pct not NoUsageData", func(t *testing.T) {
		// Session started but no API call yet: context_window_size>0, current_usage all zeros.
		ctxData := context.BuildFromTokens(0, 200_000)
		if ctxData.NoUsageData {
			t.Error("zero tokens from ContextWindow must show 0%%, not 📡 (NoUsageData)")
		}
		if ctxData.Percentage != 0 {
			t.Errorf("Percentage = %d, want 0", ctxData.Percentage)
		}
		if strings.Contains(ctxData.Info, "📡") {
			t.Errorf("Info must not contain 📡 for ContextWindow zero-token session, got %q", ctxData.Info)
		}
	})

	t.Run("tokens exceeding maxTokens clamps to 100pct", func(t *testing.T) {
		ctxData := context.BuildFromTokens(600_000, 500_000)
		if ctxData.Percentage != 100 {
			t.Errorf("Percentage = %d, want 100 when tokens > maxTokens", ctxData.Percentage)
		}
		if ctxData.Tokens != 600_000 {
			t.Errorf("Tokens = %d, want 600000 (raw value preserved)", ctxData.Tokens)
		}
		if !strings.Contains(ctxData.Info, "600k") {
			t.Errorf("Info should show raw token count even when clamped, got %q", ctxData.Info)
		}
	})
}

func TestClaudeModelVersion(t *testing.T) {
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
		// date-only format (no minor) → not parseable as valid version
		{"claude-sonnet-4-20250514", -1, -1},
		// trailing prefix, no digit
		{"claude-sonnet-4-", -1, -1},
		// unknown format
		{"claude-future-1", -1, -1},
		{"", -1, -1},
		{"sonnet-4-10", 4, 10},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			maj, min := claudeModelVersion(strings.ToLower(tc.id))
			if maj != tc.wantMajor || min != tc.wantMinor {
				t.Errorf("claudeModelVersion(%q) = (%d, %d), want (%d, %d)", tc.id, maj, min, tc.wantMajor, tc.wantMinor)
			}
		})
	}
}

func TestContextWindowForModel(t *testing.T) {
	cases := []struct {
		name    string
		modelID string
		want    int
	}{
		// Haiku: 200K（官方規格，不變）
		{"haiku-base", "claude-haiku-4-5", 200_000},
		{"haiku-dated-suffix", "claude-haiku-4-5-20251001", 200_000},
		{"haiku-uppercase", "Claude-Haiku-4-5", 200_000},
		// Sonnet 4.6+: 1M（官方規格，minor >= 6）
		{"sonnet-46", "claude-sonnet-4-6", 1_000_000},
		{"sonnet-uppercase", "CLAUDE-SONNET-4-6", 1_000_000},
		{"sonnet-47-future", "claude-sonnet-4-7", 1_000_000},
		// Sonnet 5.x：未來大版本，major >= 5 → 1M
		{"sonnet-50-future-major", "claude-sonnet-5-0", 1_000_000},
		// Sonnet 4.5 以下: 200K（官方規格）
		{"sonnet-45", "claude-sonnet-4-5", 200_000},
		{"sonnet-45-dated", "claude-sonnet-4-5-20250929", 200_000},
		// Opus 4.6+: 1M（官方規格，minor >= 6）
		{"opus-47", "claude-opus-4-7", 1_000_000},
		{"opus-46", "claude-opus-4-6", 1_000_000},
		{"opus-48-future", "claude-opus-4-8", 1_000_000},
		// Opus 5.x：未來大版本，major >= 5 → 1M
		{"opus-50-future-major", "claude-opus-5-0", 1_000_000},
		// Opus 4.5 以下: 200K（官方規格）
		{"opus-45", "claude-opus-4-5", 200_000},
		{"opus-45-dated", "claude-opus-4-5-20251101", 200_000},
		{"opus-41", "claude-opus-4-1-20250805", 200_000},
		// 未知非空模型：保守 fallback 200K
		{"unknown-future", "claude-future-1", 200_000},
		// 空字串：DefaultMaxTokens
		{"empty-fallback", "", context.DefaultMaxTokens},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := contextWindowForModel(tc.modelID)
			if got != tc.want {
				t.Errorf("contextWindowForModel(%q) = %d, want %d", tc.modelID, got, tc.want)
			}
		})
	}
}
