package transcript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.log")

	lines := []string{
		`{"type":"user","sessionId":"s1","message":{"role":"user","content":"hello"}}`,
		`{"type":"assistant","sessionId":"s1","message":{"role":"assistant","content":"hi"}}`,
		`not json line`,
		`{"type":"user","sessionId":"s2","message":{"role":"user","content":"bye"}}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ReadTail(path, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}

	// 第一行是第 2 行（index 1）— "assistant" line
	if result[0].Parsed == nil {
		t.Fatal("expected second line to be valid JSON")
	}

	// 第二行不是有效 JSON
	if result[1].Parsed != nil {
		t.Fatal("expected 'not json line' to have nil Parsed")
	}
}

func TestReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.log")
	content := `{"a":1}
{"b":2}
{"c":3}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := ReadAll(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("ReadAll: expected 3 lines, got %d", len(result))
	}
}

func TestReadTailEmptyPath(t *testing.T) {
	result, err := ReadTail("", 10)
	if err != nil {
		t.Fatal("expected no error for empty path")
	}
	if result != nil {
		t.Fatal("expected nil result for empty path")
	}
}

func TestFilterBySession(t *testing.T) {
	lines := []Line{
		{Parsed: map[string]interface{}{"sessionId": "s1", "isSidechain": false}},
		{Parsed: map[string]interface{}{"sessionId": "s2", "isSidechain": false}},
		{Parsed: map[string]interface{}{"sessionId": "s1", "isSidechain": true}},
		{Parsed: map[string]interface{}{"sessionId": "s1"}},
		{Parsed: nil},
	}

	result := FilterBySession(lines, "s1")
	if len(result) != 2 {
		t.Fatalf("expected 2 lines for session s1, got %d", len(result))
	}
}
