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

// TestContextWindowMaxTokens 驗證 maxTokens 優先順序：
// ContextWindowSize > 0 時優先使用，否則 fallback 到 contextWindowForModel。
func TestContextWindowMaxTokens(t *testing.T) {
	orig := context.RenderMode
	context.RenderMode = terminal.ModeTrueColor
	defer func() { context.RenderMode = orig }()

	t.Run("ContextWindowSize>0 takes priority over model inference", func(t *testing.T) {
		// 93545 tokens / 500000 window = 18%
		tokens := 1 + 694 + 92850
		ctxData := context.BuildFromTokens(tokens, 500_000)
		if ctxData.NoUsageData {
			t.Error("ContextWindow path must not set NoUsageData")
		}
		if ctxData.Tokens != tokens {
			t.Errorf("Tokens = %d, want %d", ctxData.Tokens, tokens)
		}
		wantPct := tokens * 100 / 500_000
		if ctxData.Percentage != wantPct {
			t.Errorf("Percentage = %d, want %d", ctxData.Percentage, wantPct)
		}
	})

	t.Run("zero ContextWindowSize falls back to model inference", func(t *testing.T) {
		// ContextWindowSize==0 → contextWindowForModel used; verify the function exists
		got := contextWindowForModel("claude-sonnet-4-6")
		if got != 500_000 {
			t.Errorf("contextWindowForModel(sonnet) = %d, want 500000", got)
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

func TestContextWindowForModel(t *testing.T) {
	cases := []struct {
		name    string
		modelID string
		want    int
	}{
		// Haiku: 200K（不變）
		{"haiku-base", "claude-haiku-4-5", 200_000},
		{"haiku-dated-suffix", "claude-haiku-4-5-20251001", 200_000},
		{"haiku-uppercase", "Claude-Haiku-4-5", 200_000},
		// Sonnet: 500K（經驗校準，反推自 ~357K compact 點 ÷ 70%）
		{"sonnet-46", "claude-sonnet-4-6", 500_000},
		{"sonnet-45", "claude-sonnet-4-5", 500_000},
		{"sonnet-uppercase", "CLAUDE-SONNET-4-6", 500_000},
		// Opus: 800K（Opus context 較大，預留更多空間）
		{"opus-47", "claude-opus-4-7", 800_000},
		{"opus-46", "claude-opus-4-6", 800_000},
		// 未知非空模型：保守 fallback 500K
		{"unknown-future", "claude-future-1", 500_000},
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
