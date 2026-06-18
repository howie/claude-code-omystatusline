package terminal

import (
	"os"
	"strconv"
	"strings"
)

// RenderMode 終端渲染模式
type RenderMode int

const (
	// ModeTrueColor 24-bit RGB 真彩色
	ModeTrueColor RenderMode = iota
	// Mode256Color 256 色模式
	Mode256Color
	// ModeASCII 純 ASCII 模式
	ModeASCII
)

// Detect 偵測終端的色彩能力
// 優先檢查 CLAUDE_STATUSLINE_ASCII 環境變數，
// 再檢查 COLORTERM 和 TERM 判斷色彩支援程度
func Detect() RenderMode {
	// 使用者強制 ASCII
	if os.Getenv("CLAUDE_STATUSLINE_ASCII") == "1" {
		return ModeASCII
	}

	// 檢查 True Color 支援
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return ModeTrueColor
	}

	// 檢查 TERM
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		return Mode256Color
	}

	// 大多數現代終端支援 True Color，預設使用
	if term != "" && term != "dumb" {
		return ModeTrueColor
	}

	return ModeASCII
}

// Width 回傳終端的欄位寬度。
// 優先用平台原生查詢（Unix: ioctl TIOCGWINSZ；Windows: 不支援，直接 fallback），
// 其次讀 COLUMNS 環境變數，最後預設 120。
// 平台相關的 ioctl 實作見 width_unix.go / width_windows.go（以 build tag 分流，
// 讓 GOOS=windows 交叉編譯不會碰到 Unix-only 的 syscall.SYS_IOCTL / TIOCGWINSZ）。
func Width() int {
	if w, ok := terminalWidthFromIoctl(); ok {
		return w
	}

	// Fallback: COLUMNS 環境變數（Claude Code 在 statusline context 下會設定此變數）
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}

	return 120
}
