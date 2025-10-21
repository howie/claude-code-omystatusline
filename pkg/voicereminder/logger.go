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

	// 新路徑優先
	newLogPath := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "data", "voice-reminder-debug.log")
	// 舊路徑作為備援
	oldLogPath := filepath.Join(homeDir, ".claude", "voice-reminder-debug.log")

	var logPath string

	// 確保新路徑的目錄存在
	newLogDir := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "data")
	if err := os.MkdirAll(newLogDir, 0755); err == nil {
		logPath = newLogPath
	} else {
		// 回退到舊路徑
		logPath = oldLogPath
	}

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
	_, _ = l.logFile.WriteString(msg)
}

// Close 關閉日誌檔案
func (l *Logger) Close() {
	if l.enabled && l.logFile != nil {
		_ = l.logFile.Close()
	}
}
