package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatModel(t *testing.T) {
	colored := formatModel("Claude 3 Opus")
	if !strings.Contains(colored, ColorGold) {
		t.Fatalf("formatModel should wrap Opus models with gold color, got %q", colored)
	}
	if !strings.HasSuffix(colored, ColorReset) {
		t.Fatalf("formatModel should reset color at the end, got %q", colored)
	}

	plain := formatModel("Custom Model")
	if plain != "Custom Model" {
		t.Fatalf("formatModel should return unchanged name for unknown models, got %q", plain)
	}
}

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

func TestGetContextColor(t *testing.T) {
	if color := getContextColor(40); color != ColorCtxGreen {
		t.Fatalf("expected ColorCtxGreen for 40%%, got %q", color)
	}
	if color := getContextColor(70); color != ColorCtxGold {
		t.Fatalf("expected ColorCtxGold for 70%%, got %q", color)
	}
	if color := getContextColor(90); color != ColorCtxRed {
		t.Fatalf("expected ColorCtxRed for 90%%, got %q", color)
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

func TestCalculateContextUsage(t *testing.T) {
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

	total := calculateContextUsage(path)
	if total != 175 {
		t.Fatalf("expected total usage 175, got %d", total)
	}
}
