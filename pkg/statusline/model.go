package statusline

// 輸入資料結構 - 支援 Claude Code v2.0.25+ 的新 JSON 格式
type Input struct {
	HookEventName string `json:"hook_event_name,omitempty"` // v2.0.25+ 新增
	SessionID     string `json:"session_id"`
	// SessionName 為 Claude Code v2.1.x+ 直接提供的 session 名稱（由 /rename 設定）。
	// 存在時優先使用，免去掃描 transcript（ExtractSessionName）；worktree session 的
	// metadata-only transcript 也能正確取得。
	SessionName    string `json:"session_name,omitempty"`
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
	// ContextWindow 為 Claude Code 直接提供的 context window 使用量。
	// 此欄位比 transcript 解析更準確（worktree session 也能正確讀取），
	// 應優先使用此欄位，transcript 作為 fallback。
	// ContextWindowSize == 0 表示舊版 Claude Code 不提供此資料；> 0 才視為有效。
	ContextWindow struct {
		// TotalInputTokens / TotalOutputTokens：整個 session 的累計量，僅供參考，不用於進度條。
		TotalInputTokens  int `json:"total_input_tokens,omitempty"`
		TotalOutputTokens int `json:"total_output_tokens,omitempty"`
		// ContextWindowSize is the current token count in the context window sent by Claude Code.
		// This is NOT the model's maximum capacity (e.g. 1M for Sonnet 4.6).
		// Use only to detect whether Claude Code provides CurrentUsage data (> 0 means yes).
		// Never use as the percentage denominator; use contextWindowForModel() instead.
		ContextWindowSize int `json:"context_window_size,omitempty"`
		CurrentUsage      struct {
			InputTokens int `json:"input_tokens,omitempty"`
			// OutputTokens 已解析但不計入 context 使用量：
			// context window 壓力由 input+cache tokens 決定；output tokens 延伸自同一 window，
			// 不單獨占用 context 空間。與 tracker.go usageFromLines 的計算方式一致。
			OutputTokens             int `json:"output_tokens,omitempty"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
		} `json:"current_usage,omitempty"`
		// UsedPercentage / RemainingPercentage：由 Claude Code 計算，僅供參考。
		// 進度條使用 BuildFromTokens 自行計算，保持與 bar 渲染邏輯一致。
		UsedPercentage      int `json:"used_percentage,omitempty"`
		RemainingPercentage int `json:"remaining_percentage,omitempty"`
	} `json:"context_window,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	Worktree  struct {
		Name string `json:"name,omitempty"`
		Path string `json:"path,omitempty"`
		// Branch 為 worktree 目前所在的分支（對應官方 schema 的 "branch"）。
		Branch string `json:"branch,omitempty"`
		// OriginalCwd 為進入 worktree 前的原始工作目錄（官方 schema key 為 "original_cwd"）。
		// 注意：舊版本曾誤用 "original_repo_dir"，但官方 statusline schema 從未提供該 key，
		// 故 OriginalCwd 在修正前永遠為空字串。
		OriginalCwd string `json:"original_cwd,omitempty"`
		// OriginalBranch 為進入 worktree 前的原始分支（官方 schema 的 "original_branch"）。
		OriginalBranch string `json:"original_branch,omitempty"`
	} `json:"worktree,omitempty"`
	// RateLimits 為 Claude Code v2.1.x+ 直接提供的 API 配額使用量。
	// 存在時可直接顯示，免去向 OAuth usage API 發 HTTP 請求（見 pkg/apilimits）。
	// 欄位語意與 OAuth API 不同：UsedPercentage 已是 0-100 的百分比（非 0-1 分數），
	// ResetsAt 為 Unix epoch 秒（非 RFC3339 字串）。
	// 判斷是否提供此資料：ResetsAt > 0（feature 存在時一定有重置時間）。
	RateLimits struct {
		FiveHour struct {
			UsedPercentage float64 `json:"used_percentage,omitempty"`
			ResetsAt       int64   `json:"resets_at,omitempty"`
		} `json:"five_hour,omitempty"`
		SevenDay struct {
			UsedPercentage float64 `json:"used_percentage,omitempty"`
			ResetsAt       int64   `json:"resets_at,omitempty"`
		} `json:"seven_day,omitempty"`
	} `json:"rate_limits,omitempty"`
}

// 結果通道資料
type Result struct {
	Type string
	Data interface{}
}
