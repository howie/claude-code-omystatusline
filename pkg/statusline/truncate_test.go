package statusline

import (
	"strings"
	"testing"
)

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"plain ascii", "hello", 5},
		{"ansi reset", "\033[0mhello", 5},
		{"ansi color", "\033[38;2;195;158;83mABC\033[0m", 3},
		{"emoji folder", "📂", 2},
		{"emoji lightning", "⚡", 1}, // U+26A1, Misc Symbols, 1-col in most terminals
		{"emoji gold", "💛", 2},
		{"emoji money", "💰", 2},
		{"block char filled", "█", 1},
		{"block char empty", "░", 1},
		{"mixed ansi and emoji", "\033[0m📂 project\033[0m", 10},
		{"gradient progress bar",
			"\033[38;2;76;175;80m█\033[0m\033[38;2;64;64;64m░░░░░░░░░\033[0m",
			10},
		{"model display", "\033[0m[💛 Opus 4.6]", 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VisibleWidth(tt.input)
			if got != tt.want {
				t.Errorf("VisibleWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateLineNoTruncation(t *testing.T) {
	segs := []Segment{
		{Content: "hello ", Priority: 1},
		{Content: "world", Priority: 2},
	}
	got := TruncateLine(segs, 100)
	if got != "hello world" {
		t.Errorf("expected no truncation, got %q", got)
	}
}

func TestTruncateLineRemovesLowPriority(t *testing.T) {
	segs := []Segment{
		{Content: "AAAA", Priority: 1}, // 4 chars
		{Content: "BBBB", Priority: 5}, // 4 chars, lower priority
		{Content: "CCCC", Priority: 9}, // 4 chars, lowest priority
	}
	// maxWidth = 9: total = 12, need to remove ≥3 chars + 1 for ellipsis
	// Remove Priority 9 (CCCC = 4), total = 8 + ellipsis(1) = 9 → fits
	got := TruncateLine(segs, 9)
	if !strings.HasPrefix(got, "AAAA") {
		t.Errorf("should keep priority 1 segment, got %q", got)
	}
	if strings.Contains(got, "CCCC") {
		t.Errorf("should remove lowest priority segment, got %q", got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("should end with ellipsis, got %q", got)
	}
}

func TestTruncateLineKeepsPriority1(t *testing.T) {
	segs := []Segment{
		{Content: strings.Repeat("A", 200), Priority: 1},
	}
	// Even when exceeding width, priority 1 is never removed
	got := TruncateLine(segs, 10)
	if !strings.Contains(got, strings.Repeat("A", 200)) {
		t.Errorf("priority 1 segment should never be removed")
	}
}

func TestTruncateLineEmptySegmentsFiltered(t *testing.T) {
	segs := []Segment{
		{Content: "hello", Priority: 1},
		{Content: "", Priority: 2},
		{Content: "world", Priority: 3},
	}
	got := TruncateLine(segs, 100)
	if got != "helloworld" {
		t.Errorf("empty segments should be filtered, got %q", got)
	}
}

func TestTruncateLineWithAnsiColors(t *testing.T) {
	// Simulate real status line segments with ANSI codes
	model := "\033[0m[💛 Opus 4.6] 📂 project\033[0m" // visible: ~22 chars
	git := " ⚡ main"                                // visible: 8
	session := " | 2h10m [3 sessions]"              // visible: 21
	cost := " | 💰 $0.13"                            // visible: 12

	segs := []Segment{
		{Content: model, Priority: 1},
		{Content: git, Priority: 3},
		{Content: session, Priority: 5},
		{Content: cost, Priority: 6},
	}

	// Full width: should include everything
	full := TruncateLine(segs, 200)
	if !strings.Contains(full, "2h10m") {
		t.Errorf("full width should include session, got %q", full)
	}

	// Narrow width: should drop cost first, then session
	narrow := TruncateLine(segs, 35)
	if strings.Contains(narrow, "$0.13") {
		t.Errorf("narrow width should drop cost segment, got %q", narrow)
	}
}
