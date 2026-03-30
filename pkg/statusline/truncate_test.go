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

func TestWrapLineFitsInOneLine(t *testing.T) {
	segs := []Segment{
		{Content: "hello ", Priority: 1},
		{Content: "world", Priority: 2},
	}
	got := WrapLine(segs, 100)
	if strings.Contains(got, "\n") {
		t.Errorf("should not wrap when fits in one line, got %q", got)
	}
	if got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestWrapLineWrapsToSecondLine(t *testing.T) {
	// AAAAAAA(7) + " | BBBB"(7) = 14 > 12，所以 " | BBBB" 換到第二行
	segs := []Segment{
		{Content: "AAAAAAA", Priority: 1}, // 7
		{Content: " | BBBB", Priority: 3}, // 7
		{Content: " | CCCC", Priority: 5}, // 7
	}
	got := WrapLine(segs, 12)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "AAAAAAA") {
		t.Errorf("line1 should contain AAAAAAA, got %q", lines[0])
	}
	if strings.Contains(lines[0], "BBBB") {
		t.Errorf("line1 should NOT contain BBBB (should be on line2), got %q", lines[0])
	}
	if VisibleWidth(lines[0]) > 12 {
		t.Errorf("line1 visible width %d exceeds maxWidth 12: %q", VisibleWidth(lines[0]), lines[0])
	}
	if !strings.HasPrefix(lines[1], " ") {
		t.Errorf("line2 should start with single-space prefix, got %q", lines[1])
	}
	if !strings.Contains(lines[1], "BBBB") {
		t.Errorf("line2 should contain BBBB, got %q", lines[1])
	}
}

func TestWrapLineStripsLeadingDivider(t *testing.T) {
	segs := []Segment{
		{Content: "AAAAAAA", Priority: 1}, // 7
		{Content: " | BBBB", Priority: 3}, // 7, starts with " | "
	}
	got := WrapLine(segs, 10)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}
	// " | " 應被去掉，改為前綴 " "
	if strings.HasPrefix(lines[1], " | ") {
		t.Errorf("line2 should not start with \" | \", got %q", lines[1])
	}
	if !strings.HasPrefix(lines[1], " ") {
		t.Errorf("line2 should start with single-space prefix, got %q", lines[1])
	}
	if !strings.Contains(lines[1], "BBBB") {
		t.Errorf("line2 should still contain BBBB, got %q", lines[1])
	}
}

func TestWrapLineFallbackTruncate(t *testing.T) {
	// maxWidth=10：
	// line1: "AAAA"(4)，" | BBBBB"(8) 4+8=12>10 → 移到 line2
	// line2: [" | BBBBB"(P3), " | CCCCC"(P5), " | DDDDD"(P9)]
	// line2 strip 後: "BBBBB"(5) + " | CCCCC"(8) + " | DDDDD"(8) = 21+前綴1 = 22 > 10
	// → 觸發 TruncateLine，Priority 9（DDDDD）被丟棄
	segs := []Segment{
		{Content: "AAAA", Priority: 1},     // 4
		{Content: " | BBBBB", Priority: 3}, // 8
		{Content: " | CCCCC", Priority: 5}, // 8
		{Content: " | DDDDD", Priority: 9}, // 8，最低優先，應被截斷
	}
	got := WrapLine(segs, 10)
	if !strings.Contains(got, "BBBBB") {
		t.Errorf("should contain BBBBB (high priority on line2), got %q", got)
	}
	if strings.Contains(got, "DDDDD") {
		t.Errorf("DDDDD should be truncated (lowest priority), got %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Errorf("should contain ellipsis when line2 is truncated, got %q", got)
	}
}

func TestWrapLineEmptySegments(t *testing.T) {
	segs := []Segment{
		{Content: "hello", Priority: 1},
		{Content: "", Priority: 2},
		{Content: " | world", Priority: 3},
	}
	got := WrapLine(segs, 100)
	if strings.Contains(got, "\n") {
		t.Errorf("empty segments should be filtered, should fit in one line")
	}
}

func TestWrapLineSingleOversizedSegment(t *testing.T) {
	// 單一段落超出 maxWidth，應強制放入不截斷
	segs := []Segment{
		{Content: strings.Repeat("A", 50), Priority: 1},
	}
	got := WrapLine(segs, 10)
	if !strings.Contains(got, strings.Repeat("A", 50)) {
		t.Errorf("oversized single segment should be preserved")
	}
}

func TestWrapLineWithAnsiColors(t *testing.T) {
	// 模擬真實段落：model(21) + " | ⚡ main"(9) = 30 > 25，git 換到第二行
	// git 以 " | " 開頭，觸發 stripLeadingDivider，line2 = " ⚡ main"（單空格前綴）
	model := "\033[0m[💛 Opus 4.6] 📂 proj\033[0m" // visible: 21
	git := " | ⚡ main"                            // visible: 9, starts with " | "
	segs := []Segment{
		{Content: model, Priority: 1},
		{Content: git, Priority: 3},
	}
	got := WrapLine(segs, 25)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines with ANSI segments, got %d: %q", len(lines), got)
	}
	if VisibleWidth(lines[0]) > 25 {
		t.Errorf("line1 visible width %d exceeds maxWidth 25", VisibleWidth(lines[0]))
	}
	// stripLeadingDivider 去掉 " | "，line2Prefix 加 " "，結果為 " ⚡ main"（精確單空格）
	if !strings.HasPrefix(lines[1], " ") {
		t.Errorf("line2 should start with single-space prefix, got %q", lines[1])
	}
	if strings.HasPrefix(lines[1], "  ") {
		t.Errorf("line2 should not start with double space (divider not stripped), got %q", lines[1])
	}
	if !strings.Contains(lines[1], "main") {
		t.Errorf("line2 should contain git branch, got %q", lines[1])
	}
}

func TestWrapLineZeroMaxWidth(t *testing.T) {
	// maxWidth <= 0 應預設為 120
	segs := []Segment{
		{Content: strings.Repeat("A", 50), Priority: 1},
		{Content: strings.Repeat("B", 50), Priority: 2},
	}
	got := WrapLine(segs, 0)
	// 總長 100 ≤ 120，應回傳單行
	if strings.Contains(got, "\n") {
		t.Errorf("with maxWidth=0 (default 120), 100-char content should fit in one line, got %q", got)
	}
}

func TestStripLeadingDivider(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{" | hello", "hello"},
		{" \ue0b1 hello", "hello"},
		{" \ue0b0 hello", "hello"},
		{" │ hello", "hello"},
		{"hello", "hello"},
		{" hello", " hello"}, // 無分隔符，保留
	}
	for _, tt := range tests {
		got := stripLeadingDivider(tt.input)
		if got != tt.want {
			t.Errorf("stripLeadingDivider(%q) = %q, want %q", tt.input, got, tt.want)
		}
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
