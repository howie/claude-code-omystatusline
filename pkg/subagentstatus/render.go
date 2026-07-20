package subagentstatus

import (
	"fmt"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
)

// Render produces the subagentStatusLine output: one line per task, in input order,
// each hard-clamped to in.Columns visible cells. Returns "" (zero bytes) when there
// are no tasks, so an idle agent panel stays truly empty. Rendering mode (ASCII vs
// color) follows context.RenderMode, which the caller sets from terminal.Detect().
func Render(in Input) string {
	if len(in.Tasks) == 0 {
		return ""
	}
	ascii := context.RenderMode == terminal.ModeASCII
	rows := make([]string, 0, len(in.Tasks))
	for i := range in.Tasks {
		rows = append(rows, renderRow(in.Tasks[i], in.Columns, ascii))
	}
	return strings.Join(rows, "\n")
}

// renderRow builds one clamped row for a single task.
func renderRow(t Task, columns int, ascii bool) string {
	body := statusGlyph(t.Status, ascii) + " " + taskName(t)

	if denom, ok := resolveWindow(t.Model, t.ContextWindowSize); ok {
		// tokenCount is current context occupancy (verified via spike), a valid
		// numerator. BuildFromTokens clamps the percentage to 0-100, so a tokenCount
		// above the window shows 100% while keeping the raw count.
		cd := context.BuildFromTokens(t.TokenCount, denom)
		body += cd.Bar + cd.Info
	} else {
		// No trusted denominator: token-only, never a guessed percentage. Do NOT call
		// BuildFromTokens(_, 0) here — it would fall back to 200K and manufacture a %.
		body += " (" + formatTokens(t.TokenCount) + ")"
	}

	// columns<=0 means Claude Code gave no width budget: render the full row rather
	// than emitting nothing (TruncateToWidth(_, 0) == "").
	if columns <= 0 {
		return body
	}
	return statusline.TruncateToWidth(body, columns)
}

// taskName picks the best display label: name -> label -> type -> shortened id -> "task".
// Real payloads often omit name and carry the text in label. Control characters are
// stripped so injected \n/\r cannot break the one-row-per-task contract.
func taskName(t Task) string {
	for _, cand := range []string{t.Name, t.Label, t.Type} {
		if s := sanitize(cand); s != "" {
			return s
		}
	}
	if id := sanitize(t.ID); id != "" {
		if len(id) > 8 {
			id = id[:8]
		}
		return id
	}
	return "task"
}

// sanitize strips C0/C1 control characters (including \n, \r, \t, ESC) and trims spaces,
// so task-provided text cannot inject newlines, move the cursor, or smuggle ANSI codes.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || (r >= 0x7F && r < 0xA0) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// statusGlyph maps a task status to a compact indicator, ASCII-safe when needed.
func statusGlyph(status string, ascii bool) string {
	if ascii {
		switch status {
		case "running":
			return ">"
		case "completed":
			return "+"
		case "failed", "error":
			return "x"
		case "pending", "queued":
			return "."
		default:
			return "*"
		}
	}
	switch status {
	case "running":
		return "▸"
	case "completed":
		return "●"
	case "failed", "error":
		return "✗"
	case "pending", "queued":
		return "◦"
	default:
		return "•"
	}
}

// formatTokens renders a compact token count for token-only rows (e.g. 12345 -> "12k").
func formatTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%dk", n/1000)
}
