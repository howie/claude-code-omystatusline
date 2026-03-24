package terminal

import (
	"os"
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
