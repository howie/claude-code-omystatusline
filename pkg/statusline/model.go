package statusline

// 輸入資料結構 - 支援 Claude Code v2.0.25+ 的新 JSON 格式
type Input struct {
	HookEventName  string `json:"hook_event_name,omitempty"` // v2.0.25+ 新增
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path,omitempty"`
	Cwd            string `json:"cwd,omitempty"` // v2.0.25+ 新增
	Model          struct {
		ID          string `json:"id,omitempty"` // v2.0.25+ 新增
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
		ProjectDir string `json:"project_dir,omitempty"` // v2.0.25+ 新增
	} `json:"workspace"`
	Version     string `json:"version,omitempty"` // v2.0.25+ 新增
	OutputStyle struct {
		Name string `json:"name,omitempty"`
	} `json:"output_style,omitempty"` // v2.0.25+ 新增
	Cost struct {
		TotalCostUSD       float64 `json:"total_cost_usd,omitempty"`
		TotalDurationMs    int64   `json:"total_duration_ms,omitempty"`
		TotalAPIDurationMs int64   `json:"total_api_duration_ms,omitempty"`
		TotalLinesAdded    int     `json:"total_lines_added,omitempty"`
		TotalLinesRemoved  int     `json:"total_lines_removed,omitempty"`
	} `json:"cost,omitempty"` // v2.0.25+ 新增
}

// 結果通道資料
type Result struct {
	Type string
	Data interface{}
}
