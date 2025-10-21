# 快速入門指南 - 重構後的專案

## 🚀 快速開始

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

# 簡單安裝
make install-simple
```

### 測試
```bash
make test
# 或
go test ./...
```

### 清理
```bash
make clean
```

---

## 📂 專案結構一覽

```
.
├── cmd/statusline/          → 主程式入口
├── pkg/                     → 可重用套件
│   ├── git/                → Git 分支偵測
│   ├── context/            → Token 追蹤
│   ├── session/            → 時間管理
│   └── statusline/         → 核心邏輯
├── scripts/                → 安裝腳本
└── docs/                   → 文檔
    ├── guides/            → 架構 & 遷移指南
    └── examples/          → 使用範例
```

---

## 🔍 快速查找

### 想修改 Git 分支顯示?
→ `pkg/git/branch.go`

### 想調整 Token 進度條?
→ `pkg/context/tracker.go`

### 想修改時間追蹤邏輯?
→ `pkg/session/tracker.go`

### 想改變顏色配置?
→ `pkg/statusline/color.go`

### 想修改狀態列格式?
→ `cmd/statusline/main.go` (組裝邏輯)
→ `pkg/statusline/builder.go` (格式化函式)

---

## 🧪 如何新增功能?

### 範例: 新增 CPU 使用率顯示

1. **建立新套件**:
   ```bash
   mkdir pkg/system
   ```

2. **實作功能**:
   ```go
   // pkg/system/cpu.go
   package system

   func GetCPUUsage() string {
       // 實作邏輯...
       return "CPU: 45%"
   }
   ```

3. **整合到主程式**:
   ```go
   // cmd/statusline/main.go
   import "github.com/howie/claude-code-omystatusline/pkg/system"

   // 在 main() 中新增 goroutine
   go func() {
       defer wg.Done()
       cpuInfo := system.GetCPUUsage()
       results <- statusline.Result{Type: "cpu", Data: cpuInfo}
   }()
   ```

4. **更新狀態列輸出**:
   ```go
   fmt.Printf("... | %s | ...", cpuUsage)
   ```

---

## 📖 詳細文檔

- **架構說明**: [docs/guides/architecture.md](docs/guides/architecture.md)
- **遷移指南**: [docs/guides/migration.md](docs/guides/migration.md)
- **重構總結**: [REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md)
- **主 README**: [README.md](README.md)

---

## 🐛 常見問題

### Q: 編譯失敗?
```bash
# 確認 Go 版本
go version  # 需要 1.21+

# 清理並重新編譯
make clean
make build
```

### Q: 測試失敗?
```bash
# 執行詳細測試
go test -v ./...

# 測試特定套件
go test -v ./pkg/git
go test -v ./pkg/context
```

### Q: 安裝失敗?
```bash
# 確認在專案根目錄
pwd  # 應該在 claude-code-omystatusline/

# 檢查目錄結構
ls -la cmd/statusline  # 應該存在
ls -la go.mod          # 應該存在
```

### Q: make install 提示找不到專案目錄?
確保在專案根目錄執行,且 `cmd/statusline/` 和 `go.mod` 存在。

---

## 🎯 開發工作流程

### 1. 建立功能分支
```bash
git checkout -b feature/new-feature
```

### 2. 開發新功能
- 在 `pkg/` 下建立或修改模組
- 在對應套件中新增測試
- 更新 `cmd/statusline/main.go` 整合

### 3. 執行測試
```bash
go test ./...
```

### 4. 編譯驗證
```bash
make build
```

### 5. 提交變更
```bash
git add .
git commit -m "feat: 新增 XXX 功能"
```

---

## 💡 實用技巧

### 快速測試單一套件
```bash
# 測試 context 套件
go test -v ./pkg/context

# 測試 statusline 套件
go test -v ./pkg/statusline
```

### 檢查測試覆蓋率
```bash
go test -cover ./pkg/context
go test -cover ./pkg/statusline
```

### 執行特定測試
```bash
go test -v -run TestFormatModel ./pkg/statusline
go test -v -run TestGetColor ./pkg/context
```

### 查看詳細編譯資訊
```bash
go build -v -o statusline-go ./cmd/statusline
```

---

## 🔗 相關連結

- **GitHub Repo**: https://github.com/howie/claude-code-omystatusline
- **Issues**: https://github.com/howie/claude-code-omystatusline/issues
- **Claude Code Docs**: https://docs.claude.com/claude-code

---

## 📞 需要幫助?

- 查看 [docs/guides/architecture.md](docs/guides/architecture.md) 了解架構
- 查看 [docs/guides/migration.md](docs/guides/migration.md) 了解如何遷移
- 開 GitHub Issue 回報問題或建議

---

**Happy Coding! 🎉**
