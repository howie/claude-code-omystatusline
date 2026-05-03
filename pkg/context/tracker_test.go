package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
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
	if data.NoUsageData {
		t.Error("expected NoUsageData=false for a transcript with message.usage")
	}
	if !data.HasData() {
		t.Error("expected HasData()=true when tokens > 0 and NoUsageData=false")
	}
}

// TestAnalyzeDetailedFromLines_MetadataOnly verifies that a transcript containing only
// administrative metadata events (custom-title, agent-name, pr-link) — as seen in
// local-agent-mode sessions — sets NoUsageData=true and renders 📡 instead of "0% --".
func TestAnalyzeDetailedFromLines_MetadataOnly(t *testing.T) {
	orig := RenderMode
	RenderMode = terminal.ModeTrueColor
	defer func() { RenderMode = orig }()

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
	if data.HasData() {
		t.Error("expected HasData()=false when NoUsageData=true")
	}
	if !strings.Contains(data.Info, "📡") {
		t.Errorf("expected Info to contain 📡 for NoUsageData, got %q", data.Info)
	}
	if !strings.HasPrefix(data.Bar, " | ") {
		t.Errorf("expected Bar to start with \" | \", got %q", data.Bar)
	}
}

// TestIsMetadataOnlyTranscript_NullMessage verifies the documented design decision:
// a line with "message": null is treated as having a "message" field (not metadata-only).
// This guards against refactoring the map-key check into a nil-check, which would break
// the invariant silently.
func TestIsMetadataOnlyTranscript_NullMessage(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "summary", "message": nil}},
	}
	data := AnalyzeDetailedFromLines(lines, 200000)
	if data.NoUsageData {
		t.Error("line with message=nil must not be treated as metadata-only: " +
			"Go map key presence check returns true even when value is nil")
	}
}

// TestAnalyzeDetailedFromLines_MetadataOnly_ASCII verifies the ASCII fallback renders "[remote]".
func TestAnalyzeDetailedFromLines_MetadataOnly_ASCII(t *testing.T) {
	orig := RenderMode
	RenderMode = terminal.ModeASCII
	defer func() { RenderMode = orig }()

	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "custom-title", "customTitle": "foo"}},
		{Parsed: map[string]interface{}{"type": "pr-link", "prNumber": float64(1)}},
	}

	data := AnalyzeDetailedFromLines(lines, 1_000_000)
	if !data.NoUsageData {
		t.Error("expected NoUsageData=true for metadata-only transcript in ASCII mode")
	}
	if data.Info != " [remote]" {
		t.Errorf("expected Info=[remote] in ASCII mode, got %q", data.Info)
	}
}

// TestAnalyzeDetailedFromLines_MixedLines verifies that a transcript starting with metadata
// events but containing at least one message line does NOT set NoUsageData.
func TestAnalyzeDetailedFromLines_MixedLines(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"type": "custom-title", "customTitle": "foo"}},
		{Parsed: map[string]interface{}{
			"message":     map[string]interface{}{"role": "user", "content": "hello"},
			"isSidechain": false,
		}},
	}
	data := AnalyzeDetailedFromLines(lines, 200000)
	if data.NoUsageData {
		t.Error("metadata lines followed by a message line must not set NoUsageData")
	}
}

// TestAnalyzeDetailedFromLines_AllNilParsed verifies that lines with all-nil Parsed fields
// (malformed JSON) do NOT set NoUsageData — the transcript is unreadable, not metadata-only.
func TestAnalyzeDetailedFromLines_AllNilParsed(t *testing.T) {
	lines := []transcript.Line{
		{Raw: "not json", Parsed: nil},
		{Raw: "also not json", Parsed: nil},
	}
	data := AnalyzeDetailedFromLines(lines, 200000)
	if data.NoUsageData {
		t.Error("all-nil Parsed lines must not set NoUsageData")
	}
}

// TestAnalyzeDetailedFromLines_NoUsageData_FalseForNewSession verifies that nil/empty lines
// and a user-message-only transcript (no assistant reply yet) do NOT trigger NoUsageData.
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

	// a user message with no usage yet should not set NoUsageData — it has a message field,
	// so isMetadataOnlyTranscript returns false
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

// TestAnalyzeDetailed_MetadataOnly verifies that the file-path-based AnalyzeDetailed
// also detects metadata-only transcripts and sets NoUsageData=true.
func TestAnalyzeDetailed_MetadataOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.jsonl")
	content := `{"type":"custom-title","customTitle":"test"}
{"type":"agent-name","agentName":"test"}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	data := AnalyzeDetailed(path, 1_000_000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if !data.NoUsageData {
		t.Error("AnalyzeDetailed should set NoUsageData=true for metadata-only file")
	}
	if data.HasData() {
		t.Error("expected HasData()=false for metadata-only file")
	}
}

// TestAnalyzeDetailed_UnreadablePath verifies that an unreadable path does NOT set NoUsageData.
func TestAnalyzeDetailed_UnreadablePath(t *testing.T) {
	data := AnalyzeDetailed("/nonexistent/path.jsonl", 200000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if data.NoUsageData {
		t.Error("AnalyzeDetailed should not set NoUsageData for unreadable file")
	}
}

// TestAnalyzeDetailed_EmptyPath verifies that an empty path returns a usable ContextData
// with NoUsageData=false and HasData()=false (treated as a new session, not metadata-only).
func TestAnalyzeDetailed_EmptyPath(t *testing.T) {
	data := AnalyzeDetailed("", 200000)
	if data == nil {
		t.Fatal("expected non-nil ContextData")
	}
	if data.NoUsageData {
		t.Error("AnalyzeDetailed with empty path must not set NoUsageData")
	}
	if data.HasData() {
		t.Error("AnalyzeDetailed with empty path must return HasData()=false")
	}
}

// TestHasData verifies the HasData() method for all three semantic states.
func TestHasData(t *testing.T) {
	// Real usage data
	realData := &ContextData{Tokens: 50000, Percentage: 25, NoUsageData: false}
	if !realData.HasData() {
		t.Error("expected HasData()=true when Tokens>0 and NoUsageData=false")
	}

	// NoUsageData (metadata-only session)
	noUsageData := &ContextData{Tokens: 0, Percentage: 0, NoUsageData: true}
	if noUsageData.HasData() {
		t.Error("expected HasData()=false when NoUsageData=true")
	}

	// New session (genuinely 0 tokens, but not metadata-only)
	newSession := &ContextData{Tokens: 0, Percentage: 0, NoUsageData: false}
	if newSession.HasData() {
		t.Error("expected HasData()=false when Tokens=0 and NoUsageData=false")
	}
}

func TestCalculateUsage(t *testing.T) {
	lines := []transcript.Line{
		{Parsed: map[string]interface{}{"message": map[string]interface{}{"usage": map[string]interface{}{"input_tokens": float64(10)}}, "isSidechain": true}},
		{Raw: "not json", Parsed: nil},
		{Parsed: map[string]interface{}{"message": map[string]interface{}{"usage": map[string]interface{}{"input_tokens": float64(100), "cache_read_input_tokens": float64(50), "cache_creation_input_tokens": float64(25)}}, "isSidechain": false}},
	}
	if total := calculateUsageFromLines(lines); total != 175 {
		t.Fatalf("expected total usage 175, got %d", total)
	}
}

func makeUsageLine(modelID string, inputTokens float64) transcript.Line {
	return transcript.Line{Parsed: map[string]interface{}{
		"message": map[string]interface{}{
			"model": modelID,
			"usage": map[string]interface{}{"input_tokens": inputTokens},
		},
		"isSidechain": false,
	}}
}

// TestInferModelFromLines 驗證從 transcript 最後一筆有 usage 的行讀出 model ID。
func TestInferModelFromLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []transcript.Line
		want  string
	}{
		{
			name:  "single sonnet line",
			lines: []transcript.Line{makeUsageLine("claude-sonnet-4-6", 100000)},
			want:  "claude-sonnet-4-6",
		},
		{
			name: "mixed model — returns last (sonnet after opus)",
			lines: []transcript.Line{
				makeUsageLine("claude-opus-4-7", 80000),
				makeUsageLine("claude-sonnet-4-6", 150000),
			},
			want: "claude-sonnet-4-6",
		},
		{
			name: "mixed model — returns last (opus after sonnet)",
			lines: []transcript.Line{
				makeUsageLine("claude-sonnet-4-6", 50000),
				makeUsageLine("claude-opus-4-7", 200000),
			},
			want: "claude-opus-4-7",
		},
		{
			name: "isSidechain line skipped",
			lines: []transcript.Line{
				{Parsed: map[string]interface{}{
					"message":     map[string]interface{}{"model": "claude-opus-4-7", "usage": map[string]interface{}{"input_tokens": float64(1)}},
					"isSidechain": true,
				}},
				makeUsageLine("claude-sonnet-4-6", 100000),
			},
			want: "claude-sonnet-4-6",
		},
		{
			name: "usage line without model field — skipped, fallback to earlier line",
			lines: []transcript.Line{
				makeUsageLine("claude-opus-4-7", 80000),
				{Parsed: map[string]interface{}{
					"message":     map[string]interface{}{"usage": map[string]interface{}{"input_tokens": float64(90000)}},
					"isSidechain": false,
				}},
			},
			want: "claude-opus-4-7",
		},
		{
			name: "user message line (no usage) skipped",
			lines: []transcript.Line{
				makeUsageLine("claude-sonnet-4-6", 100000),
				{Parsed: map[string]interface{}{
					"message":     map[string]interface{}{"role": "user", "content": "hello"},
					"isSidechain": false,
				}},
			},
			want: "claude-sonnet-4-6",
		},
		{
			name:  "empty lines",
			lines: []transcript.Line{},
			want:  "",
		},
		{
			name:  "nil lines",
			lines: nil,
			want:  "",
		},
		{
			name:  "all nil Parsed",
			lines: []transcript.Line{{Raw: "bad json", Parsed: nil}},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferModelFromLines(tt.lines)
			if got != tt.want {
				t.Errorf("InferModelFromLines() = %q, want %q", got, tt.want)
			}
		})
	}
}
