package voicereminder

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Logger Debug 日誌記錄器
type Logger struct {
	enabled bool
	logFile *os.File
}

// NewLogger 創建日誌記錄器
func NewLogger(debugMode bool) *Logger {
	if !debugMode {
		return &Logger{enabled: false}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Logger{enabled: false}
	}

	logPath := filepath.Join(homeDir, ".claude", "voice-reminder-debug.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &Logger{enabled: false}
	}

	return &Logger{
		enabled: true,
		logFile: logFile,
	}
}

// Log 記錄日誌訊息
func (l *Logger) Log(format string, args ...interface{}) {
	if !l.enabled {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
	l.logFile.WriteString(msg)
}

// Close 關閉日誌檔案
func (l *Logger) Close() {
	if l.enabled && l.logFile != nil {
		l.logFile.Close()
	}
}
