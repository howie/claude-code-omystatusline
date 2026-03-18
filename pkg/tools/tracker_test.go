package tools

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func TestAnalyzeNoTools(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "user"}},
	}
	result := Analyze(lines)
	if len(result) != 0 {
		t.Fatalf("expected no tools, got %d", len(result))
	}
}

func TestAnalyzeActiveTools(t *testing.T) {
	lines := []transcript.Line{
		// tool_use without matching tool_result
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type":  "tool_use",
						"id":    "tool1",
						"name":  "Read",
						"input": map[string]interface{}{"file_path": "/src/main.go"},
					},
				},
			},
		}},
		// another active tool
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type":  "tool_use",
						"id":    "tool2",
						"name":  "Write",
						"input": map[string]interface{}{"file_path": "/pkg/config.go"},
					},
				},
			},
		}},
	}

	result := Analyze(lines)
	if len(result) != 2 {
		t.Fatalf("expected 2 active tools, got %d", len(result))
	}
	if result[0].Name != "Read" {
		t.Fatalf("expected first tool 'Read', got %q", result[0].Name)
	}
	if result[1].Name != "Write" {
		t.Fatalf("expected second tool 'Write', got %q", result[1].Name)
	}
}

func TestAnalyzeCompletedTool(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type":  "tool_use",
						"id":    "tool1",
						"name":  "Read",
						"input": map[string]interface{}{},
					},
				},
			},
		}},
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": "tool1",
					},
				},
			},
		}},
	}

	result := Analyze(lines)
	if len(result) != 0 {
		t.Fatalf("expected no active tools (tool completed), got %d", len(result))
	}
}

func TestFormat(t *testing.T) {
	tools := []ToolInfo{
		{Name: "Read", Target: "/src/main.go"},
		{Name: "Write", Target: ""},
	}

	result := Format(tools)
	if !strings.Contains(result, "◐ Read: /src/main.go") {
		t.Fatalf("expected formatted tool with target, got %q", result)
	}
	if !strings.Contains(result, "◐ Write") {
		t.Fatalf("expected formatted tool without target, got %q", result)
	}
}

func TestTruncatePath(t *testing.T) {
	long := "/home/user/very/long/path/to/some/deeply/nested/file.go"
	result := truncatePath(long, 30)
	if len(result) > 30 {
		t.Fatalf("expected truncated path <= 30 chars, got %d: %q", len(result), result)
	}
}
