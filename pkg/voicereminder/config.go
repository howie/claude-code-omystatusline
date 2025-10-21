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

	// 新路徑優先
	newConfigPath := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "config", "voice-reminder-config.json")
	// 舊路徑作為備援
	oldConfigPath := filepath.Join(homeDir, ".claude", "voice-reminder-config.json")

	var data []byte

	// 先嘗試新路徑
	if _, err := os.Stat(newConfigPath); err == nil {
		data, err = os.ReadFile(newConfigPath)
		if err != nil {
			return getDefaultConfig(), nil
		}
	} else if _, err := os.Stat(oldConfigPath); err == nil {
		// 回退到舊路徑
		data, err = os.ReadFile(oldConfigPath)
		if err != nil {
			return getDefaultConfig(), nil
		}
	} else {
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

	// 新路徑優先
	newEnabledPath := filepath.Join(homeDir, ".claude", "omystatusline", "plugins", "voice-reminder", "data", "voice-reminder-enabled")
	// 舊路徑作為備援
	oldEnabledPath := filepath.Join(homeDir, ".claude", "voice-reminder-enabled")

	// 先嘗試新路徑
	if _, err := os.Stat(newEnabledPath); err == nil {
		data, err := os.ReadFile(newEnabledPath)
		if err != nil {
			return true // 讀取失敗，預設啟用
		}
		return string(data) == "true" || string(data) == "true\n"
	}

	// 回退到舊路徑
	if _, err := os.Stat(oldEnabledPath); err == nil {
		data, err := os.ReadFile(oldEnabledPath)
		if err != nil {
			return true // 讀取失敗，預設啟用
		}
		return string(data) == "true" || string(data) == "true\n"
	}

	return true // 預設啟用
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
