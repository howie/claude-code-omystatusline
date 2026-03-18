package todo

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func TestAnalyzeNoTodo(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "user"}},
	}
	result := Analyze(lines)
	if result != nil {
		t.Fatal("expected nil for no TodoWrite")
	}
}

func TestAnalyzeTodoInProgress(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type": "tool_use",
						"name": "TodoWrite",
						"input": map[string]interface{}{
							"todos": []interface{}{
								map[string]interface{}{"content": "Task A", "status": "completed"},
								map[string]interface{}{"content": "Task B", "status": "in_progress"},
								map[string]interface{}{"content": "Task C", "status": "pending"},
							},
						},
					},
				},
			},
		}},
	}

	result := Analyze(lines)
	if result == nil {
		t.Fatal("expected TodoInfo")
	}
	if result.Total != 3 {
		t.Fatalf("expected 3 total, got %d", result.Total)
	}
	if result.Completed != 1 {
		t.Fatalf("expected 1 completed, got %d", result.Completed)
	}
	if result.InProgressName != "Task B" {
		t.Fatalf("expected in-progress 'Task B', got %q", result.InProgressName)
	}
	if result.AllComplete {
		t.Fatal("expected AllComplete to be false")
	}
}

func TestAnalyzeTodoAllComplete(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type": "tool_use",
						"name": "TodoWrite",
						"input": map[string]interface{}{
							"todos": []interface{}{
								map[string]interface{}{"content": "Done 1", "status": "completed"},
								map[string]interface{}{"content": "Done 2", "status": "completed"},
							},
						},
					},
				},
			},
		}},
	}

	result := Analyze(lines)
	if result == nil {
		t.Fatal("expected TodoInfo")
	}
	if !result.AllComplete {
		t.Fatal("expected AllComplete to be true")
	}
}

func TestFormat(t *testing.T) {
	// In progress
	info := &TodoInfo{InProgressName: "Build features", Completed: 2, Total: 5}
	result := Format(info)
	if !strings.Contains(result, "▸ Build features") {
		t.Fatalf("expected in-progress format, got %q", result)
	}
	if !strings.Contains(result, "(2/5)") {
		t.Fatalf("expected progress count, got %q", result)
	}

	// All complete
	info = &TodoInfo{AllComplete: true, Completed: 3, Total: 3}
	result = Format(info)
	if !strings.Contains(result, "✓ All complete") {
		t.Fatalf("expected all complete format, got %q", result)
	}

	// Nil
	if Format(nil) != "" {
		t.Fatal("expected empty string for nil")
	}
}
