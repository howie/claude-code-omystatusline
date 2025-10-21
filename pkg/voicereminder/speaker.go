package voicereminder

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Speak 播放語音
func Speak(message string, speed int, language string) error {
	if runtime.GOOS == "darwin" {
		// macOS 使用 say 命令，中文需要指定語音
		voice := selectVoice(language)
		cmd := exec.Command("say", "-v", voice, "-r", fmt.Sprintf("%d", speed), message)
		return cmd.Run()
	}

	// Linux 嘗試使用 espeak
	if hasCommand("espeak") {
		cmd := exec.Command("espeak", message)
		return cmd.Run()
	}

	// 降級：播放系統音效
	return playFallbackSound()
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
