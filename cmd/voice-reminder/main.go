package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/howie/claude-code-omystatusline/pkg/voicereminder"
)

const version = "1.1.0"

func main() {
	// 處理命令列參數
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			fmt.Printf("voice-reminder version %s\n", version)
			return
		}
	}

	// 1. 讀取 stdin (Hook JSON 輸入)
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
		os.Exit(1)
	}

	// 2. 檢查啟用狀態
	if !voicereminder.IsEnabled() {
		// 直接 passthrough
		fmt.Println(string(inputBytes))
		return
	}

	// 3. 載入配置
	config, err := voicereminder.LoadConfig()
	if err != nil {
		// 配置載入失敗，passthrough
		fmt.Println(string(inputBytes))
		return
	}

	// 4. 初始化 logger
	logger := voicereminder.NewLogger(config.DebugMode || os.Getenv("VOICE_REMINDER_DEBUG") == "true")
	defer logger.Close()

	logger.Log("========== Hook 觸發 ==========")
	logger.Log("收到的原始 JSON: %s", string(inputBytes))

	// 5. 解析 JSON
	var input voicereminder.HookInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		logger.Log("JSON 解析錯誤: %v", err)
		fmt.Println(string(inputBytes))
		return
	}

	logger.Log("解析結果 - EventName: %s, Message: %s", input.HookEventName, input.Message)
	logger.Log("配置 - Language: %s, Speed: %d", config.Language, config.Speed)

	// 6. 選擇訊息
	message := voicereminder.SelectMessage(config, &input)
	logger.Log("選擇的語音訊息: %s", message)

	// 7. 播放語音
	logger.Log("開始播放語音...")
	if err := voicereminder.SpeakWithLogger(message, config.Speed, config.Language, logger); err != nil {
		logger.Log("語音播放錯誤: %v", err)
	} else {
		logger.Log("語音播放成功")
	}

	// 8. 更新統計
	if err := voicereminder.UpdateStats(input.HookEventName); err != nil {
		logger.Log("統計更新錯誤: %v", err)
	} else {
		logger.Log("統計已更新")
	}

	// 9. Passthrough (Hook 必須)
	fmt.Println(string(inputBytes))
	logger.Log("========== 處理完成 ==========\n")
}

func printHelp() {
	fmt.Println("voice-reminder - Claude Code Voice Notification System")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("DESCRIPTION:")
	fmt.Println("  A hook handler for Claude Code that provides voice notifications")
	fmt.Println("  when important events occur (confirmations, errors, completions).")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  voice-reminder < hook-input.json")
	fmt.Println()
	fmt.Println("  This command is typically used as a Claude Code hook and receives")
	fmt.Println("  JSON input from stdin containing hook event information.")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Println("  Config file:  ~/.claude/voice-reminder-config.json")
	fmt.Println("  Enable file:  ~/.claude/voice-reminder-enabled")
	fmt.Println("  Debug log:    ~/.claude/voice-reminder-debug.log")
	fmt.Println("  Stats file:   ~/.claude/voice-reminder-stats.json")
	fmt.Println()
	fmt.Println("FEATURES:")
	fmt.Println("  - 10-second timeout protection for voice playback")
	fmt.Println("  - Automatic retry on failure (1 retry)")
	fmt.Println("  - Fallback to system sounds when voice fails")
	fmt.Println("  - Multi-language support (English, Chinese)")
	fmt.Println("  - Debug logging for troubleshooting")
	fmt.Println()
	fmt.Println("SLASH COMMANDS:")
	fmt.Println("  /voice-reminder-on     Enable voice notifications")
	fmt.Println("  /voice-reminder-off    Disable voice notifications")
	fmt.Println("  /voice-reminder-stats  Show usage statistics")
	fmt.Println("  /voice-reminder-test   Test the voice system")
	fmt.Println()
	fmt.Println("For more information, visit:")
	fmt.Println("  https://github.com/howie/claude-code-omystatusline")
}
