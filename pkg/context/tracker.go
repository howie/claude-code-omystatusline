package context

import (
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
	Bar  string // 僅進度條部分（含前導分隔符，如 " | ░░░░░░░░░░"）
	Info string // 百分比 + token 數（如 " 74% 148k"）；NoUsageData 時為 " [remote]"（ASCII 模式）或帶 ANSI dim 的 " 📡"（True Color 模式）

	// Percentage 為百分比數值（0–100）。NoUsageData 為 true 時固定為 0，但不代表真正零使用量。
	Percentage int
	// Tokens 為最近一次 message.usage 解析到的 token 數。NoUsageData 為 true 時固定為 0，但不代表真正零使用量。
	// 使用 HasData() 確認數值是否有意義。
	Tokens int
	// NoUsageData 為 true 代表 lines 有可解析的項目但無任何 message.usage，
	// 例如 local-agent-mode 的純 metadata session（只含 custom-title、agent-name、pr-link 等）。
	// 此時 Tokens 與 Percentage 均為 0，但不代表真正的零使用量。
	NoUsageData bool
}

// HasData 回報此 ContextData 是否包含真實的 token 使用量。
// 回傳 false 的情況有兩種，語意不同：
//   - NoUsageData=true：transcript 可解析但無任何 message.usage（metadata-only session）
//   - NoUsageData=false, Tokens=0：transcript 不存在、空白、全為解析失敗行，或 session 尚無 assistant reply
//
// 若需要區分兩種「無資料」的原因，直接檢查 NoUsageData 欄位。
func (c *ContextData) HasData() bool {
	return !c.NoUsageData && c.Tokens > 0
}

// Analyze 分析 Context 使用量（向後相容）
// maxTokens <= 0 時使用 DefaultMaxTokens
// 讀取失敗時輸出錯誤訊息至 stderr 並回傳零值進度條。
func Analyze(transcriptPath string, maxTokens int) string {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	if transcriptPath == "" {
		return formatContext(0, maxTokens)
	}
	lines, err := transcript.ReadTail(transcriptPath, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: could not read transcript %q: %v\n", transcriptPath, err)
		return formatContext(0, maxTokens)
	}
	return formatContext(calculateUsageFromLines(lines), maxTokens)
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
	noUsageData := contextLength == 0 && isMetadataOnlyTranscript(lines)
	return buildContextData(contextLength, maxTokens, noUsageData)
}

// AnalyzeDetailed 返回詳細的 context 資料（從檔案路徑讀取）。
// 空路徑或讀取失敗時回傳 NoUsageData=false（無法判斷）；
// 成功讀取後，若 transcript 為純 metadata，NoUsageData 可能為 true。
func AnalyzeDetailed(transcriptPath string, maxTokens int) *ContextData {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	if transcriptPath == "" {
		return buildContextData(0, maxTokens, false)
	}
	lines, err := transcript.ReadTail(transcriptPath, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: could not read transcript %q: %v\n", transcriptPath, err)
		return buildContextData(0, maxTokens, false)
	}
	contextLength := calculateUsageFromLines(lines)
	noUsageData := contextLength == 0 && isMetadataOnlyTranscript(lines)
	return buildContextData(contextLength, maxTokens, noUsageData)
}

// FormatContextParts 將 context 分為進度條和資訊兩部分，讓呼叫端分別設定截斷優先級。
// bar 為含前導分隔符的進度條（如 " | ░░░░░░░░░░"）；
// info 為百分比和 token 數（如 " 74% 148k"，帶顏色）。
func FormatContextParts(contextLength, maxTokens int) (bar string, info string) {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
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
// noUsageData 為 true 時代表呼叫端確認 lines 有可解析內容但無任何 message.usage
// （例如 local-agent-mode 的純 metadata session），此時以 📡（True Color 模式）或
// "[remote]"（ASCII 模式）替代百分比顯示，避免誤導為「真的 0%」。
func buildContextData(contextLength, maxTokens int, noUsageData bool) *ContextData {
	if noUsageData {
		bar := " | " + generateProgressBar(0)
		var info string
		if RenderMode == terminal.ModeASCII {
			info = " [remote]"
		} else {
			info = fmt.Sprintf(" %s📡%s", statusline.ColorDim, statusline.ColorReset)
		}
		return &ContextData{Bar: bar, Info: info, NoUsageData: true}
	}

	percentage := int(float64(contextLength) * 100.0 / float64(maxTokens))
	if percentage > 100 {
		percentage = 100
	}

	bar, info := FormatContextParts(contextLength, maxTokens)
	return &ContextData{Bar: bar, Info: info, Percentage: percentage, Tokens: contextLength}
}

// isMetadataOnlyTranscript 當 lines 含有至少一個可解析的項目，且沒有任何一行帶 "message" 欄位時回傳 true。
// 這代表 transcript 只有管理用的 metadata 事件（如 custom-title、agent-name、pr-link），
// 而非真實對話內容，常見於 local-agent-mode session。
// Raw 行（JSON 解析失敗，l.Parsed == nil）不計入；如全為解析失敗行，則回傳 false。
// 帶有 "message": null 的行也視為「有 message 欄位」而不計入 metadata，這是有意設計。
func isMetadataOnlyTranscript(lines []transcript.Line) bool {
	hasParsed := false
	for _, l := range lines {
		if l.Parsed == nil {
			continue
		}
		hasParsed = true
		if _, hasMsg := l.Parsed["message"]; hasMsg {
			return false
		}
	}
	return hasParsed
}

func formatContext(contextLength, maxTokens int) string {
	bar, info := FormatContextParts(contextLength, maxTokens)
	return bar + info
}

// usageFromLines 從 transcript 中找出最後一筆有效使用量的行，同時回傳 token 數與模型 ID。
// 「有效」定義：input+cache token 總和 > 0，確保 model 與 token 數永遠來自同一行，
// 避免 output-only 行（total=0）影響 model 判斷而 token 數卻取自更早的行。
// sidechain 行與 nil Parsed 行一律跳過。lines 為 nil 時回傳 (0, "")。
func usageFromLines(lines []transcript.Line) (tokens int, modelID string) {
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}
		if isSide, ok := l.Parsed["isSidechain"].(bool); ok && isSide {
			continue
		}
		message, ok := l.Parsed["message"].(map[string]interface{})
		if !ok {
			continue
		}
		usage, ok := message["usage"].(map[string]interface{})
		if !ok {
			continue
		}
		var total float64
		if v, ok := usage["input_tokens"].(float64); ok {
			total += v
		}
		if v, ok := usage["cache_read_input_tokens"].(float64); ok {
			total += v
		}
		if v, ok := usage["cache_creation_input_tokens"].(float64); ok {
			total += v
		}
		if total <= 0 {
			continue
		}
		mid, _ := message["model"].(string)
		return int(total), mid
	}
	return 0, ""
}

// InferModelFromLines 回傳 transcript 中最後一筆有效 usage 行的模型 ID。
// 「有效」與 calculateUsageFromLines 一致：input+cache token 總和 > 0。
// 用於 mixed-model session（例如 Plan 用 Opus、Edit 用 Sonnet）時，
// 確保 context window 分母與產生 token 數的模型來自同一行。
// 找不到時回傳空字串；呼叫端應 fallback 到 input.Model.ID。
func InferModelFromLines(lines []transcript.Line) string {
	_, mid := usageFromLines(lines)
	return mid
}

// calculateUsageFromLines 從共享 transcript 行計算 Context 使用量
func calculateUsageFromLines(lines []transcript.Line) int {
	tokens, _ := usageFromLines(lines)
	return tokens
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
