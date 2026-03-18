package context

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func TestDetectAutocompactNone(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "user"}},
	}
	result := DetectAutocompact(lines)
	if result.Detected {
		t.Fatal("expected no autocompact detected")
	}
}

func TestDetectAutocompactSummary(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "summary"}},
		{Parsed: map[string]interface{}{"type": "user"}},
		{Parsed: map[string]interface{}{"type": "summary"}},
	}
	result := DetectAutocompact(lines)
	if !result.Detected {
		t.Fatal("expected autocompact detected")
	}
	if result.Count != 2 {
		t.Fatalf("expected count 2, got %d", result.Count)
	}
}

func TestFormatAutocompact(t *testing.T) {
	// Not detected
	if FormatAutocompact(&AutocompactInfo{}) != "" {
		t.Fatal("expected empty for not detected")
	}

	// Single compression
	result := FormatAutocompact(&AutocompactInfo{Detected: true, Count: 1})
	if !strings.Contains(result, "⚠ compressed") {
		t.Fatalf("expected compressed warning, got %q", result)
	}

	// Multiple compressions
	result = FormatAutocompact(&AutocompactInfo{Detected: true, Count: 3})
	if !strings.Contains(result, "×3") {
		t.Fatalf("expected count in format, got %q", result)
	}
}
