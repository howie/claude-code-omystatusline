package statusline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// 模型圖示和顏色
var modelConfig = map[string][2]string{
	"Opus":   {ColorGold, "💛"},
	"Sonnet": {ColorCyan, "💠"},
	"Haiku":  {ColorPink, "🌸"},
}

// FormatModel 格式化模型顯示
func FormatModel(model string) string {
	for key, config := range modelConfig {
		if strings.Contains(model, key) {
			color := config[0]
			icon := config[1]
			return fmt.Sprintf("%s%s %s%s", color, icon, model, ColorReset)
		}
	}
	return model
}

// ExtractUserMessage 提取使用者訊息（向後相容，從檔案讀取）
func ExtractUserMessage(transcriptPath, sessionID string) string {
	if transcriptPath == "" {
		return ""
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	// 讀取最後200行
	allLines := readAllLines(file)

	start := len(allLines) - 200
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	return findUserMessage(lines, sessionID)
}

// ExtractUserMessageFromLines 使用共享 transcript 行提取使用者訊息
func ExtractUserMessageFromLines(lines []transcript.Line, sessionID string) string {
	// 從後往前搜尋使用者訊息
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		isSidechain, _ := l.Parsed["isSidechain"].(bool)
		if isSidechain {
			continue
		}

		sid, _ := l.Parsed["sessionId"].(string)
		if sid != sessionID {
			continue
		}

		if message, ok := l.Parsed["message"].(map[string]interface{}); ok {
			role, _ := message["role"].(string)
			msgType, _ := l.Parsed["type"].(string)

			if role == "user" && msgType == "user" {
				if content, ok := message["content"].(string); ok {
					if isSystemMessage(content) {
						continue
					}
					return formatUserMessage(content)
				}
			}
		}
	}

	return ""
}

// readAllLines 從 file 讀取所有行
func readAllLines(file *os.File) []string {
	scanner := bufio.NewScanner(file)
	const maxScanTokenSize = 1024 * 1024
	buf := make([]byte, 0, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// findUserMessage 從原始字串行中找使用者訊息
func findUserMessage(lines []string, sessionID string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		isSidechain, _ := data["isSidechain"].(bool)
		sessionMatch := false
		if sid, ok := data["sessionId"].(string); ok && sid == sessionID {
			sessionMatch = true
		}

		if !isSidechain && sessionMatch {
			if message, ok := data["message"].(map[string]interface{}); ok {
				role, _ := message["role"].(string)
				msgType, _ := data["type"].(string)

				if role == "user" && msgType == "user" {
					if content, ok := message["content"].(string); ok {
						if isSystemMessage(content) {
							continue
						}
						return formatUserMessage(content)
					}
				}
			}
		}
	}

	return ""
}

// isSystemMessage 檢查是否為系統訊息
func isSystemMessage(content string) bool {
	// 過濾 JSON 格式
	if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
		return true
	}
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return true
	}

	// 過濾 XML 標籤
	xmlTags := []string{
		"<local-command-stdout>", "<command-name>",
		"<command-message>", "<command-args>",
	}
	for _, tag := range xmlTags {
		if strings.Contains(content, tag) {
			return true
		}
	}

	// 過濾 Caveat 訊息
	if strings.HasPrefix(content, "Caveat:") {
		return true
	}

	return false
}

// formatUserMessage 格式化使用者訊息
func formatUserMessage(message string) string {
	if message == "" {
		return ""
	}

	maxLines := 3
	lineWidth := 80

	lines := strings.Split(message, "\n")
	var result []string

	for i, line := range lines {
		if i >= maxLines {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) > lineWidth {
			line = line[:lineWidth-3] + "..."
		}

		result = append(result, fmt.Sprintf("%s｜%s%s%s",
			ColorReset, ColorGreen, line, ColorReset))
	}

	if len(lines) > maxLines {
		result = append(result, fmt.Sprintf("%s｜... (還有 %d 行)%s",
			ColorReset, len(lines)-maxLines, ColorReset))
	}

	if len(result) > 0 {
		return strings.Join(result, "\n") + "\n"
	}

	return ""
}

// RenderLine 建構一行狀態列
type RenderLine struct {
	Parts []string
}

// Add 加入一個部分（忽略空字串）
func (rl *RenderLine) Add(parts ...string) {
	for _, p := range parts {
		if p != "" {
			rl.Parts = append(rl.Parts, p)
		}
	}
}

// String 輸出行內容
func (rl *RenderLine) String() string {
	return strings.Join(rl.Parts, "")
}

// FormatToolsLine 格式化工具行
func FormatToolsLine(toolsStr string) string {
	if toolsStr == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", ColorYellow, toolsStr, ColorReset)
}

// FormatAgentsLine 格式化代理行
func FormatAgentsLine(agentsStr string) string {
	if agentsStr == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", ColorMagenta, agentsStr, ColorReset)
}

// FormatTodoLine 格式化 Todo 行
func FormatTodoLine(todoStr, limitsStr string) string {
	if todoStr == "" && limitsStr == "" {
		return ""
	}

	var parts []string
	if todoStr != "" {
		parts = append(parts, fmt.Sprintf("%s%s%s", ColorYellow, todoStr, ColorReset))
	}
	if limitsStr != "" {
		parts = append(parts, fmt.Sprintf("%s%s%s", ColorBrightBlue, limitsStr, ColorReset))
	}

	return strings.Join(parts, " | ")
}

// FormatSpeedDisplay 格式化速度顯示
func FormatSpeedDisplay(speedStr string) string {
	if speedStr == "" {
		return ""
	}
	return fmt.Sprintf(" %s%s%s", ColorDim, speedStr, ColorReset)
}

// FormatSessionNameDisplay 格式化 session 名稱
func FormatSessionNameDisplay(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf(" %s[%s]%s", ColorDim, name, ColorReset)
}

// FormatConfigCountsDisplay 格式化配置統計
func FormatConfigCountsDisplay(countsStr string) string {
	if countsStr == "" {
		return ""
	}
	return fmt.Sprintf(" %s%s%s", ColorDim, countsStr, ColorReset)
}

// FormatAutocompactDisplay 格式化自動壓縮警告
func FormatAutocompactDisplay(autocompactStr string) string {
	if autocompactStr == "" {
		return ""
	}
	return fmt.Sprintf(" %s%s%s", ColorRed, autocompactStr, ColorReset)
}

// FormatGitStatusDisplay 格式化 Git 狀態
func FormatGitStatusDisplay(gitStatusStr string) string {
	if gitStatusStr == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", ColorDim, gitStatusStr, ColorReset)
}
