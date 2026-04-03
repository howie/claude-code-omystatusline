package context

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// RenderMode 用於控制進度條的渲染模式（預設 True Color）
var RenderMode = terminal.ModeTrueColor

// DefaultMaxTokens 預設 token 上限
const DefaultMaxTokens = 200000

// ContextData 包含 context 分析的完整結果
type ContextData struct {
	Formatted  string // 格式化的進度條字串（向後相容）
	Bar        string // 僅進度條部分（含前導分隔符，如 " | ░░░░░░░░░░"）
	Info       string // 百分比 + token 數（如 " 74% 148k"）
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
	return buildContextData(contextLength, maxTokens)
}

// AnalyzeDetailed 返回詳細的 context 資料（從檔案路徑讀取）
func AnalyzeDetailed(transcriptPath string, maxTokens int) *ContextData {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	var contextLength int
	if transcriptPath != "" {
		contextLength = calculateUsage(transcriptPath)
	}
	return buildContextData(contextLength, maxTokens)
}

// FormatContextParts 將 context 分為進度條和資訊兩部分，讓呼叫端分別設定截斷優先級。
// bar 為含前導分隔符的進度條（如 " | ░░░░░░░░░░"）；
// info 為百分比和 token 數（如 " 74% 148k"，帶顏色）。
func FormatContextParts(contextLength, maxTokens int) (bar string, info string) {
	percentage := int(float64(contextLength) * 100.0 / float64(maxTokens))
	if percentage > 100 {
		percentage = 100
	}

	progressBar := generateProgressBar(percentage)
	formattedNum := formatNumber(contextLength)

	bar = " | " + progressBar

	if RenderMode == terminal.ModeASCII {
		info = fmt.Sprintf(" %d%% %s", percentage, formattedNum)
	} else {
		color := getColor(percentage)
		info = fmt.Sprintf(" %s%d%% %s%s", color, percentage, formattedNum, statusline.ColorReset)
	}
	return bar, info
}

// buildContextData 從 contextLength 和 maxTokens 建立完整的 ContextData。
func buildContextData(contextLength, maxTokens int) *ContextData {
	percentage := int(float64(contextLength) * 100.0 / float64(maxTokens))
	if percentage > 100 {
		percentage = 100
	}

	bar, info := FormatContextParts(contextLength, maxTokens)
	return &ContextData{
		Formatted:  bar + info,
		Bar:        bar,
		Info:       info,
		Percentage: percentage,
		Tokens:     contextLength,
	}
}

func formatContext(contextLength, maxTokens int) string {
	bar, info := FormatContextParts(contextLength, maxTokens)
	return bar + info
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

// gradientColors 定義 10 格漸層色（綠→黃→橙→紅）
var gradientColors = [10]string{
	"\033[38;2;76;175;80m",  // green
	"\033[38;2;108;175;72m", // green-yellow
	"\033[38;2;139;175;64m", // yellow-green
	"\033[38;2;171;175;56m", // yellow
	"\033[38;2;202;165;48m", // gold
	"\033[38;2;224;150;40m", // gold-orange
	"\033[38;2;234;130;36m", // orange
	"\033[38;2;244;110;32m", // orange-red
	"\033[38;2;244;80;30m",  // red-orange
	"\033[38;2;244;67;54m",  // red
}

// generateProgressBar 生成進度條，根據 RenderMode 選擇渲染方式
func generateProgressBar(percentage int) string {
	switch RenderMode {
	case terminal.ModeASCII:
		return generateProgressBarASCII(percentage)
	default:
		return generateProgressBarTrueColor(percentage)
	}
}

// generateProgressBarTrueColor 生成漸層色進度條（True Color / 256 色）
func generateProgressBarTrueColor(percentage int) string {
	width := 10
	filled := percentage * width / 100
	if filled > width {
		filled = width
	}

	empty := width - filled

	var bar strings.Builder

	// 填充部分：每格獨立漸層色
	for i := 0; i < filled; i++ {
		bar.WriteString(gradientColors[i])
		bar.WriteString("█")
	}
	if filled > 0 {
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

// generateProgressBarASCII 生成純 ASCII 進度條
func generateProgressBarASCII(percentage int) string {
	width := 10
	filled := percentage * width / 100
	if filled > width {
		filled = width
	}
	empty := width - filled

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", empty) + "]"
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
