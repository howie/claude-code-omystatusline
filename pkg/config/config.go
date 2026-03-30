package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SeparatorStyle 分隔符風格
type SeparatorStyle struct {
	Left    string // 左分隔符
	Right   string // 右分隔符
	Divider string // 區段內分隔符
}

// 預定義的分隔符風格
var (
	SepPipe      = SeparatorStyle{Left: " | ", Right: " | ", Divider: " | "}
	SepPowerline = SeparatorStyle{Left: " \ue0b0 ", Right: " \ue0b2 ", Divider: " \ue0b1 "}
	SepNerdFont  = SeparatorStyle{Left: " \ue0b0 ", Right: " \ue0b2 ", Divider: " │ "}
)

// Config 使用者配置
type Config struct {
	DisplayMode    string            `json:"display_mode"`    // "expanded" or "compact"
	SeparatorStyle string            `json:"separator_style"` // "pipe", "powerline", "nerdfont"
	OverflowMode   string            `json:"overflow_mode"`   // "wrap" or "truncate" (default: "wrap"); unknown values fall back to "wrap" with a stderr warning
	Sections       SectionVisibility `json:"sections"`
}

// GetSeparator 取得目前的分隔符設定
func (c *Config) GetSeparator() SeparatorStyle {
	// 環境變數優先
	if os.Getenv("CLAUDE_STATUSLINE_POWERLINE") == "1" {
		return SepPowerline
	}
	if os.Getenv("CLAUDE_STATUSLINE_NERDFONT") == "1" {
		return SepNerdFont
	}

	// 配置檔設定
	switch c.SeparatorStyle {
	case "powerline":
		return SepPowerline
	case "nerdfont":
		return SepNerdFont
	default:
		return SepPipe
	}
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
		DisplayMode:  "expanded",
		OverflowMode: "wrap",
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
		fmt.Fprintf(os.Stderr, "statusline: could not determine home directory, using defaults: %v\n", err)
		return cfg
	}

	configPath := filepath.Join(homeDir, ".claude", "omystatusline", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "statusline: could not read config file %s: %v\n", configPath, err)
		}
		return cfg
	}

	// 先用預設值填充，再用檔案內容覆蓋
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}
