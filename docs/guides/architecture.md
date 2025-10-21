# 專案架構說明

## 目錄結構

```
claude-code-omystatusline/
├── cmd/                        # 可執行程式入口
│   └── statusline/
│       └── main.go            # 主程式入口，協調各模組運作
├── pkg/                        # 可重用的套件庫
│   ├── git/                   # Git 相關功能
│   │   └── branch.go         # 分支偵測、worktree 檢測、快取機制
│   ├── context/               # Context 追蹤
│   │   └── tracker.go        # Token 計算、進度條、百分比格式化
│   ├── session/               # Session 時間追蹤
│   │   └── tracker.go        # 時間累積、多 session 偵測
│   └── statusline/            # 狀態列核心邏輯
│       ├── builder.go        # 狀態列組裝、訊息提取
│       ├── color.go          # ANSI 顏色定義
│       └── model.go          # 資料結構定義
├── scripts/                   # 安裝和輔助腳本
│   ├── install.sh            # 互動式安裝程式
│   ├── statusline-wrapper.sh # 狀態列包裝腳本
│   └── statusline.sh         # Bash 版本 (備用)
├── docs/                      # 文檔
│   ├── guides/               # 使用指南
│   │   └── architecture.md   # 架構說明 (本文件)
│   ├── features/             # 功能文檔
│   │   ├── audio-notifications/
│   │   └── limitation-warning/
│   ├── examples/             # 使用範例
│   └── images/               # 圖片資源
├── .github/                   # GitHub Actions CI/CD
├── .githooks/                 # Git hooks
├── Makefile                   # 建置和安裝自動化
├── go.mod                     # Go 模組定義
├── README.md                  # 專案說明
├── LICENSE                    # Apache 2.0
├── CHANGELOG.md               # 版本變更記錄
└── VERSION                    # 版本號碼
```

## 模組職責

### cmd/statusline
- **職責**: 程式入口點
- **功能**:
  - 解析來自 Claude Code 的 JSON 輸入
  - 並行啟動各項資訊收集 goroutine
  - 組裝最終狀態列輸出

### pkg/git
- **職責**: Git 倉庫資訊獲取
- **功能**:
  - 偵測當前分支
  - 判斷是否在 worktree 中
  - 提供 5 秒快取避免頻繁呼叫 git 指令

### pkg/context
- **職責**: Token 使用量追蹤
- **功能**:
  - 分析 transcript 檔案提取 token 計數
  - 計算使用百分比 (基於 200k 上限)
  - 生成視覺化進度條
  - 根據使用量提供顏色編碼警告

### pkg/session
- **職責**: Session 時間管理
- **功能**:
  - 追蹤每日累積工作時間
  - 偵測多個活躍 session
  - 智慧間隔管理 (10 分鐘以上視為新區段)
  - 持久化 session 資料到 `~/.claude/session-tracker/`

### pkg/statusline
- **職責**: 狀態列核心邏輯
- **功能**:
  - 定義資料結構 (Input, Result)
  - 定義 ANSI 顏色常數
  - 格式化模型顯示 (Opus/Sonnet/Haiku)
  - 提取和格式化使用者訊息
  - 過濾系統訊息

## 設計原則

1. **關注點分離**: 每個套件專注單一職責
2. **並行處理**: 使用 goroutine 同時取得 git/context/session 資訊
3. **效能優化**:
   - Git 分支資訊快取 5 秒
   - Transcript 只讀最後 100-200 行
   - 使用 channel 進行並行結果收集
4. **容錯設計**: 所有外部操作 (git, 檔案讀取) 失敗時優雅降級
5. **可測試性**: 模組化設計便於單元測試

## 資料流

```
Claude Code
    │
    ▼
JSON Input (stdin)
    │
    ▼
main.go 解析
    │
    ├─── goroutine 1: git.GetBranch()
    ├─── goroutine 2: session.CalculateTotalHours()
    ├─── goroutine 3: context.Analyze()
    └─── goroutine 4: statusline.ExtractUserMessage()
    │
    ▼
收集結果 (via channel)
    │
    ▼
組裝狀態列字串
    │
    ▼
輸出到 stdout (Claude Code 顯示)
```

## 編譯和安裝

### 編譯
```bash
make build
# 或
go build -o statusline-go ./cmd/statusline
```

### 安裝
```bash
# 互動式安裝 (推薦)
make install

# 簡單安裝 (僅狀態列)
make install-simple
```

### 測試
```bash
make test
# 或
go test ./...
```

## 擴展指南

### 新增資訊來源

1. 在 `pkg/` 下建立新套件 (如 `pkg/diagnostics/`)
2. 實作資訊收集函式
3. 在 `cmd/statusline/main.go` 中新增對應 goroutine
4. 更新結果收集邏輯
5. 調整狀態列格式化輸出

### 範例: 新增 CPU 使用率顯示

```go
// pkg/system/cpu.go
package system

func GetCPUUsage() string {
    // 實作邏輯...
    return "CPU: 45%"
}

// cmd/statusline/main.go
go func() {
    defer wg.Done()
    cpuInfo := system.GetCPUUsage()
    results <- statusline.Result{Type: "cpu", Data: cpuInfo}
}()
```

## 效能考量

- **狀態列更新延遲**: < 100ms
- **記憶體使用**: < 10MB
- **Git 快取**: 5 秒 TTL
- **Transcript 讀取**: 僅最後 100 行 (context) / 200 行 (message)

## 相依性

- **Go**: 1.21+
- **Git**: 任何版本 (用於分支偵測)
- **執行環境**: macOS / Linux (支援 ANSI 顏色的終端)

## 授權

Apache License 2.0 - 詳見 [LICENSE](../../LICENSE)
