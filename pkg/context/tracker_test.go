package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
)

func TestFormatNumber(t *testing.T) {
	tests := map[int]string{
		0:         "--",
		999:       "999",
		1500:      "1k",
		2_500_000: "2M",
	}

	for input, expected := range tests {
		t.Run(fmt.Sprintf("num_%d", input), func(t *testing.T) {
			if got := formatNumber(input); got != expected {
				t.Fatalf("formatNumber(%d) = %q, want %q", input, got, expected)
			}
		})
	}
}

func TestGetColor(t *testing.T) {
	if color := getColor(40); color != statusline.ColorCtxGreen {
		t.Fatalf("expected ColorCtxGreen for 40%%, got %q", color)
	}
	if color := getColor(70); color != statusline.ColorCtxGold {
		t.Fatalf("expected ColorCtxGold for 70%%, got %q", color)
	}
	if color := getColor(90); color != statusline.ColorCtxRed {
		t.Fatalf("expected ColorCtxRed for 90%%, got %q", color)
	}
}

func TestAnalyzeMaxTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.log")

	// 寫入一個有 100000 tokens 的 transcript
	line := `{"message":{"usage":{"input_tokens":100000}},"isSidechain":false}`
	if err := os.WriteFile(path, []byte(line), 0644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	tests := []struct {
		name      string
		maxTokens int
		wantPct   string // 預期百分比字串
	}{
		{"default (200k)", 0, "50%"},
		{"200k explicit", 200000, "50%"},
		{"1M extended", 1000000, "10%"},
		{"negative fallback", -1, "50%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Analyze(path, tt.maxTokens)
			if !strings.Contains(result, tt.wantPct) {
				t.Fatalf("Analyze with maxTokens=%d: expected %q in result, got %q", tt.maxTokens, tt.wantPct, result)
			}
		})
	}
}

func TestAnalyzeEmptyPath(t *testing.T) {
	result := Analyze("", 0)
	if !strings.Contains(result, "0%") {
		t.Fatalf("Analyze with empty path should show 0%%, got %q", result)
	}
}

func TestCalculateUsage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.log")

	lines := []string{
		`{"message":{"usage":{"input_tokens":10}},"isSidechain":true}`,
		`not json`,
		`{"message":{"usage":{"input_tokens":100,"cache_read_input_tokens":50,"cache_creation_input_tokens":25}},"isSidechain":false}`,
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	total := calculateUsage(path)
	if total != 175 {
		t.Fatalf("expected total usage 175, got %d", total)
	}
}
