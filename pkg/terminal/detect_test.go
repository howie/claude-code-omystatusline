package terminal

import (
	"os"
	"testing"
)

func TestDetectASCIForced(t *testing.T) {
	os.Setenv("CLAUDE_STATUSLINE_ASCII", "1")
	defer os.Unsetenv("CLAUDE_STATUSLINE_ASCII")

	if mode := Detect(); mode != ModeASCII {
		t.Fatalf("expected ModeASCII when CLAUDE_STATUSLINE_ASCII=1, got %d", mode)
	}
}

func TestDetectTrueColor(t *testing.T) {
	os.Unsetenv("CLAUDE_STATUSLINE_ASCII")
	os.Setenv("COLORTERM", "truecolor")
	defer os.Unsetenv("COLORTERM")

	if mode := Detect(); mode != ModeTrueColor {
		t.Fatalf("expected ModeTrueColor for COLORTERM=truecolor, got %d", mode)
	}
}

func TestDetect24bit(t *testing.T) {
	os.Unsetenv("CLAUDE_STATUSLINE_ASCII")
	os.Setenv("COLORTERM", "24bit")
	defer os.Unsetenv("COLORTERM")

	if mode := Detect(); mode != ModeTrueColor {
		t.Fatalf("expected ModeTrueColor for COLORTERM=24bit, got %d", mode)
	}
}

func TestDetect256Color(t *testing.T) {
	os.Unsetenv("CLAUDE_STATUSLINE_ASCII")
	os.Unsetenv("COLORTERM")
	os.Setenv("TERM", "xterm-256color")
	defer os.Unsetenv("TERM")

	if mode := Detect(); mode != Mode256Color {
		t.Fatalf("expected Mode256Color for TERM=xterm-256color, got %d", mode)
	}
}

func TestDetectDumbTerminal(t *testing.T) {
	os.Unsetenv("CLAUDE_STATUSLINE_ASCII")
	os.Unsetenv("COLORTERM")
	os.Setenv("TERM", "dumb")
	defer os.Unsetenv("TERM")

	if mode := Detect(); mode != ModeASCII {
		t.Fatalf("expected ModeASCII for TERM=dumb, got %d", mode)
	}
}
