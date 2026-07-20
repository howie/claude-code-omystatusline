package statusline

import "strings"

// TruncateToWidth hard-clamps s to at most maxWidth visible cells. Unlike TruncateLine
// (which protects high-priority segments and can exceed the requested width), this is a
// guaranteed upper bound: VisibleWidth(result) <= maxWidth always.
//
// It is ANSI-aware — CSI color sequences and OSC 8 hyperlinks contribute 0 width and are
// copied through — and never splits a wide (width-2) rune. If the cut lands inside colored
// text, a trailing ColorReset is appended so color does not bleed past the row. maxWidth
// <= 0 returns "".
//
// Used by the subagent status line to bound each row to the payload's `columns`, which
// TruncateLine cannot guarantee.
func TruncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if VisibleWidth(s) <= maxWidth {
		return s
	}

	var b strings.Builder
	width := 0
	sawEscape := false
	inEscape := false // CSI (\033[...)
	inOSC := false    // OSC (\033]...)
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inOSC {
			b.WriteRune(r)
			if r == '\a' {
				inOSC = false
			} else if r == '\033' && i+1 < len(runes) && runes[i+1] == '\\' {
				b.WriteRune(runes[i+1])
				i++
				inOSC = false
			}
			continue
		}
		if inEscape {
			b.WriteRune(r)
			if r >= 0x40 && r <= 0x7E {
				inEscape = false
			}
			continue
		}
		if r == '\033' && i+1 < len(runes) && runes[i+1] == '[' {
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i++
			inEscape = true
			sawEscape = true
			continue
		}
		if r == '\033' && i+1 < len(runes) && runes[i+1] == ']' {
			b.WriteRune(r)
			b.WriteRune(runes[i+1])
			i++
			inOSC = true
			sawEscape = true
			continue
		}
		w := runeWidth(r)
		if width+w > maxWidth {
			break
		}
		b.WriteRune(r)
		width += w
	}

	if sawEscape {
		b.WriteString(ColorReset)
	}
	return b.String()
}
