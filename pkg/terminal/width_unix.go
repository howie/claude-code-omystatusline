//go:build !windows

package terminal

import (
	"syscall"
	"unsafe"
)

// winsize 是 TIOCGWINSZ ioctl 的回傳結構
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// terminalWidthFromIoctl 嘗試從 stdout/stderr/stdin 透過 ioctl TIOCGWINSZ 取得終端欄位寬度。
// 回傳 (width, true) 代表成功；(0, false) 代表無法取得（呼叫端會 fallback 到 COLUMNS）。
func terminalWidthFromIoctl() (int, bool) {
	for _, fd := range []uintptr{1, 2, 0} {
		var ws winsize
		if _, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			fd,
			syscall.TIOCGWINSZ,
			uintptr(unsafe.Pointer(&ws)),
		); errno == 0 && ws.Col > 0 {
			return int(ws.Col), true
		}
	}
	return 0, false
}
