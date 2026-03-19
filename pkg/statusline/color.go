package statusline

// ANSI 顏色定義
const (
	ColorReset  = "\033[0m"
	ColorGold   = "\033[38;2;195;158;83m"
	ColorCyan   = "\033[38;2;118;170;185m"
	ColorPink   = "\033[38;2;255;182;193m"
	ColorGreen  = "\033[38;2;152;195;121m"
	ColorGray   = "\033[38;2;64;64;64m"
	ColorSilver = "\033[38;2;192;192;192m"

	ColorCtxGreen = "\033[38;2;108;167;108m"
	ColorCtxGold  = "\033[38;2;188;155;83m"
	ColorCtxRed   = "\033[38;2;185;102;82m"

	// 新增顏色（claude-hud 風格）
	ColorDim        = "\033[2m"
	ColorYellow     = "\033[38;2;229;192;123m"
	ColorMagenta    = "\033[38;2;198;120;221m"
	ColorBlue       = "\033[38;2;97;175;239m"
	ColorBrightBlue = "\033[38;2;130;170;255m"
	ColorRed        = "\033[38;2;224;108;117m"
	ColorWhite      = "\033[38;2;220;220;220m"

	// API 配額顏色
	ColorQuotaOk      = "\033[38;2;130;170;255m" // bright blue <75%
	ColorQuotaWarning = "\033[38;2;198;120;221m" // bright magenta 75-90%
	ColorQuotaCrit    = "\033[38;2;224;108;117m" // red 90%+
)
