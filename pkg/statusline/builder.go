package statusline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// ExtractUserMessage 提取使用者訊息
func ExtractUserMessage(transcriptPath, sessionID string) string {
	if transcriptPath == "" {
		return ""
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// 讀取最後200行
	scanner := bufio.NewScanner(file)

	allLines := make([]string, 0)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := len(allLines) - 200
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	// 從後往前搜尋使用者訊息
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// 檢查是否為當前 session 的使用者訊息
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
						// 過濾系統訊息
						if isSystemMessage(content) {
							continue
						}

						// 格式化並返回
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
