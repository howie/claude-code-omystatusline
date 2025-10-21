# Makefile for Claude Code Statusline

# 變數定義
INSTALL_DIR = $(HOME)/.claude/omystatusline
CLAUDE_DIR = $(HOME)/.claude
OUTPUT_DIR = output
BINARY_NAME = statusline-go
VOICE_REMINDER_BINARY = voice-reminder
CMD_SOURCE = cmd/statusline
VOICE_REMINDER_SOURCE = cmd/voice-reminder
WRAPPER_SCRIPT = scripts/statusline-wrapper.sh
BASH_SCRIPT = scripts/statusline.sh
INSTALL_SCRIPT = scripts/install.sh

# Go 編譯選項
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOFLAGS = -ldflags="-s -w"

.PHONY: all build build-voice-reminder install install-simple uninstall clean test lint fmt install-hooks uninstall-hooks help

# 預設目標
all: build build-voice-reminder

# 編譯 Go binary
build:
	@echo "正在編譯 $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	@go build $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./$(CMD_SOURCE)
	@echo "編譯完成: $(OUTPUT_DIR)/$(BINARY_NAME)"

# 編譯 voice-reminder binary
build-voice-reminder:
	@echo "正在編譯 $(VOICE_REMINDER_BINARY)..."
	@mkdir -p $(OUTPUT_DIR)
	@go build $(GOFLAGS) -o $(OUTPUT_DIR)/$(VOICE_REMINDER_BINARY) ./$(VOICE_REMINDER_SOURCE)
	@echo "編譯完成: $(OUTPUT_DIR)/$(VOICE_REMINDER_BINARY)"

# 互動式安裝（推薦）
install:
	@./$(INSTALL_SCRIPT)

# 簡單安裝（不含音訊提醒）
install-simple: build
	@echo "正在安裝到 $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)/bin
	@mkdir -p $(INSTALL_DIR)/scripts
	@cp $(OUTPUT_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/bin/$(BINARY_NAME)
	@cp $(WRAPPER_SCRIPT) $(INSTALL_DIR)/bin/statusline-wrapper.sh
	@cp $(BASH_SCRIPT) $(INSTALL_DIR)/scripts/statusline.sh
	@chmod +x $(INSTALL_DIR)/bin/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/bin/statusline-wrapper.sh
	@chmod +x $(INSTALL_DIR)/scripts/statusline.sh
	@echo "✓ 安裝完成！"
	@echo "已安裝檔案："
	@echo "  - $(INSTALL_DIR)/bin/$(BINARY_NAME)"
	@echo "  - $(INSTALL_DIR)/bin/statusline-wrapper.sh"
	@echo "  - $(INSTALL_DIR)/scripts/statusline.sh"
	@echo ""
	@echo "提示：使用 'make install' 進行互動式安裝，可設定音訊提醒功能"

# 卸載
uninstall:
	@echo "正在卸載..."
	@rm -rf $(INSTALL_DIR)
	@rm -f $(CLAUDE_DIR)/commands/voice-reminder-*.md
	@echo "✓ 卸載完成！"
	@echo "提示：如需完全清理，請手動檢查 $(CLAUDE_DIR)/config.json"

# 清理編譯產物
clean:
	@echo "正在清理..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✓ 清理完成！"

# 執行測試
test:
	@echo "正在執行測試..."
	@go test -v ./...
	@echo "✓ 測試完成！"

# 格式化程式碼
fmt:
	@echo "正在格式化程式碼..."
	@gofmt -s -w .
	@echo "✓ 程式碼格式化完成！"

# 執行 linting
lint:
	@echo "正在執行 gofmt 檢查..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ 以下檔案需要格式化："; \
		gofmt -s -l .; \
		echo "請執行 'gofmt -s -w .' 或 'make fmt' 來格式化"; \
		exit 1; \
	fi
	@echo "✓ gofmt 檢查通過！"
	@echo "正在執行 golangci-lint..."
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "❌ golangci-lint 未安裝"; \
		echo "請執行以下命令安裝："; \
		echo "  brew install golangci-lint  # macOS"; \
		echo "  或訪問 https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@golangci-lint run --timeout=5m
	@echo "✓ Linting 完成！"

# 安裝 Git hooks
install-hooks:
	@echo "正在安裝 Git hooks..."
	@chmod +x .githooks/install-hooks.sh
	@./.githooks/install-hooks.sh

# 卸載 Git hooks
uninstall-hooks:
	@echo "正在卸載 Git hooks..."
	@rm -f .git/hooks/pre-push
	@echo "✓ Git hooks 已卸載"

# 顯示幫助
help:
	@echo "Claude Code Statusline - Makefile 使用說明"
	@echo ""
	@echo "可用的 targets:"
	@echo "  make build          - 編譯 Go binary"
	@echo "  make install        - 互動式安裝（推薦，包含音訊提醒設定）"
	@echo "  make install-simple - 簡單安裝（僅狀態列，不含音訊提醒）"
	@echo "  make uninstall      - 從 ~/.claude/ 卸載所有檔案"
	@echo "  make clean          - 清理編譯產物"
	@echo "  make test           - 執行測試"
	@echo "  make fmt            - 格式化程式碼 (gofmt)"
	@echo "  make lint           - 執行程式碼檢查 (gofmt + golangci-lint)"
	@echo "  make help           - 顯示此幫助訊息"
	@echo ""
	@echo "環境變數:"
	@echo "  GOOS                - 目標作業系統 (預設: $(GOOS))"
	@echo "  GOARCH              - 目標架構 (預設: $(GOARCH))"
	@echo ""
	@echo "推薦使用互動式安裝："
	@echo "  make install"
