package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ANSI 顏色定義
const (
	ColorReset  = "\033[0m"
	ColorGold   = "\033[38;2;195;158;83m"
	ColorCyan   = "\033[38;2;118;170;185m"
	ColorPink   = "\033[38;2;255;182;193m"
	ColorGreen  = "\033[38;2;152;195;121m"
	ColorGray   = "\033[38;2;64;64;64m"
	ColorSilver = "\033[38;2;192;192;192m"

	ColorCtxGreen = "\033[38;2;108;167;108m"
	ColorCtxGold  = "\033[38;2;188;155;83m"
	ColorCtxRed   = "\033[38;2;185;102;82m"
)

// 模型圖示和顏色
var modelConfig = map[string][2]string{
	"Opus":   {ColorGold, "💛"},
	"Sonnet": {ColorCyan, "💠"},
	"Haiku":  {ColorPink, "🌸"},
}

// 輸入資料結構
type Input struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	SessionID string `json:"session_id"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	TranscriptPath string `json:"transcript_path,omitempty"`
}

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

// 結果通道資料
type Result struct {
	Type string
	Data interface{}
}

// 簡單快取
var (
	gitBranchCache   string
	gitBranchExpires time.Time
	cacheMutex       sync.RWMutex
)

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode input: %v\n", err)
		os.Exit(1)
	}

	// 建立結果通道
	results := make(chan Result, 4)
	var wg sync.WaitGroup

	// 並行獲取各種資訊
	wg.Add(4)

	go func() {
		defer wg.Done()
		branch := getGitBranch()
		results <- Result{"git", branch}
	}()

	go func() {
		defer wg.Done()
		totalHours := calculateTotalHours(input.SessionID)
		results <- Result{"hours", totalHours}
	}()

	go func() {
		defer wg.Done()
		contextInfo := analyzeContext(input.TranscriptPath)
		results <- Result{"context", contextInfo}
	}()

	go func() {
		defer wg.Done()
		userMsg := extractUserMessage(input.TranscriptPath, input.SessionID)
		results <- Result{"message", userMsg}
	}()

	// 等待所有 goroutines 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集結果
	var gitBranch, totalHours, contextUsage, userMessage string

	for result := range results {
		switch result.Type {
		case "git":
			gitBranch = result.Data.(string)
		case "hours":
			totalHours = result.Data.(string)
		case "context":
			contextUsage = result.Data.(string)
		case "message":
			userMessage = result.Data.(string)
		}
	}

	// 更新 session（同步操作，避免競爭條件）
	updateSession(input.SessionID)

	// 格式化模型顯示
	modelDisplay := formatModel(input.Model.DisplayName)
	projectName := filepath.Base(input.Workspace.CurrentDir)

	// 輸出狀態列
	fmt.Printf("%s[%s] 📂 %s%s%s | %s%s\n",
		ColorReset, modelDisplay, projectName, gitBranch,
		contextUsage, totalHours, ColorReset)

	// 輸出使用者訊息
	if userMessage != "" {
		fmt.Print(userMessage)
	}
}

// 格式化模型顯示
func formatModel(model string) string {
	for key, config := range modelConfig {
		if strings.Contains(model, key) {
			color := config[0]
			icon := config[1]
			return fmt.Sprintf("%s%s %s%s", color, icon, model, ColorReset)
		}
	}
	return model
}

// 獲取 Git 分支（帶快取）
func getGitBranch() string {
	cacheMutex.RLock()
	if time.Now().Before(gitBranchExpires) && gitBranchCache != "" {
		result := gitBranchCache
		cacheMutex.RUnlock()
		return result
	}
	cacheMutex.RUnlock()

	// 檢查是否為 Git 倉庫
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// 嘗試找到 Git 根目錄
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		if err := cmd.Run(); err != nil {
			return ""
		}
	}

	// 獲取當前分支
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return ""
	}

	// 檢測是否在 worktree 中
	icon := "⚡"
	worktreeLabel := ""
	gitDirCmd := exec.Command("git", "rev-parse", "--git-dir")
	gitDirOutput, err1 := gitDirCmd.Output()

	gitCommonDirCmd := exec.Command("git", "rev-parse", "--git-common-dir")
	gitCommonDirOutput, err2 := gitCommonDirCmd.Output()

	if err1 == nil && err2 == nil {
		gitDir := strings.TrimSpace(string(gitDirOutput))
		gitCommonDir := strings.TrimSpace(string(gitCommonDirOutput))

		// 如果 git-dir 和 git-common-dir 不同，表示在 worktree 中
		if gitDir != gitCommonDir {
			icon = "🔀"
			worktreeLabel = " (worktree)"
		}
	}

	result := fmt.Sprintf(" %s %s%s", icon, branch, worktreeLabel)

	// 更新快取
	cacheMutex.Lock()
	gitBranchCache = result
	gitBranchExpires = time.Now().Add(5 * time.Second)
	cacheMutex.Unlock()

	return result
}

// 更新 Session
func updateSession(sessionID string) {
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

// 計算總時數
func calculateTotalHours(currentSessionID string) string {
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

// 分析 Context 使用量
func analyzeContext(transcriptPath string) string {
	var contextLength int

	if transcriptPath == "" {
		// 當 transcriptPath 為空時（對話剛開始），顯示初始狀態
		contextLength = 0
	} else {
		contextLength = calculateContextUsage(transcriptPath)
	}

	// 即使 contextLength 為 0 也顯示進度條

	// 計算百分比（基於 200k tokens）
	percentage := int(float64(contextLength) * 100.0 / 200000.0)
	if percentage > 100 {
		percentage = 100
	}

	// 生成進度條
	progressBar := generateProgressBar(percentage)
	formattedNum := formatNumber(contextLength)
	color := getContextColor(percentage)

	return fmt.Sprintf(" | %s %s%d%% %s%s",
		progressBar, color, percentage, formattedNum, ColorReset)
}

// 計算 Context 使用量
func calculateContextUsage(transcriptPath string) int {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	// 讀取最後100行
	scanner := bufio.NewScanner(file)

	// 設定更大的 buffer（1MB）以處理長 JSON 行
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, 0, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	// 先讀取所有行到切片
	allLines := make([]string, 0)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// 取最後100行
	start := len(allLines) - 100
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	// 從後往前分析
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// 空行跳過
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 先嘗試解析 JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// 檢查 isSidechain 欄位（處理 bool 和可能的其他類型）
		if sidechain, ok := data["isSidechain"]; ok {
			// 如果是 sidechain，跳過
			if isSide, ok := sidechain.(bool); ok && isSide {
				continue
			}
		}

		// 檢查並提取 usage 資料
		if message, ok := data["message"].(map[string]interface{}); ok {
			if usage, ok := message["usage"].(map[string]interface{}); ok {
				var total float64

				// 計算所有 token 類型
				if input, ok := usage["input_tokens"].(float64); ok {
					total += input
				}
				if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
					total += cacheRead
				}
				if cacheCreation, ok := usage["cache_creation_input_tokens"].(float64); ok {
					total += cacheCreation
				}

				// 如果找到有效的 token 數量，立即返回
				if total > 0 {
					return int(total)
				}
			}
		}
	}

	return 0
}

// 生成進度條
func generateProgressBar(percentage int) string {
	width := 10
	filled := percentage * width / 100
	if filled > width {
		filled = width
	}

	empty := width - filled
	color := getContextColor(percentage)

	var bar strings.Builder

	// 填充部分
	if filled > 0 {
		bar.WriteString(color)
		bar.WriteString(strings.Repeat("█", filled))
		bar.WriteString(ColorReset)
	}

	// 空白部分
	if empty > 0 {
		bar.WriteString(ColorGray)
		bar.WriteString(strings.Repeat("░", empty))
		bar.WriteString(ColorReset)
	}

	return bar.String()
}

// 獲取 Context 顏色
func getContextColor(percentage int) string {
	if percentage < 60 {
		return ColorCtxGreen
	} else if percentage < 80 {
		return ColorCtxGold
	}
	return ColorCtxRed
}

// 格式化數字
func formatNumber(num int) string {
	if num == 0 {
		return "--"
	}

	if num >= 1000000 {
		return fmt.Sprintf("%dM", num/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%dk", num/1000)
	}
	return strconv.Itoa(num)
}

// 提取使用者訊息
func extractUserMessage(transcriptPath, sessionID string) string {
	if transcriptPath == "" {
		return ""
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// 讀取最後200行
	scanner := bufio.NewScanner(file)

	allLines := make([]string, 0)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := len(allLines) - 200
	if start < 0 {
		start = 0
	}
	lines := allLines[start:]

	// 從後往前搜尋使用者訊息
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// 檢查是否為當前 session 的使用者訊息
		isSidechain, _ := data["isSidechain"].(bool)
		sessionMatch := false
		if sid, ok := data["sessionId"].(string); ok && sid == sessionID {
			sessionMatch = true
		}

		if !isSidechain && sessionMatch {
			if message, ok := data["message"].(map[string]interface{}); ok {
				role, _ := message["role"].(string)
				msgType, _ := data["type"].(string)

				if role == "user" && msgType == "user" {
					if content, ok := message["content"].(string); ok {
						// 過濾系統訊息
						if isSystemMessage(content) {
							continue
						}

						// 格式化並返回
						return formatUserMessage(content)
					}
				}
			}
		}
	}

	return ""
}

// 檢查是否為系統訊息
func isSystemMessage(content string) bool {
	// 過濾 JSON 格式
	if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
		return true
	}
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return true
	}

	// 過濾 XML 標籤
	xmlTags := []string{
		"<local-command-stdout>", "<command-name>",
		"<command-message>", "<command-args>",
	}
	for _, tag := range xmlTags {
		if strings.Contains(content, tag) {
			return true
		}
	}

	// 過濾 Caveat 訊息
	if strings.HasPrefix(content, "Caveat:") {
		return true
	}

	return false
}

// 格式化使用者訊息
func formatUserMessage(message string) string {
	if message == "" {
		return ""
	}

	maxLines := 3
	lineWidth := 80

	lines := strings.Split(message, "\n")
	var result []string

	for i, line := range lines {
		if i >= maxLines {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) > lineWidth {
			line = line[:lineWidth-3] + "..."
		}

		result = append(result, fmt.Sprintf("%s｜%s%s%s",
			ColorReset, ColorGreen, line, ColorReset))
	}

	if len(lines) > maxLines {
		result = append(result, fmt.Sprintf("%s｜... (還有 %d 行)%s",
			ColorReset, len(lines)-maxLines, ColorReset))
	}

	if len(result) > 0 {
		return strings.Join(result, "\n") + "\n"
	}

	return ""
}
