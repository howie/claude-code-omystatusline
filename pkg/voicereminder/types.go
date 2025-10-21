package voicereminder

import "time"

// Config 語音提醒配置
type Config struct {
	DebugMode    bool                     `json:"debug_mode"`
	Language     string                   `json:"language"`
	Speed        int                      `json:"speed"`
	Messages     map[string]EventMessages `json:"messages"`
	MessagesEN   map[string]EventMessages `json:"messages_en"`
	SoundEffects SoundConfig              `json:"sound_effects"`
}

// EventMessages 事件訊息配置
type EventMessages struct {
	Confirmation interface{} `json:"confirmation"` // string or []string
	Error        interface{} `json:"error"`
	Completed    interface{} `json:"completed"`
	Default      interface{} `json:"default"`
}

// SoundConfig 音效配置
type SoundConfig struct {
	Enabled       bool   `json:"enabled"`
	FallbackSound string `json:"fallback_sound"`
}

// HookInput Claude Code Hook 輸入
type HookInput struct {
	Message        string                 `json:"message"`
	HookEventName  string                 `json:"hook_event_name"`
	SessionID      string                 `json:"session_id"`
	TranscriptPath string                 `json:"transcript_path"`
	Cwd            string                 `json:"cwd"`
	ToolName       string                 `json:"tool_name"`        // PreToolUse/PostToolUse 專用
	ToolInput      map[string]interface{} `json:"tool_input"`       // PreToolUse/PostToolUse 專用
}

// Stats 觸發統計
type Stats struct {
	NotificationCount int       `json:"notification_count"`
	StopCount         int       `json:"stop_count"`
	SubagentStopCount int       `json:"subagent_stop_count"`
	PreToolUseCount   int       `json:"pre_tool_use_count"`
	LastTriggered     time.Time `json:"last_triggered"`
}

// GetMessages 根據語言返回對應的訊息配置
func (c *Config) GetMessages() map[string]EventMessages {
	if c.Language == "zh" {
		return c.Messages
	}
	return c.MessagesEN
}
