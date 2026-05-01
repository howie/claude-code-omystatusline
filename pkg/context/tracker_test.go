package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/transcript"
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

func TestAnalyzeMaxTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.log")

	// 寫入一個有 100000 tokens 的 transcript
	line := `{"message":{"usage":{"input_tokens":100000}},"isSidechain":false}`
	if err := os.WriteFile(path, []byte(line), 0644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	tests := []struct {
		name      string
		maxTokens int
		wantPct   string // 預期百分比字串
	}{
		{"default (200k)", 0, "50%"},
		{"200k explicit", 200000, "50%"},
		{"1M extended", 1000000, "10%"},
		{"negative fallback", -1, "50%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Analyze(path, tt.maxTokens)
			if !strings.Contains(result, tt.wantPct) {
				t.Fatalf("Analyze with maxTokens=%d: expected %q in result, got %q", tt.maxTokens, tt.wantPct, result)
			}
		})
	}
}

func TestAnalyzeEmptyPath(t *testing.T) {
	result := Analyze("", 0)
	if !strings.Contains(result, "0%") {
		t.Fatalf("Analyze with empty path should show 0%%, got %q", result)
	}
}

func TestFormatContextParts(t *testing.T) {
	tests := []struct {
		name          string
		contextLength int
		maxTokens     int
		wantBarPrefix string // bar 應以 " | " 開頭
		wantInfo      string // info 應包含此字串
	}{
		{"zero tokens", 0, 200000, " | ", "0%"},
		{"50 percent", 100000, 200000, " | ", "50%"},
		{"100 percent", 200000, 200000, " | ", "100%"},
		{"token count", 148000, 200000, " | ", "148k"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar, info := FormatContextParts(tt.contextLength, tt.maxTokens)
			if !strings.HasPrefix(bar, tt.wantBarPrefix) {
				t.Errorf("bar should start with %q, got %q", tt.wantBarPrefix, bar)
			}
			if !strings.Contains(info, tt.wantInfo) {
				t.Errorf("info should contain %q, got %q", tt.wantInfo, info)
			}
			// bar 不應包含百分比，info 不應包含進度條字元
			if strings.Contains(bar, "%") {
				t.Errorf("bar should not contain percentage, got %q", bar)
			}
		})
	}
}

func TestAnalyzeDetailedFromLines(t *testing.T) {
	lines := []transcript.Line{
		{Raw: `{"message":{"usage":{"input_tokens":148027}},"isSidechain":false}`, Parsed: map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{"input_tokens": float64(148027)},
			},
			"isSidechain": false,
		}},
	}

	data := AnalyzeDetailedFromLines(lines, 200000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if data.Tokens != 148027 {
		t.Errorf("expected Tokens=148027, got %d", data.Tokens)
	}
	if data.Percentage != 74 {
		t.Errorf("expected Percentage=74, got %d", data.Percentage)
	}
	if data.Bar == "" {
		t.Error("Bar should not be empty")
	}
	if data.Info == "" {
		t.Error("Info should not be empty")
	}
	// Formatted should be Bar + Info
	if data.Formatted != data.Bar+data.Info {
		t.Errorf("Formatted should equal Bar+Info, got Formatted=%q, Bar+Info=%q",
			data.Formatted, data.Bar+data.Info)
	}
}

func TestAnalyzeDetailedFromLines_MetadataOnly(t *testing.T) {
	// Simulates a local-agent-mode transcript that only contains metadata events
	// (custom-title, agent-name, pr-link) with no actual conversation entries.
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "custom-title", "customTitle": "figma-flutter-rule-setup", "sessionId": "abc"}},
		{Parsed: map[string]interface{}{"type": "agent-name", "agentName": "figma-flutter-rule-setup", "sessionId": "abc"}},
		{Parsed: map[string]interface{}{"type": "pr-link", "sessionId": "abc", "prNumber": float64(213)}},
	}

	data := AnalyzeDetailedFromLines(lines, 1_000_000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if !data.NoUsageData {
		t.Error("expected NoUsageData=true for metadata-only transcript")
	}
	if data.Percentage != 0 {
		t.Errorf("expected Percentage=0, got %d", data.Percentage)
	}
	if data.Tokens != 0 {
		t.Errorf("expected Tokens=0, got %d", data.Tokens)
	}
	if !strings.Contains(data.Info, "📡") {
		t.Errorf("expected Info to contain 📡 for NoUsageData, got %q", data.Info)
	}
	if data.Formatted != data.Bar+data.Info {
		t.Errorf("Formatted should equal Bar+Info")
	}
}

func TestAnalyzeDetailedFromLines_NoUsageData_FalseForNewSession(t *testing.T) {
	// nil lines (no transcript yet) should NOT set NoUsageData
	data := AnalyzeDetailedFromLines(nil, 200000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if data.NoUsageData {
		t.Error("expected NoUsageData=false for nil lines")
	}

	// empty slice should also not set NoUsageData
	data2 := AnalyzeDetailedFromLines([]transcript.Line{}, 200000)
	if data2.NoUsageData {
		t.Error("expected NoUsageData=false for empty lines")
	}

	// a user message (no usage yet) should not set NoUsageData — it has a message field
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{
			"message":     map[string]interface{}{"role": "user", "content": "hello"},
			"isSidechain": false,
		}},
	}
	data3 := AnalyzeDetailedFromLines(lines, 200000)
	if data3.NoUsageData {
		t.Error("expected NoUsageData=false when line has message field but no usage yet")
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
