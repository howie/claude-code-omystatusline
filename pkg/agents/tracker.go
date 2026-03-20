package agents

import (
	"fmt"
	"strings"
	"time"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// AgentInfo 代表一個正在執行的子代理
type AgentInfo struct {
	Type        string // e.g. "Agent", "Explore"
	Model       string // e.g. "Sonnet", "Haiku"
	Description string
	StartTime   time.Time
	ElapsedSec  int
	Completed   bool
}

// MaxAgents 最多顯示的代理數量
const MaxAgents = 3

// Analyze 從 transcript 行中找出正在執行的子代理
func Analyze(lines []transcript.Line, sessionID string) []AgentInfo {
	// 追蹤 agent 生命週期
	activeAgents := make(map[string]*AgentInfo) // agentId -> AgentInfo
	var agentOrder []string

	now := time.Now()

	for _, l := range lines {
		if l.Parsed == nil {
			continue
		}

		// 檢查 agent 相關事件
		agentID, _ := l.Parsed["agentId"].(string)
		agentType, _ := l.Parsed["agent_type"].(string)
		hookEvent, _ := l.Parsed["hook_event_name"].(string)
		msgType, _ := l.Parsed["type"].(string)

		// 檢查 tool_use 中的 Agent 工具呼叫
		if msg, ok := l.Parsed["message"].(map[string]interface{}); ok {
			role, _ := msg["role"].(string)
			if role == "assistant" {
				if content, ok := msg["content"].([]interface{}); ok {
					for _, block := range content {
						blockMap, ok := block.(map[string]interface{})
						if !ok {
							continue
						}
						blockType, _ := blockMap["type"].(string)
						name, _ := blockMap["name"].(string)
						toolID, _ := blockMap["id"].(string)

						if blockType == "tool_use" && name == "Agent" {
							input, _ := blockMap["input"].(map[string]interface{})
							desc, _ := input["description"].(string)
							subType, _ := input["subagent_type"].(string)
							if subType == "" {
								subType = "Agent"
							}

							// 提取時間戳
							startTime := extractTimestamp(l.Parsed, now)

							activeAgents[toolID] = &AgentInfo{
								Type:        subType,
								Description: truncateDesc(desc, 40),
								StartTime:   startTime,
							}
							agentOrder = append(agentOrder, toolID)
						}
					}
				}
			}

			// 檢查 tool_result 來標記完成
			if role == "user" {
				if content, ok := msg["content"].([]interface{}); ok {
					for _, block := range content {
						blockMap, ok := block.(map[string]interface{})
						if !ok {
							continue
						}
						blockType, _ := blockMap["type"].(string)
						toolUseID, _ := blockMap["tool_use_id"].(string)
						if blockType == "tool_result" && toolUseID != "" {
							if agent, exists := activeAgents[toolUseID]; exists {
								agent.Completed = true
							}
						}
					}
				}
			}
		}

		// 處理子代理的事件
		if agentID != "" && agentType != "" {
			if hookEvent == "SubagentStop" || msgType == "agent_stop" {
				// 標記完成
				for _, a := range activeAgents {
					if a.Type == agentType && !a.Completed {
						a.Completed = true
						break
					}
				}
			}
		}
	}

	// 收集未完成的 agents，計算耗時
	var running []AgentInfo
	for i := len(agentOrder) - 1; i >= 0; i-- {
		id := agentOrder[i]
		agent := activeAgents[id]
		if agent != nil && !agent.Completed {
			agent.ElapsedSec = int(now.Sub(agent.StartTime).Seconds())
			if agent.ElapsedSec < 0 {
				agent.ElapsedSec = 0
			}
			running = append(running, *agent)
			if len(running) >= MaxAgents {
				break
			}
		}
	}

	// 若沒有正在執行的 agents，顯示最近完成的作為 fallback
	if len(running) == 0 {
		for i := len(agentOrder) - 1; i >= 0; i-- {
			id := agentOrder[i]
			agent := activeAgents[id]
			if agent != nil && agent.Completed {
				agent.ElapsedSec = int(now.Sub(agent.StartTime).Seconds())
				if agent.ElapsedSec < 0 {
					agent.ElapsedSec = 0
				}
				running = append(running, *agent)
				if len(running) >= MaxAgents {
					break
				}
			}
		}
	}

	// 反轉
	for i, j := 0, len(running)-1; i < j; i, j = i+1, j-1 {
		running[i], running[j] = running[j], running[i]
	}

	return running
}

// extractTimestamp 從 transcript 行提取時間戳
func extractTimestamp(parsed map[string]interface{}, fallback time.Time) time.Time {
	if ts, ok := parsed["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t
		}
	}
	if ts, ok := parsed["timestamp"].(float64); ok {
		return time.Unix(int64(ts/1000), 0)
	}
	return fallback
}

func truncateDesc(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen-3] + "..."
}

// Format 格式化代理列表為顯示字串
func Format(agents []AgentInfo) string {
	if len(agents) == 0 {
		return ""
	}

	var parts []string
	for _, a := range agents {
		elapsed := formatElapsed(a.ElapsedSec)
		icon := "◐"
		if a.Completed {
			icon = "✓"
		}

		model := ""
		if a.Model != "" {
			model = fmt.Sprintf(" [%s]", a.Model)
		}

		desc := ""
		if a.Description != "" {
			desc = fmt.Sprintf(" \"%s\"", a.Description)
		}

		parts = append(parts, fmt.Sprintf("%s %s%s%s %s", icon, a.Type, model, desc, elapsed))
	}

	return strings.Join(parts, "  ")
}

func formatElapsed(seconds int) string {
	if seconds < 1 {
		return "<1s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	m := seconds / 60
	s := seconds % 60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm%ds", m, s)
}
