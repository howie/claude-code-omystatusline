//go:build windows

package terminal

// terminalWidthFromIoctl is a no-op on Windows: there is no TIOCGWINSZ ioctl, and the Unix-only
// syscall.SYS_IOCTL / syscall.TIOCGWINSZ symbols do not exist there (cross-compiling with them
// breaks the build). Return (0, false) so Width() falls back to the COLUMNS environment variable,
// which Claude Code sets before running the status line command.
func terminalWidthFromIoctl() (int, bool) {
	return 0, false
}
