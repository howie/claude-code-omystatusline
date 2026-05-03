package main

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
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
