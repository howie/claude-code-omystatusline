// Package subagentstatus renders Claude Code's subagentStatusLine payload: the
// per-subagent rows shown in the agent panel below the prompt. Its input schema is
// distinct from the main statusline's flat single-session object — it carries all
// visible subagent rows in one JSON object per refresh tick.
package subagentstatus

// Input mirrors the subagentStatusLine stdin payload. Only the fields this renderer
// needs are modeled; encoding/json ignores the rest (real payloads also carry
// agent_type, cwd, prompt_id, session_id, transcript_path at the top level).
type Input struct {
	// Columns is the usable row width Claude Code allots each rendered row.
	Columns int    `json:"columns,omitempty"`
	Tasks   []Task `json:"tasks,omitempty"`
}

// Task is one subagent row. Per-task model/contextWindowSize require Claude Code
// v2.1.205+ and are omitted for a task whose model isn't resolved yet, so every
// field is treated as optional. tokenSamples/startTime/cwd are not needed for the
// v1 row. effort is intentionally NOT modeled: it was absent from every captured
// payload and its exact JSON type is unverified — adding it blind risks a decode error.
type Task struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Status      string `json:"status,omitempty"`
	// Model is the resolved model ID; real payloads carry the "[1m]" suffix
	// (e.g. "claude-opus-4-8[1m]"), which modelwindow.Infer handles.
	Model string `json:"model,omitempty"`
	// ContextWindowSize is Claude Code's reported per-task window. Trusted as a
	// percentage denominator only via resolveWindow's two-part guard.
	ContextWindowSize int `json:"contextWindowSize,omitempty"`
	// TokenCount is the subagent's CURRENT context occupancy (verified via spike:
	// gauge, not cumulative), so it is a valid numerator for the context bar.
	TokenCount int `json:"tokenCount,omitempty"`
}
