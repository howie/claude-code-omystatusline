# Makefile for Claude Code Statusline

# 變數定義
INSTALL_DIR = $(HOME)/.claude
BINARY_NAME = statusline-go
GO_SOURCE = statusline.go
WRAPPER_SCRIPT = statusline-wrapper.sh
BASH_SCRIPT = statusline.sh

# Go 編譯選項
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOFLAGS = -ldflags="-s -w"

.PHONY: all build install uninstall clean help

# 預設目標
all: build

# 編譯 Go binary
build:
	@echo "正在編譯 $(BINARY_NAME)..."
	@go build $(GOFLAGS) -o $(BINARY_NAME) $(GO_SOURCE)
	@echo "編譯完成: $(BINARY_NAME)"

# 安裝到 ~/.claude/
install: build
	@echo "正在安裝到 $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@cp $(WRAPPER_SCRIPT) $(INSTALL_DIR)/$(WRAPPER_SCRIPT)
	@cp $(BASH_SCRIPT) $(INSTALL_DIR)/$(BASH_SCRIPT)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(WRAPPER_SCRIPT)
	@chmod +x $(INSTALL_DIR)/$(BASH_SCRIPT)
	@echo "✓ 安裝完成！"
	@echo "已安裝檔案："
	@echo "  - $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "  - $(INSTALL_DIR)/$(WRAPPER_SCRIPT)"
	@echo "  - $(INSTALL_DIR)/$(BASH_SCRIPT)"

# 卸載
uninstall:
	@echo "正在卸載..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -f $(INSTALL_DIR)/$(WRAPPER_SCRIPT)
	@rm -f $(INSTALL_DIR)/$(BASH_SCRIPT)
	@echo "✓ 卸載完成！"

# 清理編譯產物
clean:
	@echo "正在清理..."
	@rm -f $(BINARY_NAME)
	@echo "✓ 清理完成！"

# 顯示幫助
help:
	@echo "Claude Code Statusline - Makefile 使用說明"
	@echo ""
	@echo "可用的 targets:"
	@echo "  make build      - 編譯 Go binary"
	@echo "  make install    - 編譯並安裝所有檔案到 ~/.claude/"
	@echo "  make uninstall  - 從 ~/.claude/ 卸載所有檔案"
	@echo "  make clean      - 清理編譯產物"
	@echo "  make help       - 顯示此幫助訊息"
	@echo ""
	@echo "環境變數:"
	@echo "  GOOS            - 目標作業系統 (預設: $(GOOS))"
	@echo "  GOARCH          - 目標架構 (預設: $(GOARCH))"
