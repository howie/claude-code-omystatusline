package context

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
)

// Analyze 分析 Context 使用量
func Analyze(transcriptPath string) string {
	var contextLength int

	if transcriptPath == "" {
		// 當 transcriptPath 為空時（對話剛開始），顯示初始狀態
		contextLength = 0
	} else {
		contextLength = calculateUsage(transcriptPath)
	}

	// 即使 contextLength 為 0 也顯示進度條

	// 計算百分比（基於 200k tokens）
	percentage := int(float64(contextLength) * 100.0 / 200000.0)
	if percentage > 100 {
		percentage = 100
	}

	// 生成進度條
	progressBar := generateProgressBar(percentage)
	formattedNum := formatNumber(contextLength)
	color := getColor(percentage)

	return fmt.Sprintf(" | %s %s%d%% %s%s",
		progressBar, color, percentage, formattedNum, statusline.ColorReset)
}

// calculateUsage 計算 Context 使用量
func calculateUsage(transcriptPath string) int {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	// 讀取最後100行
	scanner := bufio.NewScanner(file)

	// 設定更大的 buffer（1MB）以處理長 JSON 行
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, 0, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	// 先讀取所有行到切片
	allLines := make([]string, 0)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// 取最後100行
	start := len(allLines) - 100
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	// 從後往前分析
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// 空行跳過
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 先嘗試解析 JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// 檢查 isSidechain 欄位（處理 bool 和可能的其他類型）
		if sidechain, ok := data["isSidechain"]; ok {
			// 如果是 sidechain，跳過
			if isSide, ok := sidechain.(bool); ok && isSide {
				continue
			}
		}

		// 檢查並提取 usage 資料
		if message, ok := data["message"].(map[string]interface{}); ok {
			if usage, ok := message["usage"].(map[string]interface{}); ok {
				var total float64

				// 計算所有 token 類型
				if input, ok := usage["input_tokens"].(float64); ok {
					total += input
				}
				if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
					total += cacheRead
				}
				if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
					total += cacheCreation
				}

				// 如果找到有效的 token 數量，立即返回
				if total > 0 {
					return int(total)
				}
			}
		}
	}

	return 0
}

// generateProgressBar 生成進度條
func generateProgressBar(percentage int) string {
	width := 10
	filled := percentage * width / 100
	if filled > width {
		filled = width
	}

	empty := width - filled
	color := getColor(percentage)

	var bar strings.Builder

	// 填充部分
	if filled > 0 {
		bar.WriteString(color)
		bar.WriteString(strings.Repeat("█", filled))
		bar.WriteString(statusline.ColorReset)
	}

	// 空白部分
	if empty > 0 {
		bar.WriteString(statusline.ColorGray)
		bar.WriteString(strings.Repeat("░", empty))
		bar.WriteString(statusline.ColorReset)
	}

	return bar.String()
}

// getColor 獲取 Context 顏色
func getColor(percentage int) string {
	if percentage < 60 {
		return statusline.ColorCtxGreen
	} else if percentage < 80 {
		return statusline.ColorCtxGold
	}
	return statusline.ColorCtxRed
}

// formatNumber 格式化數字
func formatNumber(num int) string {
	if num == 0 {
		return "--"
	}

	if num >= 1000000 {
		return fmt.Sprintf("%dM", num/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%dk", num/1000)
	}
	return strconv.Itoa(num)
}
