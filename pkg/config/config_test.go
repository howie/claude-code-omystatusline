package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DisplayMode != "expanded" {
		t.Fatalf("expected display_mode 'expanded', got %q", cfg.DisplayMode)
		return
	}
	if cfg.OverflowMode != "wrap" {
		t.Fatalf("expected overflow_mode 'wrap', got %q", cfg.OverflowMode)
		return
	}
	if !cfg.Sections.Model {
		t.Fatal("expected sections.model to be true by default")
		return
	}
	if !cfg.Sections.Tools {
		t.Fatal("expected sections.tools to be true by default")
		return
	}
	if !cfg.Sections.APILimits {
		t.Fatal("expected sections.api_limits to be true by default")
		return
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load should return default config when file is missing")
		return
	}
	if cfg.DisplayMode != "expanded" {
		t.Fatalf("expected default display_mode, got %q", cfg.DisplayMode)
		return
	}
}

func TestLoadCustomConfig(t *testing.T) {
	dir := t.TempDir()

	// 模擬 home 目錄
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".claude", "omystatusline")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
		return
	}

	configJSON := `{"display_mode":"compact","sections":{"tools":false}}` // 刻意省略 overflow_mode
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
		return
	}

	cfg := Load()
	if cfg.DisplayMode != "compact" {
		t.Fatalf("expected display_mode 'compact', got %q", cfg.DisplayMode)
		return
	}
	if cfg.Sections.Tools {
		t.Fatal("expected sections.tools to be false")
		return
	}
	// overflow_mode 不在 JSON 中，應保持預設值 "wrap"
	if cfg.OverflowMode != "wrap" {
		t.Fatalf("expected overflow_mode 'wrap' when omitted from JSON, got %q", cfg.OverflowMode)
		return
	}
}

func TestLoadOverflowModeTruncate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".claude", "omystatusline")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
		return
	}

	configJSON := `{"overflow_mode":"truncate"}`
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
		return
	}

	cfg := Load()
	if cfg.OverflowMode != "truncate" {
		t.Fatalf("expected overflow_mode 'truncate', got %q", cfg.OverflowMode)
		return
	}
	// 其他欄位應保持預設值
	if cfg.DisplayMode != "expanded" {
		t.Fatalf("expected default display_mode 'expanded', got %q", cfg.DisplayMode)
		return
	}
}
