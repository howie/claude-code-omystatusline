// Package modelwindow maps a Claude model ID to its context window size.
//
// It is deliberately dependency-free (standard library only) and side-effect-free
// (no logging, no globals) so it can be shared by both the main status line and the
// per-subagent status line without one caller's policy or stderr noise leaking into
// the other. Callers layer their own precedence and warnings on top of Infer.
package modelwindow

import (
	"strconv"
	"strings"
)

// Context window capacities (official Anthropic specs).
const (
	Window200K = 200_000
	Window1M   = 1_000_000
)

// Infer reports a model's context window size and whether that size was determined
// with confidence.
//
// confident is false in exactly two cases: an empty modelID, or an unrecognized
// non-versioned family that only reaches the fail-safe 200K guess. Callers that must
// never render a falsely-inflated percentage from a guessed denominator (the
// per-subagent bar) should treat !confident as "no trusted denominator" and fall
// back to a token-only display. The main status line ignores confident and always
// uses size, preserving its historical fail-safe behavior.
//
// Rules (an explicit "[1m]" marker wins over family/version inference):
//   - "[1m]" in the ID              -> 1M, confident
//   - Haiku                         -> 200K, confident
//   - Sonnet/Opus/Fable            -> 1M if major>=5 or 4.6+, else 200K; confident
//   - Unknown family, version>=1M  -> 1M, confident
//   - Unknown family, guessed      -> 200K, NOT confident
//   - Empty                         -> 200K, NOT confident
func Infer(modelID string) (size int, confident bool) {
	id := strings.ToLower(modelID)
	if id == "" {
		return Window200K, false
	}
	if strings.Contains(id, "[1m]") {
		return Window1M, true
	}
	switch {
	case strings.Contains(id, "haiku"):
		return Window200K, true
	case strings.Contains(id, "sonnet"), strings.Contains(id, "opus"), strings.Contains(id, "fable"):
		if is1M(id) {
			return Window1M, true
		}
		return Window200K, true
	default:
		// Unknown non-empty family: apply the version rule so the next new family
		// (e.g. claude-nova-5) is treated as 1M instead of silently capped at 200K.
		// A version-resolved 1M is confident; a bare 200K guess is not.
		if is1M(id) {
			return Window1M, true
		}
		return Window200K, false
	}
}

// is1M reports whether a lowercased model ID has a 1M context window by version:
// major>=5, or major==4 && minor>=6.
func is1M(id string) bool {
	major, minor := version(id)
	return major >= 5 || (major == 4 && minor >= 6)
}

// version parses (major, minor) from a claude-*-{major}-{minor}[-date] ID.
// It scans from the end for the last two consecutive small integers (< 100, to skip
// date suffixes like 20250514). Failing that, it falls back to the last standalone
// small integer as {major}.0 (e.g. claude-fable-5 -> (5, 0)). Returns (-1, -1) when
// nothing parseable is found. The input is expected to already be lowercased.
func version(id string) (int, int) {
	parts := strings.Split(id, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		minor, err1 := strconv.Atoi(parts[i])
		major, err2 := strconv.Atoi(parts[i-1])
		if err1 != nil || err2 != nil {
			continue
		}
		if major > 0 && major < 100 && minor >= 0 && minor < 100 {
			return major, minor
		}
	}
	for i := len(parts) - 1; i >= 0; i-- {
		major, err := strconv.Atoi(parts[i])
		if err == nil && major > 0 && major < 100 {
			return major, 0
		}
	}
	return -1, -1
}
