package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session 資料結構
type Session struct {
	ID            string     `json:"id"`
	Date          string     `json:"date"`
	Start         int64      `json:"start"`
	LastHeartbeat int64      `json:"last_heartbeat"`
	TotalSeconds  int64      `json:"total_seconds"`
	Intervals     []Interval `json:"intervals"`
}

type Interval struct {
	Start int64  `json:"start"`
	End   *int64 `json:"end"`
}

// Update 更新 Session
func Update(sessionID string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	sessionsDir := filepath.Join(homeDir, ".claude", "session-tracker", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return
	}

	sessionFile := filepath.Join(sessionsDir, sessionID+".json")
	currentTime := time.Now().Unix()
	today := time.Now().Format("2006-01-02")

	var session Session

	// 讀取現有 session
	if data, err := os.ReadFile(sessionFile); err == nil {
		_ = json.Unmarshal(data, &session) // Ignore error, use zero value if unmarshal fails
	} else {
		// 新 session
		session = Session{
			ID:            sessionID,
			Date:          today,
			Start:         currentTime,
			LastHeartbeat: currentTime,
			TotalSeconds:  0,
			Intervals:     []Interval{{Start: currentTime, End: nil}},
		}
	}

	// 若 session 檔案是舊日期，歸零改為今日的新區段
	if session.Date != "" && session.Date != today {
		session.Date = today
		session.Start = currentTime
		session.LastHeartbeat = currentTime
		session.TotalSeconds = 0
		session.Intervals = []Interval{{Start: currentTime, End: nil}}
	}

	if session.Date == "" {
		session.Date = today
	}

	// 更新心跳
	gap := currentTime - session.LastHeartbeat
	session.LastHeartbeat = currentTime

	if gap < 600 { // 10分鐘內為連續
		// 延伸當前區間
		if len(session.Intervals) > 0 {
			session.Intervals[len(session.Intervals)-1].End = &currentTime
		}
	} else {
		// 新增新區間
		session.Intervals = append(session.Intervals, Interval{
			Start: currentTime,
			End:   &currentTime,
		})
	}

	// 計算總時數
	var total int64
	for _, interval := range session.Intervals {
		if interval.End != nil {
			total += *interval.End - interval.Start
		}
	}
	session.TotalSeconds = total

	// 儲存
	if data, err := json.Marshal(session); err == nil {
		_ = os.WriteFile(sessionFile, data, 0644) // Ignore error, session tracking is non-critical
	}
}

// CalculateTotalHours 計算總時數
func CalculateTotalHours(currentSessionID string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "0m"
	}

	sessionsDir := filepath.Join(homeDir, ".claude", "session-tracker", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return "0m"
	}

	var totalSeconds int64
	activeSessions := 0
	today := time.Now().Format("2006-01-02")
	currentTime := time.Now().Unix()

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		sessionFile := filepath.Join(sessionsDir, entry.Name())
		data, err := os.ReadFile(sessionFile)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		// 只計算今日的 session
		if session.Date == today {
			totalSeconds += session.TotalSeconds

			// 檢查是否活躍（10分鐘內有心跳）
			if currentTime-session.LastHeartbeat < 600 {
				activeSessions++
			}
		}
	}

	// 格式化輸出
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60

	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dh", hours)
		if minutes > 0 {
			timeStr += fmt.Sprintf("%dm", minutes)
		}
	} else {
		timeStr = fmt.Sprintf("%dm", minutes)
	}

	if activeSessions > 1 {
		return fmt.Sprintf("%s [%d sessions]", timeStr, activeSessions)
	}
	return timeStr
}
