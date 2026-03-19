package speed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// SpeedInfo 代表輸出速度資訊
type SpeedInfo struct {
	TokensPerSec int
}

// Measurement 快取的測量值
type Measurement struct {
	OutputTokens int   `json:"output_tokens"`
	TimestampMs  int64 `json:"timestamp_ms"`
}

// Calculate 計算輸出速度（tokens/sec）
func Calculate(lines []transcript.Line, sessionID string) *SpeedInfo {
	// 從 transcript 中找出最新的 output_tokens
	currentTokens := extractOutputTokens(lines)
	if currentTokens <= 0 {
		return nil
	}

	now := time.Now().UnixMilli()

	// 讀取上次測量值
	prev := loadMeasurement(sessionID)

	// 儲存當前測量值
	saveMeasurement(sessionID, &Measurement{
		OutputTokens: currentTokens,
		TimestampMs:  now,
	})

	if prev == nil || prev.OutputTokens <= 0 {
		return nil
	}

	// 計算 delta
	tokenDelta := currentTokens - prev.OutputTokens
	timeDeltaMs := now - prev.TimestampMs

	// 只在 2 秒內的測量才計算速度
	if timeDeltaMs <= 0 || timeDeltaMs > 2000 || tokenDelta <= 0 {
		return nil
	}

	tokPerSec := int(float64(tokenDelta) * 1000.0 / float64(timeDeltaMs))
	if tokPerSec <= 0 {
		return nil
	}

	return &SpeedInfo{TokensPerSec: tokPerSec}
}

func extractOutputTokens(lines []transcript.Line) int {
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		msg, ok := l.Parsed["message"].(map[string]interface{})
		if !ok {
			continue
		}

		usage, ok := msg["usage"].(map[string]interface{})
		if !ok {
			continue
		}

		if output, ok := usage["output_tokens"].(float64); ok && output > 0 {
			return int(output)
		}
	}
	return 0
}

func getCachePath(sessionID string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "omystatusline", "cache", fmt.Sprintf("speed-%s.json", sessionID))
}

func loadMeasurement(sessionID string) *Measurement {
	path := getCachePath(sessionID)
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var m Measurement
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	return &m
}

func saveMeasurement(sessionID string, m *Measurement) {
	path := getCachePath(sessionID)
	if path == "" {
		return
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	data, err := json.Marshal(m)
	if err != nil {
		return
	}

	_ = os.WriteFile(path, data, 0644)
}

// Format 格式化速度顯示
func Format(info *SpeedInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("%d tok/s", info.TokensPerSec)
}
