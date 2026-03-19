package speed

import (
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func TestExtractOutputTokens(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"output_tokens": float64(500),
				},
			},
		}},
	}

	result := extractOutputTokens(lines)
	if result != 500 {
		t.Fatalf("expected 500, got %d", result)
	}
}

func TestExtractOutputTokensEmpty(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "user"}},
	}

	result := extractOutputTokens(lines)
	if result != 0 {
		t.Fatalf("expected 0, got %d", result)
	}
}

func TestFormat(t *testing.T) {
	info := &SpeedInfo{TokensPerSec: 42}
	result := Format(info)
	if result != "42 tok/s" {
		t.Fatalf("expected '42 tok/s', got %q", result)
	}

	if Format(nil) != "" {
		t.Fatal("expected empty string for nil")
	}
}
