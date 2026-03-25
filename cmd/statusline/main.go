package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/howie/claude-code-omystatusline/pkg/agents"
	"github.com/howie/claude-code-omystatusline/pkg/apilimits"
	"github.com/howie/claude-code-omystatusline/pkg/config"
	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/git"
	"github.com/howie/claude-code-omystatusline/pkg/gitstatus"
	"github.com/howie/claude-code-omystatusline/pkg/session"
	"github.com/howie/claude-code-omystatusline/pkg/speed"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
	"github.com/howie/claude-code-omystatusline/pkg/todo"
	"github.com/howie/claude-code-omystatusline/pkg/tools"
	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func main() {
	var input statusline.Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode input: %v\n", err)
		os.Exit(1)
	}

	// Phase 1: 載入配置（同步，快速本地檔案讀取）
	cfg := config.Load()

	// 偵測終端渲染能力
	context.RenderMode = terminal.Detect()

	// 取得分隔符設定
	sep := cfg.GetSeparator()

	// 讀取環境變數 STATUSLINE_MAX_TOKENS
	maxTokens := 0
	if envMax := os.Getenv("STATUSLINE_MAX_TOKENS"); envMax != "" {
		if v, err := strconv.Atoi(envMax); err == nil && v > 0 {
			maxTokens = v
		}
	}

	// Phase 2: 讀取 transcript（一次 I/O）
	lines, _ := transcript.ReadTail(input.TranscriptPath, 200)

	// Phase 3: 並行處理所有資料收集
	results := make(chan statusline.Result, 12)
	var wg sync.WaitGroup

	// --- Transcript-based goroutines ---

	if cfg.Sections.Context {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var contextUsage string
			if lines != nil {
				contextUsage = context.AnalyzeFromLines(lines, maxTokens)
			} else {
				contextUsage = context.Analyze(input.TranscriptPath, maxTokens)
			}
			results <- statusline.Result{Type: "context", Data: contextUsage}
		}()
	}

	if cfg.Sections.UserMessage {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var userMsg string
			if lines != nil {
				userMsg = statusline.ExtractUserMessageFromLines(lines, input.SessionID)
			} else {
				userMsg = statusline.ExtractUserMessage(input.TranscriptPath, input.SessionID)
			}
			results <- statusline.Result{Type: "message", Data: userMsg}
		}()
	}

	if cfg.Sections.Tools {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil {
				results <- statusline.Result{Type: "tools", Data: ""}
				return
			}
			activeTools := tools.Analyze(lines)
			results <- statusline.Result{Type: "tools", Data: tools.Format(activeTools)}
		}()
	}

	if cfg.Sections.Agents {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil {
				results <- statusline.Result{Type: "agents", Data: ""}
				return
			}
			activeAgents := agents.Analyze(lines, input.SessionID)
			results <- statusline.Result{Type: "agents", Data: agents.Format(activeAgents)}
		}()
	}

	if cfg.Sections.Todo {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil {
				results <- statusline.Result{Type: "todo", Data: ""}
				return
			}
			todoInfo := todo.Analyze(lines)
			results <- statusline.Result{Type: "todo", Data: todo.Format(todoInfo)}
		}()
	}

	if cfg.Sections.Speed {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil || input.SessionID == "" {
				results <- statusline.Result{Type: "speed", Data: ""}
				return
			}
			speedInfo := speed.Calculate(lines, input.SessionID)
			results <- statusline.Result{Type: "speed", Data: speed.Format(speedInfo)}
		}()
	}

	if cfg.Sections.Autocompact {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil {
				results <- statusline.Result{Type: "autocompact", Data: ""}
				return
			}
			acInfo := context.DetectAutocompact(lines)
			results <- statusline.Result{Type: "autocompact", Data: context.FormatAutocompact(acInfo)}
		}()
	}

	if cfg.Sections.SessionName {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lines == nil {
				results <- statusline.Result{Type: "session_name", Data: ""}
				return
			}
			name := statusline.ExtractSessionName(lines, input.SessionID)
			results <- statusline.Result{Type: "session_name", Data: name}
		}()
	}

	// --- External goroutines ---

	if cfg.Sections.Git {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if input.Worktree.Branch != "" {
				branch := git.FormatWorktreeBranch(input.Worktree.Name, input.Worktree.Branch)
				results <- statusline.Result{Type: "git", Data: branch}
				return
			}
			branch := git.GetBranch(input.Workspace.CurrentDir)
			results <- statusline.Result{Type: "git", Data: branch}
		}()
	}

	if cfg.Sections.GitStatus {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gitStatusInfo := gitstatus.Get(input.Workspace.CurrentDir)
			results <- statusline.Result{Type: "git_status", Data: gitstatus.Format(gitStatusInfo)}
		}()
	}

	if cfg.Sections.Session {
		wg.Add(1)
		go func() {
			defer wg.Done()
			totalHours := session.CalculateTotalHours(input.SessionID)
			results <- statusline.Result{Type: "hours", Data: totalHours}
		}()
	}

	if cfg.Sections.APILimits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limitsInfo := apilimits.Fetch()
			results <- statusline.Result{Type: "api_limits", Data: apilimits.Format(limitsInfo)}
		}()
	}

	if cfg.Sections.ConfigInfo {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counts := statusline.CountConfigFiles(input.Workspace.CurrentDir)
			results <- statusline.Result{Type: "config_info", Data: statusline.FormatConfigCounts(counts)}
		}()
	}

	// 等待所有 goroutines 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// Phase 4: 收集結果
	var (
		gitBranch    string
		gitStatusStr string
		totalHours   string
		contextUsage string
		userMessage  string
		toolsStr     string
		agentsStr    string
		todoStr      string
		speedStr     string
		autocompact  string
		sessionName  string
		apiLimits    string
		configInfo   string
	)

	for result := range results {
		switch result.Type {
		case "git":
			gitBranch = result.Data.(string)
		case "git_status":
			gitStatusStr = result.Data.(string)
		case "hours":
			totalHours = result.Data.(string)
		case "context":
			contextUsage = result.Data.(string)
		case "message":
			userMessage = result.Data.(string)
		case "tools":
			toolsStr = result.Data.(string)
		case "agents":
			agentsStr = result.Data.(string)
		case "todo":
			todoStr = result.Data.(string)
		case "speed":
			speedStr = result.Data.(string)
		case "autocompact":
			autocompact = result.Data.(string)
		case "session_name":
			sessionName = result.Data.(string)
		case "api_limits":
			apiLimits = result.Data.(string)
		case "config_info":
			configInfo = result.Data.(string)
		}
	}

	// 更新 session（同步操作）
	session.Update(input.SessionID)

	// Phase 5: 格式化輸出
	modelDisplay := statusline.FormatModel(input.Model.DisplayName)
	projectName := filepath.Base(input.Workspace.CurrentDir)

	// Cost 顯示（顏色分級：<$5 預設，≥$5 黃色，≥$10 紅色）
	costDisplay := ""
	if cfg.Sections.Cost {
		costDisplay = statusline.FormatCostColored(input.Cost.TotalCostUSD, sep.Divider)
	}

	// 程式碼行數變化 (+N/-M)
	linesDisplay := statusline.FormatLinesChanged(
		input.Cost.TotalLinesAdded, input.Cost.TotalLinesRemoved)

	// Git 狀態附加在分支名後
	gitDisplay := gitBranch
	if gitStatusStr != "" {
		gitDisplay += statusline.FormatGitStatusDisplay(gitStatusStr)
	}

	// Speed 附加在 context 後
	speedDisplay := statusline.FormatSpeedDisplay(speedStr)

	// Autocompact 附加在 context 後
	autocompactDisplay := statusline.FormatAutocompactDisplay(autocompact)

	// Session name
	sessionNameDisplay := statusline.FormatSessionNameDisplay(sessionName)

	// Config info
	configInfoDisplay := statusline.FormatConfigCountsDisplay(configInfo)

	// 零值智慧隱藏：session time 為 "0m" 時不顯示
	sessionDisplay := totalHours
	if sessionDisplay == "0m" {
		sessionDisplay = ""
	}

	// 偵測終端寬度
	termWidth := terminal.Width()

	// Session time 與前導分隔符合併為一個段落，避免移除 session 後留下孤立分隔符
	sessionWithDivider := ""
	if sessionDisplay != "" {
		sessionWithDivider = sep.Divider + sessionDisplay
	}

	// Line 1: 主要狀態列（段落優先級截斷）
	line1Segments := []statusline.Segment{
		{Content: fmt.Sprintf("%s[%s] 📂 %s", statusline.ColorReset, modelDisplay, projectName), Priority: 1},
		{Content: sessionNameDisplay, Priority: 10},
		{Content: gitDisplay, Priority: 3},
		{Content: contextUsage, Priority: 4},
		{Content: speedDisplay, Priority: 7},
		{Content: autocompactDisplay, Priority: 8},
		{Content: linesDisplay, Priority: 9},
		{Content: sessionWithDivider, Priority: 5},
		{Content: costDisplay, Priority: 6},
		{Content: configInfoDisplay, Priority: 11},
		{Content: statusline.ColorReset, Priority: 0},
	}
	fmt.Println(statusline.TruncateLine(line1Segments, termWidth))

	// Line 2: 工具行（expanded 模式）
	if cfg.DisplayMode == "expanded" {
		if toolsLine := statusline.FormatToolsLine(toolsStr); toolsLine != "" {
			fmt.Println(toolsLine)
		}

		// Line 3: 代理行
		if agentsLine := statusline.FormatAgentsLine(agentsStr); agentsLine != "" {
			fmt.Println(agentsLine)
		}

		// Line 4: Todo + API Limits
		if todoLine := statusline.FormatTodoLine(todoStr, apiLimits); todoLine != "" {
			fmt.Println(todoLine)
		}
	} else {
		// Compact 模式：壓縮到一行
		var compactParts []string
		if toolsStr != "" {
			compactParts = append(compactParts, toolsStr)
		}
		if agentsStr != "" {
			compactParts = append(compactParts, agentsStr)
		}
		if todoStr != "" {
			compactParts = append(compactParts, todoStr)
		}
		if apiLimits != "" {
			compactParts = append(compactParts, apiLimits)
		}
		if len(compactParts) > 0 {
			compactLine := fmt.Sprintf("%s%s%s", statusline.ColorDim,
				joinWithSep(compactParts, sep.Divider), statusline.ColorReset)
			compactSegs := []statusline.Segment{{Content: compactLine, Priority: 1}}
			fmt.Println(statusline.TruncateLine(compactSegs, termWidth))
		}
	}

	// 最後一行: 使用者訊息
	if userMessage != "" {
		fmt.Print(userMessage)
	}
}

func joinWithSep(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	result := ""
	for i, p := range nonEmpty {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
