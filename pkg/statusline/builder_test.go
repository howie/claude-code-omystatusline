package statusline

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatModel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantColor string
		wantPlain bool
	}{
		{"Claude 3 Opus", "Claude 3 Opus", ColorGold, false},
		{"Opus 4.6", "Opus 4.6", ColorGold, false},
		{"Claude Opus 4.6", "Claude Opus 4.6", ColorGold, false},
		{"Sonnet 4.6", "Sonnet 4.6", ColorCyan, false},
		{"Haiku 4.5", "Haiku 4.5", ColorPink, false},
		{"Custom Model", "Custom Model", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatModel(tt.input)
			if tt.wantPlain {
				if result != tt.input {
					t.Fatalf("FormatModel(%q) should return unchanged, got %q", tt.input, result)
				}
				return
			}
			if !strings.Contains(result, tt.wantColor) {
				t.Fatalf("FormatModel(%q) should contain color %q, got %q", tt.input, tt.wantColor, result)
			}
			if !strings.HasSuffix(result, ColorReset) {
				t.Fatalf("FormatModel(%q) should end with ColorReset, got %q", tt.input, result)
			}
		})
	}
}

func TestIsSystemMessage(t *testing.T) {
	systemMessages := []string{
		`{"key":"value"}`,
		`[1,2,3]`,
		"<local-command-stdout>foo</local-command-stdout>",
		"Caveat: something happened",
	}
	for _, msg := range systemMessages {
		if !isSystemMessage(msg) {
			t.Fatalf("expected %q to be classified as system message", msg)
		}
	}

	if isSystemMessage("normal user content") {
		t.Fatal("expected regular content to be treated as user message")
	}
}

func TestFormatUserMessage(t *testing.T) {
	longLine := strings.Repeat("A", 90)
	message := strings.Join([]string{
		longLine,
		"Second line of text",
		"Third line here",
		"Hidden fourth line",
	}, "\n")

	formatted := formatUserMessage(message)
	lines := strings.Split(strings.TrimSuffix(formatted, "\n"), "\n")

	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (3 content + ellipsis), got %d", len(lines))
	}

	if !strings.Contains(lines[0], strings.Repeat("A", 77)+"...") {
		t.Fatalf("first line should be truncated with ellipsis, got %q", lines[0])
	}

	for i := 0; i < 3; i++ {
		if !strings.HasPrefix(lines[i], fmt.Sprintf("%s｜%s", ColorReset, ColorGreen)) {
			t.Fatalf("content line %d missing expected color prefix: %q", i, lines[i])
		}
		if !strings.HasSuffix(lines[i], ColorReset) {
			t.Fatalf("content line %d missing color reset suffix: %q", i, lines[i])
		}
	}

	if !strings.Contains(lines[3], "還有 1 行") {
		t.Fatalf("ellipsis line should mention remaining content, got %q", lines[3])
	}
}

func TestFormatLinesChanged(t *testing.T) {
	tests := []struct {
		name    string
		added   int
		removed int
		want    string
	}{
		{"both zero", 0, 0, ""},
		{"only added", 50, 0, ""},
		{"only removed", 0, 10, ""},
		{"both non-zero", 50, 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLinesChanged(tt.added, tt.removed)
			if tt.added == 0 && tt.removed == 0 {
				if result != "" {
					t.Fatalf("expected empty string for zero values, got %q", result)
				}
				return
			}
			if tt.added > 0 && !strings.Contains(result, fmt.Sprintf("+%d", tt.added)) {
				t.Fatalf("expected +%d in result, got %q", tt.added, result)
			}
			if tt.removed > 0 && !strings.Contains(result, fmt.Sprintf("-%d", tt.removed)) {
				t.Fatalf("expected -%d in result, got %q", tt.removed, result)
			}
			if tt.added > 0 && !strings.Contains(result, ColorGreen) {
				t.Fatalf("expected green color for additions, got %q", result)
			}
			if tt.removed > 0 && !strings.Contains(result, ColorRed) {
				t.Fatalf("expected red color for removals, got %q", result)
			}
		})
	}
}

func TestFormatCostColored(t *testing.T) {
	tests := []struct {
		name      string
		cost      float64
		wantEmpty bool
		wantColor string
	}{
		{"zero", 0, true, ""},
		{"negative", -1, true, ""},
		{"low cost", 2.50, false, ColorDim},
		{"medium cost", 5.00, false, ColorYellow},
		{"high cost", 10.00, false, ColorRed},
		{"very high", 25.50, false, ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCostColored(tt.cost)
			if tt.wantEmpty {
				if result != "" {
					t.Fatalf("expected empty for cost %.2f, got %q", tt.cost, result)
				}
				return
			}
			if !strings.Contains(result, tt.wantColor) {
				t.Fatalf("expected color %q for cost %.2f, got %q", tt.wantColor, tt.cost, result)
			}
			if !strings.Contains(result, fmt.Sprintf("$%.2f", tt.cost)) {
				t.Fatalf("expected formatted cost in result, got %q", result)
			}
		})
	}
}
