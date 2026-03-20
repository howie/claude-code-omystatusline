package tools

import (
	"fmt"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// ToolInfo 代表一個工具（正在執行或已完成）
type ToolInfo struct {
	Name      string
	Target    string // 截斷的路徑或參數
	Completed bool   // 是否已完成
}

// MaxTools 最多顯示的工具數量
const MaxTools = 2

// Analyze 從 transcript 行中找出正在執行的工具
func Analyze(lines []transcript.Line) []ToolInfo {
	// 追蹤工具狀態：tool_use 開始，tool_result 結束
	// 使用 toolUseId 來匹配
	activeTools := make(map[string]ToolInfo) // toolUseId -> ToolInfo
	completedTools := make(map[string]bool)  // 已完成的 toolUseId
	var toolOrder []string                   // 保持順序

	for _, l := range lines {
		if l.Parsed == nil {
			continue
		}

		msg, ok := l.Parsed["message"].(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		content, _ := msg["content"].([]interface{})

		for _, block := range content {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}

			blockType, _ := blockMap["type"].(string)

			if role == "assistant" && blockType == "tool_use" {
				toolID, _ := blockMap["id"].(string)
				name, _ := blockMap["name"].(string)
				if toolID == "" || name == "" {
					continue
				}

				target := extractTarget(blockMap)
				activeTools[toolID] = ToolInfo{Name: name, Target: target}
				toolOrder = append(toolOrder, toolID)
			}

			if role == "user" && blockType == "tool_result" {
				toolID, _ := blockMap["tool_use_id"].(string)
				if toolID != "" {
					completedTools[toolID] = true
				}
			}
		}
	}

	// 先找出仍在執行的工具（從最新往回）
	var running []ToolInfo
	for i := len(toolOrder) - 1; i >= 0; i-- {
		id := toolOrder[i]
		if !completedTools[id] {
			running = append(running, activeTools[id])
			if len(running) >= MaxTools {
				break
			}
		}
	}

	// 若沒有正在執行的工具，顯示最近完成的工具作為 fallback
	if len(running) == 0 {
		for i := len(toolOrder) - 1; i >= 0; i-- {
			id := toolOrder[i]
			if completedTools[id] {
				info := activeTools[id]
				info.Completed = true
				running = append(running, info)
				if len(running) >= MaxTools {
					break
				}
			}
		}
	}

	// 反轉順序，讓最舊的在前
	for i, j := 0, len(running)-1; i < j; i, j = i+1, j-1 {
		running[i], running[j] = running[j], running[i]
	}

	return running
}

// extractTarget 從工具輸入中提取目標路徑
func extractTarget(block map[string]interface{}) string {
	input, ok := block["input"].(map[string]interface{})
	if !ok {
		return ""
	}

	// 常見的路徑欄位
	for _, key := range []string{"file_path", "path", "command", "pattern", "url"} {
		if val, ok := input[key].(string); ok && val != "" {
			return truncatePath(val, 30)
		}
	}

	return ""
}

// truncatePath 截斷路徑顯示
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// 嘗試只保留檔名
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		short := fmt.Sprintf(".../%s", parts[len(parts)-1])
		if len(short) <= maxLen {
			return short
		}
	}

	return path[:maxLen-3] + "..."
}

// Format 格式化工具列表為顯示字串
func Format(tools []ToolInfo) string {
	if len(tools) == 0 {
		return ""
	}

	var parts []string
	for _, t := range tools {
		icon := "◐"
		if t.Completed {
			icon = "✓"
		}
		if t.Target != "" {
			parts = append(parts, fmt.Sprintf("%s %s: %s", icon, t.Name, t.Target))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", icon, t.Name))
		}
	}

	return strings.Join(parts, "  ")
}
