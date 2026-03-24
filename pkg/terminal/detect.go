package terminal

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
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

// winsize 是 TIOCGWINSZ ioctl 的回傳結構
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// Width 回傳終端的欄位寬度。
// 優先用 ioctl TIOCGWINSZ，其次讀 COLUMNS 環境變數，最後預設 120。
func Width() int {
	// 嘗試從 stdout/stderr/stdin 取得終端寬度
	for _, fd := range []uintptr{1, 2, 0} {
		var ws winsize
		if _, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			fd,
			syscall.TIOCGWINSZ,
			uintptr(unsafe.Pointer(&ws)),
		); errno == 0 && ws.Col > 0 {
			return int(ws.Col)
		}
	}

	// Fallback: COLUMNS 環境變數
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}

	return 120
}
