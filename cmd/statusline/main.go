package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/git"
	"github.com/howie/claude-code-omystatusline/pkg/session"
	"github.com/howie/claude-code-omystatusline/pkg/statusline"
)

func main() {
	var input statusline.Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode input: %v\n", err)
		os.Exit(1)
	}

	// 建立結果通道
	results := make(chan statusline.Result, 4)
	var wg sync.WaitGroup

	// 並行獲取各種資訊
	wg.Add(4)

	go func() {
		defer wg.Done()
		branch := git.GetBranch()
		results <- statusline.Result{Type: "git", Data: branch}
	}()

	go func() {
		defer wg.Done()
		totalHours := session.CalculateTotalHours(input.SessionID)
		results <- statusline.Result{Type: "hours", Data: totalHours}
	}()

	go func() {
		defer wg.Done()
		contextInfo := context.Analyze(input.TranscriptPath)
		results <- statusline.Result{Type: "context", Data: contextInfo}
	}()

	go func() {
		defer wg.Done()
		userMsg := statusline.ExtractUserMessage(input.TranscriptPath, input.SessionID)
		results <- statusline.Result{Type: "message", Data: userMsg}
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
	session.Update(input.SessionID)

	// 格式化模型顯示
	modelDisplay := statusline.FormatModel(input.Model.DisplayName)
	projectName := filepath.Base(input.Workspace.CurrentDir)

	// 輸出狀態列
	fmt.Printf("%s[%s] 📂 %s%s%s | %s%s\n",
		statusline.ColorReset, modelDisplay, projectName, gitBranch,
		contextUsage, totalHours, statusline.ColorReset)

	// 輸出使用者訊息
	if userMessage != "" {
		fmt.Print(userMessage)
	}
}
