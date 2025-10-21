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
