package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/subagentstatus"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
)

// runMode selects which status line the binary renders. The mode is chosen from argv
// BEFORE stdin is read or decoded, because the two modes take incompatible JSON schemas.
type runMode int

const (
	modeMain runMode = iota
	modeSubagent
)

// parseArgs resolves the run mode from CLI arguments. Only the "--subagent" flag is
// recognized; anything else (unknown flag or stray positional) is an error so a
// misconfigured command fails loudly instead of silently rendering the wrong schema.
func parseArgs(args []string) (runMode, error) {
	mode := modeMain
	for _, a := range args {
		switch a {
		case "--subagent":
			mode = modeSubagent
		default:
			return modeMain, fmt.Errorf("unexpected argument %q", a)
		}
	}
	return mode, nil
}

// renderSubagentJSON decodes a subagentStatusLine payload and renders its rows. It is
// the pure, testable core of the subagent path: no stdin, no os.Exit. An empty/absent
// tasks list yields "" (zero bytes). A flat main-line payload decodes with no tasks and
// so also yields "" — fail-safe rather than garbage.
func renderSubagentJSON(raw []byte) (string, error) {
	var in subagentstatus.Input
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", err
	}
	return subagentstatus.Render(in), nil
}

// runSubagent is the I/O wrapper around renderSubagentJSON for the --subagent path.
func runSubagent() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: failed to read subagent input: %v\n", err)
		os.Exit(1)
	}
	// Same inert capture hook as the main path (owner-only; may contain task text).
	if dumpPath := os.Getenv("STATUSLINE_DUMP_INPUT"); dumpPath != "" {
		if werr := os.WriteFile(dumpPath, raw, 0o600); werr != nil {
			fmt.Fprintf(os.Stderr, "statusline: STATUSLINE_DUMP_INPUT write to %q failed: %v\n", dumpPath, werr)
		}
	}
	context.RenderMode = terminal.Detect()
	out, err := renderSubagentJSON(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: failed to decode subagent input: %v\n", err)
		os.Exit(1)
	}
	if out != "" {
		fmt.Println(out)
	}
}
