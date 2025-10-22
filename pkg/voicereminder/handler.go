package voicereminder

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SelectMessage 根據 hook 事件和訊息內容選擇要播放的訊息
func SelectMessage(config *Config, input *HookInput) string {
	messages := config.GetMessages()

	switch input.HookEventName {
	case "Notification":
		return selectNotificationMessage(messages["notification"], input.Message)
	case "Stop":
		return pickRandom(messages["stop"].Default)
	case "SubagentStop":
		return pickRandom(messages["subagent_stop"].Default)
	case "PreToolUse":
		return selectPreToolUseMessage(messages["pre_tool_use"], input, config)
	default:
		return "請注意"
	}
}

// selectNotificationMessage 根據通知訊息內容選擇對應的語音
func selectNotificationMessage(eventMsgs EventMessages, message string) string {
	lowerMsg := strings.ToLower(message)

	// 檢查是否需要確認（問號或關鍵字）
	if containsAny(message, "?", "？") || containsAnyKeyword(lowerMsg, "permission", "confirm", "approve") {
		return pickRandom(eventMsgs.Confirmation)
	}

	// 檢查錯誤
	if containsAnyKeyword(lowerMsg, "error", "failed", "fail") {
		return pickRandom(eventMsgs.Error)
	}

	// 檢查完成
	if containsAnyKeyword(lowerMsg, "completed", "finished", "done", "success") {
		return pickRandom(eventMsgs.Completed)
	}

	// 預設訊息
	return pickRandom(eventMsgs.Default)
}

// selectPreToolUseMessage 根據 PreToolUse 事件選擇對應的語音
func selectPreToolUseMessage(eventMsgs EventMessages, input *HookInput, config *Config) string {
	// 如果過濾器未啟用，使用舊邏輯（只提醒 AskUserQuestion）
	if !config.PreToolUseFilters.Enabled {
		if input.ToolName == "AskUserQuestion" {
			return pickRandom(eventMsgs.Confirmation)
		}
		return ""
	}

	// 檢查是否在忽略列表中
	for _, ignoreTool := range config.PreToolUseFilters.IgnoreTools {
		if input.ToolName == ignoreTool {
			return "" // 靜默
		}
	}

	// 檢查是否在通知列表中
	for _, notifyTool := range config.PreToolUseFilters.NotifyTools {
		if input.ToolName == notifyTool {
			// 根據工具類型選擇訊息
			if input.ToolName == "AskUserQuestion" {
				return pickRandom(eventMsgs.Confirmation)
			}
			// 其他需要通知的工具使用預設訊息
			return pickRandom(eventMsgs.Default)
		}
	}

	// 不在任何列表中，使用預設行為（靜默）
	return ""
}

// pickRandom 從字串或字串陣列中隨機選擇一個
func pickRandom(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		if len(val) == 0 {
			return ""
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		idx := r.Intn(len(val))
		if str, ok := val[idx].(string); ok {
			return str
		}
		return ""
	default:
		return ""
	}
}

// containsAny 檢查字串是否包含任一子字串
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// containsAnyKeyword 檢查小寫字串是否包含任一關鍵字
func containsAnyKeyword(lowerStr string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(lowerStr, keyword) {
			return true
		}
	}
	return false
}

// UpdateStats 更新觸發統計
func UpdateStats(eventName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 新路徑優先
	newStatsPath := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "data", "voice-reminder-stats.json")
	// 舊路徑作為備援
	oldStatsPath := filepath.Join(homeDir, ".claude", "voice-reminder-stats.json")

	var statsPath string

	// 確保新路徑的目錄存在
	newStatsDir := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "data")
	if err := os.MkdirAll(newStatsDir, 0755); err == nil {
		statsPath = newStatsPath
	} else {
		// 回退到舊路徑
		statsPath = oldStatsPath
	}

	// 讀取現有統計
	var stats Stats
	data, err := os.ReadFile(statsPath)
	if err == nil {
		_ = json.Unmarshal(data, &stats)
	}

	// 更新統計
	stats.LastTriggered = time.Now()
	switch eventName {
	case "Notification":
		stats.NotificationCount++
	case "Stop":
		stats.StopCount++
	case "SubagentStop":
		stats.SubagentStopCount++
	case "PreToolUse":
		stats.PreToolUseCount++
	}

	// 寫回檔案
	data, err = json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statsPath, data, 0644)
}
