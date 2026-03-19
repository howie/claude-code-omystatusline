package statusline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// ConfigCounts 代表配置檔案統計
type ConfigCounts struct {
	ClaudeMD   int
	MCPServers int
	Hooks      int
}

// ExtractSessionName 從 transcript 行中提取 session 名稱
func ExtractSessionName(lines []transcript.Line, sessionID string) string {
	// 從後往前找 /rename 命令或 session name 設定
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		// 檢查是否有 sessionName 欄位
		if name, ok := l.Parsed["sessionName"].(string); ok && name != "" {
			return name
		}

		// 檢查 /rename 命令
		if msg, ok := l.Parsed["message"].(map[string]interface{}); ok {
			role, _ := msg["role"].(string)
			if role == "user" {
				if content, ok := msg["content"].(string); ok {
					if strings.HasPrefix(content, "/rename ") {
						name := strings.TrimPrefix(content, "/rename ")
						name = strings.TrimSpace(name)
						if name != "" {
							return name
						}
					}
				}
			}
		}
	}

	return ""
}

// CountConfigFiles 統計配置檔案數量
func CountConfigFiles(projectDir string) *ConfigCounts {
	counts := &ConfigCounts{}

	// 計算 CLAUDE.md 檔案
	counts.ClaudeMD = countClaudeMD(projectDir)

	// 計算 MCP servers 和 hooks
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return counts
	}

	// 從 Claude Code settings 讀取 MCP 和 hooks
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	counts.MCPServers, counts.Hooks = parseSettings(settingsPath)

	// 也檢查專案級別的 settings
	projectSettingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	projectMCP, projectHooks := parseSettings(projectSettingsPath)
	counts.MCPServers += projectMCP
	counts.Hooks += projectHooks

	return counts
}

func countClaudeMD(projectDir string) int {
	count := 0

	// 檢查專案根目錄
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err == nil {
		count++
	}

	// 檢查 .claude/CLAUDE.md
	if _, err := os.Stat(filepath.Join(projectDir, ".claude", "CLAUDE.md")); err == nil {
		count++
	}

	// 檢查使用者家目錄
	if homeDir, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(filepath.Join(homeDir, ".claude", "CLAUDE.md")); err == nil {
			count++
		}
	}

	return count
}

func parseSettings(path string) (mcpCount, hookCount int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return 0, 0
	}

	// 計算 MCP servers
	if mcps, ok := settings["mcpServers"].(map[string]interface{}); ok {
		mcpCount = len(mcps)
	}

	// 計算 hooks
	if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
		for _, v := range hooks {
			if hookList, ok := v.([]interface{}); ok {
				hookCount += len(hookList)
			}
		}
	}

	return mcpCount, hookCount
}

// FormatConfigCounts 格式化配置統計為顯示字串
func FormatConfigCounts(counts *ConfigCounts) string {
	if counts == nil {
		return ""
	}

	var parts []string
	if counts.ClaudeMD > 0 {
		parts = append(parts, fmt.Sprintf("%dmd", counts.ClaudeMD))
	}
	if counts.MCPServers > 0 {
		parts = append(parts, fmt.Sprintf("%dmcp", counts.MCPServers))
	}
	if counts.Hooks > 0 {
		parts = append(parts, fmt.Sprintf("%dhook", counts.Hooks))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, " "))
}
