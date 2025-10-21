package voicereminder

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LoadConfig 載入配置檔案
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".claude", "voice-reminder-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return getDefaultConfig(), nil // 如果配置不存在，使用預設配置
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// IsEnabled 檢查語音提醒是否啟用
func IsEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	enabledPath := filepath.Join(homeDir, ".claude", "voice-reminder-enabled")
	data, err := os.ReadFile(enabledPath)
	if err != nil {
		return true // 預設啟用
	}

	return string(data) == "true" || string(data) == "true\n"
}

// getDefaultConfig 返回預設配置
func getDefaultConfig() *Config {
	return &Config{
		DebugMode: false,
		Language:  "zh",
		Speed:     180,
		Messages: map[string]EventMessages{
			"notification": {
				Confirmation: "Claude 需要您的確認",
				Error:        "任務失敗，請檢查",
				Completed:    "任務完成",
				Default:      "請注意",
			},
			"stop": {
				Default: "Claude 回應完成",
			},
			"subagent_stop": {
				Default: "子任務已完成",
			},
		},
		MessagesEN: map[string]EventMessages{
			"notification": {
				Confirmation: "Claude needs your confirmation",
				Error:        "Task failed, please check",
				Completed:    "Task completed",
				Default:      "Attention needed",
			},
			"stop": {
				Default: "Claude finished responding",
			},
			"subagent_stop": {
				Default: "Subagent task completed",
			},
		},
		SoundEffects: SoundConfig{
			Enabled:       true,
			FallbackSound: "/System/Library/Sounds/Glass.aiff",
		},
	}
}
