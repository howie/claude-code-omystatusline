# 重構遷移指南

## 概述

本次重構將單一檔案的專案重新組織為模組化架構，提升可維護性和擴展性。

## 主要變更

### 1. 程式碼拆分

| 原始檔案 | 新位置 | 說明 |
|---------|--------|------|
| `statusline.go` | `cmd/statusline/main.go` | 主程式入口 |
| `statusline.go` (Git 部分) | `pkg/git/branch.go` | Git 分支偵測 |
| `statusline.go` (Session 部分) | `pkg/session/tracker.go` | 時間追蹤 |
| `statusline.go` (Context 部分) | `pkg/context/tracker.go` | Token 追蹤 |
| `statusline.go` (顏色/格式) | `pkg/statusline/*.go` | 核心邏輯 |

### 2. 腳本檔案移動

| 原始位置 | 新位置 |
|---------|--------|
| `install.sh` | `scripts/install.sh` |
| `statusline-wrapper.sh` | `scripts/statusline-wrapper.sh` |
| `statusline.sh` | `scripts/statusline.sh` |

### 3. 文檔重組

| 原始位置 | 新位置 |
|---------|--------|
| `examples/*.sh` | `docs/examples/*.sh` |
| `examples/README.md` | `docs/examples/README.md` |
| (新增) | `docs/guides/architecture.md` |
| (新增) | `docs/guides/migration.md` |

## 向後相容性

### 安裝路徑保持不變

雖然原始碼重組,但安裝後的檔案路徑**完全相同**:

```bash
~/.claude/statusline-go
~/.claude/statusline-wrapper.sh
~/.claude/statusline.sh
```

### 設定檔無需修改

`~/.claude/config.json` 中的設定保持不變:

```json
{
  "statusLineCommand": "~/.claude/statusline-wrapper.sh"
}
```

## 升級步驟

如果你已經安裝舊版本:

### 方法 1: 重新安裝 (推薦)

```bash
# 1. 拉取最新程式碼
git pull origin main

# 2. 重新安裝
make install
```

### 方法 2: 手動更新

```bash
# 1. 拉取最新程式碼
git pull origin main

# 2. 清理舊檔案
make clean

# 3. 編譯新版本
make build

# 4. 複製到安裝目錄
make install-simple
```

## 程式碼變更對照

### 函式移動對照表

| 原始函式 | 新位置 | 包名 |
|---------|--------|------|
| `main()` | `cmd/statusline/main.go` | `main` |
| `getGitBranch()` | `pkg/git/branch.go` | `git.GetBranch()` |
| `updateSession()` | `pkg/session/tracker.go` | `session.Update()` |
| `calculateTotalHours()` | `pkg/session/tracker.go` | `session.CalculateTotalHours()` |
| `analyzeContext()` | `pkg/context/tracker.go` | `context.Analyze()` |
| `calculateContextUsage()` | `pkg/context/tracker.go` | (私有函式) |
| `generateProgressBar()` | `pkg/context/tracker.go` | (私有函式) |
| `formatModel()` | `pkg/statusline/builder.go` | `statusline.FormatModel()` |
| `extractUserMessage()` | `pkg/statusline/builder.go` | `statusline.ExtractUserMessage()` |
| `formatUserMessage()` | `pkg/statusline/builder.go` | (私有函式) |
| `isSystemMessage()` | `pkg/statusline/builder.go` | (私有函式) |

### 常數移動對照表

| 原始常數 | 新位置 |
|---------|--------|
| `ColorReset`, `ColorGold`, ... | `pkg/statusline/color.go` |
| `modelConfig` | `pkg/statusline/builder.go` |

### 資料結構移動對照表

| 原始結構 | 新位置 |
|---------|--------|
| `Input` | `pkg/statusline/model.go` |
| `Result` | `pkg/statusline/model.go` |
| `Session`, `Interval` | `pkg/session/tracker.go` |

## 開發者注意事項

### Import 路徑

如果你有 fork 或自定義版本,請更新 import 路徑:

```go
// 新的 import 路徑
import (
    "github.com/howie/claude-code-omystatusline/pkg/git"
    "github.com/howie/claude-code-omystatusline/pkg/context"
    "github.com/howie/claude-code-omystatusline/pkg/session"
    "github.com/howie/claude-code-omystatusline/pkg/statusline"
)
```

### 編譯指令

```bash
# 舊版
go build statusline.go

# 新版
go build ./cmd/statusline
# 或
make build
```

### 測試指令

```bash
# 舊版
go test -v

# 新版
go test ./...
# 或
make test
```

## 優點總結

1. **模組化**: 各功能獨立,易於維護
2. **可測試**: 每個套件可獨立測試
3. **可擴展**: 新增功能只需加入新套件
4. **標準化**: 符合 Go 專案標準結構
5. **整潔**: 檔案有組織,不再散落一地

## 問題排查

### 編譯失敗

```bash
# 確認 Go 版本
go version  # 需要 1.21+

# 清理並重新編譯
make clean
make build
```

### 測試失敗

```bash
# 執行詳細測試
go test -v ./...

# 檢查特定套件
go test -v ./pkg/git
```

### 安裝問題

```bash
# 使用互動式安裝
make install

# 或手動複製
cp statusline-go ~/.claude/
cp scripts/statusline-wrapper.sh ~/.claude/
```

## 參考資源

- [架構文檔](./architecture.md)
- [主 README](../../README.md)
- [安裝指南](../installation.md) (如果存在)

## 回饋

如有問題或建議,請開 [GitHub Issue](https://github.com/howie/claude-code-omystatusline/issues)。
