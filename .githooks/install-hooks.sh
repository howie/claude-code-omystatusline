#!/bin/bash
#
# 安裝 Git hooks 腳本
# Install Git hooks for claude-code-omystatusline
#

set -e

# 顏色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Installing Git Hooks for claude-code-omystatusline${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo ""

# 取得腳本所在目錄
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

# 檢查是否在 Git 倉庫中
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "錯誤：不在 Git 倉庫中"
    exit 1
fi

# 創建 hooks 目錄（如果不存在）
mkdir -p "$GIT_HOOKS_DIR"

# 安裝 pre-push hook
if [ -f "$SCRIPT_DIR/pre-push" ]; then
    echo -e "${BLUE}ℹ${NC} 安裝 pre-push hook..."
    cp "$SCRIPT_DIR/pre-push" "$GIT_HOOKS_DIR/pre-push"
    chmod +x "$GIT_HOOKS_DIR/pre-push"
    echo -e "${GREEN}✓${NC} pre-push hook 已安裝"
else
    echo "警告：找不到 pre-push hook"
fi

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Git Hooks 安裝完成！${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
echo ""
echo "已安裝的 hooks："
echo "  ✓ pre-push - 推送前執行測試檢查"
echo ""
echo "如需停用 hooks，執行："
echo "  rm .git/hooks/pre-push"
echo ""
