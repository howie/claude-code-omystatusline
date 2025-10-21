package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/howie/claude-code-omystatusline/pkg/voicereminder"
)

func main() {
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
	if err := voicereminder.Speak(message, config.Speed, config.Language); err != nil {
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
