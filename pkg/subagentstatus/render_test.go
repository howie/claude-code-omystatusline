package subagentstatus

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
)

func init() {
	// Deterministic, color-free output for golden-style assertions.
	context.RenderMode = terminal.ModeASCII
}

func TestRenderFullRow(t *testing.T) {
	in := Input{
		Columns: 80,
		Tasks: []Task{{
			Name:              "verifier",
			Status:            "completed",
			Model:             "claude-opus-4-8[1m]",
			ContextWindowSize: 1_000_000,
			TokenCount:        159195,
		}},
	}
	got := Render(in)
	if strings.Count(got, "\n") != 0 {
		t.Fatalf("single task should be one line, got %q", got)
	}
	if !strings.Contains(got, "verifier") {
		t.Errorf("row missing name: %q", got)
	}
	if !strings.Contains(got, "15%") {
		t.Errorf("row missing percentage 15%%: %q", got)
	}
	if !strings.Contains(got, "159k") {
		t.Errorf("row missing token count 159k: %q", got)
	}
	if statusline.VisibleWidth(got) > 80 {
		t.Errorf("row width %d > 80: %q", statusline.VisibleWidth(got), got)
	}
}

func TestRenderTokenOnlyWhenNoDenominator(t *testing.T) {
	in := Input{
		Columns: 80,
		Tasks: []Task{{
			Name:       "scout",
			Status:     "running",
			Model:      "totally-unknown",
			TokenCount: 12345,
		}},
	}
	got := Render(in)
	if !strings.Contains(got, "scout") {
		t.Errorf("row missing name: %q", got)
	}
	if strings.Contains(got, "%") {
		t.Errorf("token-only row must not show a percentage: %q", got)
	}
	if !strings.Contains(got, "12k") {
		t.Errorf("row missing token count 12k: %q", got)
	}
}

func TestRenderNameFallback(t *testing.T) {
	in := Input{
		Columns: 80,
		Tasks: []Task{{
			// no Name, no Label -> falls back to Type.
			Type:       "local_agent",
			Status:     "running",
			Model:      "claude-haiku-4-5",
			TokenCount: 1000,
		}},
	}
	got := Render(in)
	if !strings.Contains(got, "local_agent") {
		t.Errorf("expected type fallback, got %q", got)
	}
}

func TestRenderEmptyTasksIsZeroBytes(t *testing.T) {
	if got := Render(Input{Columns: 80}); got != "" {
		t.Errorf("empty tasks must render zero bytes, got %q", got)
	}
	if got := Render(Input{Columns: 80, Tasks: []Task{}}); got != "" {
		t.Errorf("empty task slice must render zero bytes, got %q", got)
	}
}

func TestRenderNarrowColumnsHardBound(t *testing.T) {
	in := Input{
		Columns: 8,
		Tasks: []Task{{
			Name:              "a-very-long-subagent-name-that-overflows",
			Status:            "running",
			Model:             "claude-opus-4-8[1m]",
			ContextWindowSize: 1_000_000,
			TokenCount:        500000,
		}},
	}
	got := Render(in)
	if statusline.VisibleWidth(got) > 8 {
		t.Errorf("row width %d > 8: %q", statusline.VisibleWidth(got), got)
	}
}

func TestRenderMultiTaskOrderPreserved(t *testing.T) {
	in := Input{
		Columns: 80,
		Tasks: []Task{
			{Name: "first", Status: "running", Model: "claude-haiku-4-5", TokenCount: 1000},
			{Name: "second", Status: "completed", Model: "claude-haiku-4-5", TokenCount: 2000},
		},
	}
	got := Render(in)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "first") || !strings.Contains(lines[1], "second") {
		t.Errorf("task order not preserved: %q", got)
	}
}

func TestRenderSanitizesControlChars(t *testing.T) {
	in := Input{
		Columns: 80,
		Tasks: []Task{{
			Name:       "evil\nname\r2",
			Status:     "running",
			Model:      "claude-haiku-4-5",
			TokenCount: 1000,
		}},
	}
	got := Render(in)
	if strings.ContainsAny(got, "\r") {
		t.Errorf("carriage return leaked into output: %q", got)
	}
	// One task must stay one line: injected \n must not create a second row.
	if strings.Count(got, "\n") != 0 {
		t.Errorf("injected newline created extra line(s): %q", got)
	}
}
