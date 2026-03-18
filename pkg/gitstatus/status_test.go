package gitstatus

import (
	"strings"
	"testing"
)

func TestFormatEmpty(t *testing.T) {
	info := &GitStatusInfo{}
	result := Format(info)
	if result != "" {
		t.Fatalf("expected empty string for clean repo, got %q", result)
	}
}

func TestFormatDirty(t *testing.T) {
	info := &GitStatusInfo{IsDirty: true, Modified: 3, Added: 1}
	result := Format(info)
	if !strings.Contains(result, "*") {
		t.Fatalf("expected dirty indicator, got %q", result)
	}
	if !strings.Contains(result, "!3") {
		t.Fatalf("expected modified count, got %q", result)
	}
	if !strings.Contains(result, "+1") {
		t.Fatalf("expected added count, got %q", result)
	}
}

func TestFormatAheadBehind(t *testing.T) {
	info := &GitStatusInfo{Ahead: 2, Behind: 1}
	result := Format(info)
	if !strings.Contains(result, "↑2") {
		t.Fatalf("expected ahead indicator, got %q", result)
	}
	if !strings.Contains(result, "↓1") {
		t.Fatalf("expected behind indicator, got %q", result)
	}
}

func TestFormatNil(t *testing.T) {
	if Format(nil) != "" {
		t.Fatal("expected empty for nil")
	}
}

func TestFormatFullStatus(t *testing.T) {
	info := &GitStatusInfo{
		IsDirty:   true,
		Ahead:     1,
		Behind:    0,
		Modified:  2,
		Added:     1,
		Deleted:   1,
		Untracked: 3,
	}
	result := Format(info)
	if !strings.Contains(result, "*") {
		t.Fatalf("missing dirty, got %q", result)
	}
	if !strings.Contains(result, "↑1") {
		t.Fatalf("missing ahead, got %q", result)
	}
	if !strings.Contains(result, "!2") {
		t.Fatalf("missing modified, got %q", result)
	}
	if !strings.Contains(result, "✘1") {
		t.Fatalf("missing deleted, got %q", result)
	}
	if !strings.Contains(result, "?3") {
		t.Fatalf("missing untracked, got %q", result)
	}
}
