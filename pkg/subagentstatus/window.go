package subagentstatus

import "github.com/howie/claude-code-omystatusline/pkg/modelwindow"

// resolveWindow decides the percentage denominator for one task, applying the same
// two-part guard that protects the main statusline from the #35/#36 "stuck at 100%"
// bug: a reported size is trusted only when it is a KNOWN capacity (200K/1M) AND it
// does not demote a confident inference.
//
// Returns (denominator, true) when a percentage may be drawn, or (0, false) when the
// caller must fall back to a token-only display (no guessed denominator — a wrong
// guess of 200K would make a real 1M subagent look full, the exact bug we avoid).
func resolveWindow(model string, reported int) (denominator int, ok bool) {
	inferred, confident := modelwindow.Infer(model)
	if isKnownCapacity(reported) {
		// Non-demotion: a confident inference is never lowered by a smaller reported
		// value (e.g. inferred 1M, reported 200K -> keep 1M).
		if confident && inferred > reported {
			return inferred, true
		}
		return reported, true
	}
	// Reported is absent or not a recognized capacity: use inference only if confident,
	// otherwise token-only.
	if confident {
		return inferred, true
	}
	return 0, false
}

func isKnownCapacity(n int) bool {
	return n == modelwindow.Window200K || n == modelwindow.Window1M
}
