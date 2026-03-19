package todo

import (
	"fmt"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// TodoInfo 代表 Todo 追蹤狀態
type TodoInfo struct {
	InProgressName string // 目前進行中的任務名稱
	Completed      int    // 已完成數量
	Total          int    // 總數
	AllComplete    bool   // 是否全部完成
}

// Analyze 從 transcript 行中找出最新的 TodoWrite 狀態
func Analyze(lines []transcript.Line) *TodoInfo {
	// 從後往前找最新的 TodoWrite tool_use
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		msg, ok := l.Parsed["message"].(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		if role != "assistant" {
			continue
		}

		content, ok := msg["content"].([]interface{})
		if !ok {
			continue
		}

		for _, block := range content {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}

			blockType, _ := blockMap["type"].(string)
			name, _ := blockMap["name"].(string)

			if blockType == "tool_use" && name == "TodoWrite" {
				return parseTodoInput(blockMap)
			}
		}
	}

	return nil
}

func parseTodoInput(block map[string]interface{}) *TodoInfo {
	input, ok := block["input"].(map[string]interface{})
	if !ok {
		return nil
	}

	todos, ok := input["todos"].([]interface{})
	if !ok {
		return nil
	}

	info := &TodoInfo{Total: len(todos)}
	var inProgressName string

	for _, item := range todos {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		status, _ := itemMap["status"].(string)
		content, _ := itemMap["content"].(string)

		switch status {
		case "completed":
			info.Completed++
		case "in_progress":
			if inProgressName == "" {
				inProgressName = content
			}
		}
	}

	info.InProgressName = truncateContent(inProgressName, 50)
	info.AllComplete = info.Completed == info.Total && info.Total > 0

	return info
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}

// Format 格式化 Todo 資訊為顯示字串
func Format(info *TodoInfo) string {
	if info == nil || info.Total == 0 {
		return ""
	}

	if info.AllComplete {
		return fmt.Sprintf("✓ All complete (%d/%d)", info.Completed, info.Total)
	}

	var parts []string

	if info.InProgressName != "" {
		parts = append(parts, fmt.Sprintf("▸ %s", info.InProgressName))
	}

	parts = append(parts, fmt.Sprintf("(%d/%d)", info.Completed, info.Total))

	return strings.Join(parts, " ")
}
