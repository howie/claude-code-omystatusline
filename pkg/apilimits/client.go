package apilimits

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// APILimitsInfo 代表 API 配額使用資訊
type APILimitsInfo struct {
	FiveHourPct   int    // 5 小時窗口使用百分比
	SevenDayPct   int    // 7 天窗口使用百分比
	FiveHourReset string // 5 小時窗口重置倒計時
	SevenDayReset string // 7 天窗口重置倒計時
	LimitReached  bool   // 是否已達上限
}

// 快取
var (
	limitsCache   *APILimitsInfo
	limitsExpires time.Time
	limitsMutex   sync.RWMutex
)

const (
	cacheTTL       = 5 * time.Minute
	errorCacheTTL  = 15 * time.Second
	requestTimeout = 2 * time.Second
	usageAPIURL    = "https://api.anthropic.com/api/oauth/usage"
)

// apiResponse 代表 API 回應結構
type apiResponse struct {
	FiveHour struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"five_hour"`
	SevenDay struct {
		Utilization float64 `json:"utilization"`
		ResetsAt    string  `json:"resets_at"`
	} `json:"seven_day"`
}

// cachedData 用於檔案快取
type cachedData struct {
	Info      APILimitsInfo `json:"info"`
	ExpiresAt int64         `json:"expires_at"`
}

// Fetch 取得 API 使用量資訊
func Fetch() *APILimitsInfo {
	// 檢查記憶體快取
	limitsMutex.RLock()
	if time.Now().Before(limitsExpires) && limitsCache != nil {
		result := limitsCache
		limitsMutex.RUnlock()
		return result
	}
	limitsMutex.RUnlock()

	// 檢查檔案快取
	if cached := loadFileCache(); cached != nil {
		updateMemoryCache(cached, cacheTTL)
		return cached
	}

	// 取得 OAuth token
	token := getOAuthToken()
	if token == "" {
		return nil
	}

	// 呼叫 API
	info := fetchFromAPI(token)
	if info == nil {
		// 錯誤時使用短快取
		updateMemoryCache(nil, errorCacheTTL)
		return nil
	}

	// 更新快取
	updateMemoryCache(info, cacheTTL)
	saveFileCache(info)

	return info
}

func fetchFromAPI(token string) *APILimitsInfo {
	client := &http.Client{Timeout: requestTimeout}

	req, err := http.NewRequest("GET", usageAPIURL, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil
	}

	info := &APILimitsInfo{
		FiveHourPct:   int(apiResp.FiveHour.Utilization * 100),
		SevenDayPct:   int(apiResp.SevenDay.Utilization * 100),
		FiveHourReset: formatResetTime(apiResp.FiveHour.ResetsAt),
		SevenDayReset: formatResetTime(apiResp.SevenDay.ResetsAt),
	}

	if info.FiveHourPct >= 100 || info.SevenDayPct >= 100 {
		info.LimitReached = true
	}

	return info
}

func formatResetTime(resetsAt string) string {
	if resetsAt == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, resetsAt)
	if err != nil {
		return ""
	}

	remaining := time.Until(t)
	if remaining <= 0 {
		return "now"
	}

	if remaining < time.Hour {
		return fmt.Sprintf("%dm", int(remaining.Minutes()))
	}
	if remaining < 24*time.Hour {
		h := int(remaining.Hours())
		m := int(remaining.Minutes()) % 60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dd", int(remaining.Hours()/24))
}

func getOAuthToken() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// 嘗試從 Claude Code credentials 讀取
	credPath := filepath.Join(homeDir, ".claude", ".credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return ""
	}

	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return ""
	}

	// 嘗試多種 token 欄位
	for _, key := range []string{"accessToken", "access_token", "oauthToken", "oauth_token"} {
		if token, ok := creds[key].(string); ok && token != "" {
			return token
		}
	}

	return ""
}

func updateMemoryCache(info *APILimitsInfo, ttl time.Duration) {
	limitsMutex.Lock()
	limitsCache = info
	limitsExpires = time.Now().Add(ttl)
	limitsMutex.Unlock()
}

func getCachePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "omystatusline", "cache", "api-limits.json")
}

func loadFileCache() *APILimitsInfo {
	path := getCachePath()
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cached cachedData
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil
	}

	if time.Now().Unix() > cached.ExpiresAt {
		return nil
	}

	return &cached.Info
}

func saveFileCache(info *APILimitsInfo) {
	path := getCachePath()
	if path == "" {
		return
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	cached := cachedData{
		Info:      *info,
		ExpiresAt: time.Now().Add(cacheTTL).Unix(),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return
	}

	_ = os.WriteFile(path, data, 0644)
}

// Format 格式化 API 配額為顯示字串
func Format(info *APILimitsInfo) string {
	if info == nil {
		return ""
	}

	result := fmt.Sprintf("5h: %d%%", info.FiveHourPct)
	if info.FiveHourReset != "" {
		result += fmt.Sprintf(" (%s)", info.FiveHourReset)
	}

	result += fmt.Sprintf(" | 7d: %d%%", info.SevenDayPct)
	if info.SevenDayReset != "" {
		result += fmt.Sprintf(" (%s)", info.SevenDayReset)
	}

	if info.LimitReached {
		result += " ⚠ Limit reached"
	}

	return result
}
