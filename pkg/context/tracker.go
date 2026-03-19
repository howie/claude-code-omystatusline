package context

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// DefaultMaxTokens 預設 token 上限
const DefaultMaxTokens = 200000

// ContextData 包含 context 分析的完整結果
type ContextData struct {
	Formatted  string // 格式化的進度條字串
	Percentage int    // 百分比數值
	Tokens     int    // token 數量
}

// Analyze 分析 Context 使用量（向後相容）
// maxTokens <= 0 時使用 DefaultMaxTokens
func Analyze(transcriptPath string, maxTokens int) string {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	var contextLength int

	if transcriptPath == "" {
		contextLength = 0
	} else {
		contextLength = calculateUsage(transcriptPath)
	}

	return formatContext(contextLength, maxTokens)
}

// AnalyzeFromLines 使用共享 transcript 行分析 Context 使用量
func AnalyzeFromLines(lines []transcript.Line, maxTokens int) string {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	contextLength := calculateUsageFromLines(lines)
	return formatContext(contextLength, maxTokens)
}

// AnalyzeDetailedFromLines 返回詳細的 context 資料（用於進階顯示）
func AnalyzeDetailedFromLines(lines []transcript.Line, maxTokens int) *ContextData {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	contextLength := calculateUsageFromLines(lines)
	percentage := int(float64(contextLength) * 100.0 / float64(maxTokens))
	if percentage > 100 {
		percentage = 100
	}

	return &ContextData{
		Formatted:  formatContext(contextLength, maxTokens),
		Percentage: percentage,
		Tokens:     contextLength,
	}
}

func formatContext(contextLength, maxTokens int) string {
	percentage := int(float64(contextLength) * 100.0 / float64(maxTokens))
	if percentage > 100 {
		percentage = 100
	}

	progressBar := generateProgressBar(percentage)
	formattedNum := formatNumber(contextLength)
	color := getColor(percentage)

	return fmt.Sprintf(" | %s %s%d%% %s%s",
		progressBar, color, percentage, formattedNum, statusline.ColorReset)
}

// calculateUsageFromLines 從共享 transcript 行計算 Context 使用量
func calculateUsageFromLines(lines []transcript.Line) int {
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		// 檢查 isSidechain
		if isSide, ok := l.Parsed["isSidechain"].(bool); ok && isSide {
			continue
		}

		// 檢查並提取 usage 資料
		if message, ok := l.Parsed["message"].(map[string]interface{}); ok {
			if usage, ok := message["usage"].(map[string]interface{}); ok {
				var total float64

				if input, ok := usage["input_tokens"].(float64); ok {
					total += input
				}
				if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
					total += cacheRead
				}
				if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
					total += cacheCreation
				}

				if total > 0 {
					return int(total)
				}
			}
		}
	}

	return 0
}

// calculateUsage 計算 Context 使用量（從檔案讀取，向後相容）
func calculateUsage(transcriptPath string) int {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return 0
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	const maxScanTokenSize = 1024 * 1024
	buf := make([]byte, 0, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	allLines := make([]string, 0)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := len(allLines) - 100
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		if sidechain, ok := data["isSidechain"]; ok {
			if isSide, ok := sidechain.(bool); ok && isSide {
				continue
			}
		}

		if message, ok := data["message"].(map[string]interface{}); ok {
			if usage, ok := message["usage"].(map[string]interface{}); ok {
				var total float64

				if input, ok := usage["input_tokens"].(float64); ok {
					total += input
				}
				if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
					total += cacheRead
				}
				if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
					total += cacheCreation
				}

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
