package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 使用者配置
type Config struct {
	DisplayMode string            `json:"display_mode"` // "expanded" or "compact"
	Sections    SectionVisibility `json:"sections"`
}

// SectionVisibility 各區段的可見性設定
type SectionVisibility struct {
	Model       bool `json:"model"`
	Git         bool `json:"git"`
	GitStatus   bool `json:"git_status"`
	Context     bool `json:"context"`
	Session     bool `json:"session"`
	Cost        bool `json:"cost"`
	Tools       bool `json:"tools"`
	Agents      bool `json:"agents"`
	Todo        bool `json:"todo"`
	APILimits   bool `json:"api_limits"`
	Speed       bool `json:"speed"`
	SessionName bool `json:"session_name"`
	ConfigInfo  bool `json:"config_info"`
	Autocompact bool `json:"autocompact"`
	UserMessage bool `json:"user_message"`
}

// DefaultConfig 返回預設配置（所有區段可見）
func DefaultConfig() *Config {
	return &Config{
		DisplayMode: "expanded",
		Sections: SectionVisibility{
			Model:       true,
			Git:         true,
			GitStatus:   true,
			Context:     true,
			Session:     true,
			Cost:        true,
			Tools:       true,
			Agents:      true,
			Todo:        true,
			APILimits:   true,
			Speed:       true,
			SessionName: true,
			ConfigInfo:  true,
			Autocompact: true,
			UserMessage: true,
		},
	}
}

// Load 載入配置檔，找不到則返回預設值
func Load() *Config {
	cfg := DefaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	configPath := filepath.Join(homeDir, ".claude", "omystatusline", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	// 先用預設值填充，再用檔案內容覆蓋
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}

	// 驗證 DisplayMode
	if cfg.DisplayMode != "expanded" && cfg.DisplayMode != "compact" {
		cfg.DisplayMode = "expanded"
	}

	return cfg
}
