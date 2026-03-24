package terminal

import (
	"testing"
)

func TestDetectASCIForced(t *testing.T) {
	t.Setenv("CLAUDE_STATUSLINE_ASCII", "1")

	if mode := Detect(); mode != ModeASCII {
		t.Fatalf("expected ModeASCII when CLAUDE_STATUSLINE_ASCII=1, got %d", mode)
	}
}

func TestDetectTrueColor(t *testing.T) {
	t.Setenv("CLAUDE_STATUSLINE_ASCII", "")
	t.Setenv("COLORTERM", "truecolor")

	if mode := Detect(); mode != ModeTrueColor {
		t.Fatalf("expected ModeTrueColor for COLORTERM=truecolor, got %d", mode)
	}
}

func TestDetect24bit(t *testing.T) {
	t.Setenv("CLAUDE_STATUSLINE_ASCII", "")
	t.Setenv("COLORTERM", "24bit")

	if mode := Detect(); mode != ModeTrueColor {
		t.Fatalf("expected ModeTrueColor for COLORTERM=24bit, got %d", mode)
	}
}

func TestDetect256Color(t *testing.T) {
	t.Setenv("CLAUDE_STATUSLINE_ASCII", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")

	if mode := Detect(); mode != Mode256Color {
		t.Fatalf("expected Mode256Color for TERM=xterm-256color, got %d", mode)
	}
}

func TestDetectDumbTerminal(t *testing.T) {
	t.Setenv("CLAUDE_STATUSLINE_ASCII", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "dumb")

	if mode := Detect(); mode != ModeASCII {
		t.Fatalf("expected ModeASCII for TERM=dumb, got %d", mode)
	}
}
