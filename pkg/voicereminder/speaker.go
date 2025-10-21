package voicereminder

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

const (
	// SpeakTimeout 定義語音播放的最大等待時間
	SpeakTimeout = 10 * time.Second
	// RetryDelay 定義重試之間的延遲時間
	RetryDelay = 500 * time.Millisecond
	// MaxRetries 定義最大重試次數
	MaxRetries = 1
)

// Speak 播放語音（帶超時和重試機制）
func Speak(message string, speed int, language string) error {
	return SpeakWithLogger(message, speed, language, nil)
}

// SpeakWithLogger 播放語音，支援 logger（帶超時和重試機制）
func SpeakWithLogger(message string, speed int, language string, logger *Logger) error {
	// 重試邏輯
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			if logger != nil {
				logger.Log("重試播放語音... (第 %d/%d 次)", attempt, MaxRetries)
			}
			time.Sleep(RetryDelay)
		}

		// 嘗試播放語音
		err := speakOnce(message, speed, language, logger)
		if err == nil {
			return nil
		}

		if logger != nil {
			logger.Log("播放失敗: %v", err)
		}
	}

	// 所有重試都失敗，使用降級音效
	if logger != nil {
		logger.Log("所有重試失敗，使用降級音效")
	}
	return playFallbackSound()
}

// speakOnce 執行一次語音播放（帶超時機制）
func speakOnce(message string, speed int, language string, logger *Logger) error {
	if runtime.GOOS == "darwin" {
		return speakMacOS(message, speed, language, logger)
	}

	// Linux 嘗試使用 espeak
	if hasCommand("espeak") {
		return speakWithTimeout("espeak", []string{message}, logger)
	}

	// 降級：播放系統音效
	return playFallbackSound()
}

// speakMacOS 在 macOS 上播放語音（帶超時機制）
func speakMacOS(message string, speed int, language string, logger *Logger) error {
	voice := selectVoice(language)
	args := []string{"-v", voice, "-r", fmt.Sprintf("%d", speed), message}
	return speakWithTimeout("say", args, logger)
}

// speakWithTimeout 執行語音命令並帶超時機制
func speakWithTimeout(command string, args []string, logger *Logger) error {
	cmd := exec.Command(command, args...)

	// 建立 channel 來接收執行結果
	done := make(chan error, 1)

	// 在 goroutine 中執行命令
	go func() {
		done <- cmd.Run()
	}()

	// 等待命令完成或超時
	select {
	case err := <-done:
		// 命令正常完成
		if err != nil {
			return fmt.Errorf("命令執行失敗: %v", err)
		}
		return nil

	case <-time.After(SpeakTimeout):
		// 超時，嘗試終止進程
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		if logger != nil {
			logger.Log("⚠️ 語音播放超時 (%.0f 秒)，已終止進程", SpeakTimeout.Seconds())
		}
		return fmt.Errorf("語音播放超時")
	}
}

// selectVoice 根據語言選擇合適的語音
func selectVoice(language string) string {
	switch language {
	case "zh":
		// 優先使用 Meijia（美佳，繁體中文台灣）
		return "Meijia"
	case "en":
		return "Samantha"
	default:
		// 預設使用繁體中文語音
		return "Meijia"
	}
}

// PlayFallbackSound 播放降級音效
func playFallbackSound() error {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("afplay", "/System/Library/Sounds/Glass.aiff")
		return cmd.Run()
	}

	if hasCommand("paplay") {
		cmd := exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
		return cmd.Run()
	}

	if hasCommand("aplay") {
		cmd := exec.Command("aplay", "/usr/share/sounds/alsa/Front_Center.wav")
		return cmd.Run()
	}

	// 最後降級：終端機鈴聲
	fmt.Print("\a")
	return nil
}

// hasCommand 檢查命令是否存在
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
