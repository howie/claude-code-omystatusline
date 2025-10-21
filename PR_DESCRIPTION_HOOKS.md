# 🛡️ 新增 Pre-Push Git Hook 與品質檢查

## 📋 概述

此 PR 新增了 Git pre-push hook 系統，在推送前自動執行品質檢查，並修正了安裝程式的編譯檢查邏輯。

## ✨ 主要變更

### 1. 🛡️ Git Pre-Push Hook 系統

新增完整的 Git hooks 基礎設施：

**檔案結構：**
```
.githooks/
├── pre-push           # Pre-push hook 腳本
├── install-hooks.sh   # 自動化安裝腳本
└── README.md          # 完整使用文件
```

**五項自動檢查：**

1. ✅ **工作目錄狀態檢查**
   - 偵測未提交的變更
   - 提示使用者是否繼續推送

2. ✅ **Go 代碼編譯檢查**
   - 確保 Go 代碼可以成功編譯
   - 編譯失敗時阻止推送

3. ✅ **Go 測試執行**
   - 執行所有 Go 測試
   - 測試失敗時阻止推送

4. ✅ **Shell 腳本語法檢查**
   - 驗證 install.sh、statusline-wrapper.sh、statusline.sh
   - 語法錯誤時阻止推送

5. ✅ **安裝腳本測試**
   - 測試安裝腳本載入
   - 確保腳本基本可執行

### 2. 🔧 修正編譯檢查邏輯

**問題：**
```bash
# 原本的程式碼
if go build ... | grep -v "^#"; then
```

- `grep` 沒有匹配時返回退出碼 1
- 即使編譯成功也會判定失敗

**修正：**
```bash
# 修正後的程式碼
if go build ... 2>&1; then
```

- 直接檢查 `go build` 的退出碼
- 正確判斷編譯是否成功

### 3. 📦 Makefile 新增指令

```bash
make test              # 執行 Go 測試
make install-hooks     # 安裝 Git hooks
make uninstall-hooks   # 卸載 Git hooks
```

## 🎯 使用方式

### 安裝 Git Hooks

```bash
# 方法 1: 使用 Makefile
make install-hooks

# 方法 2: 直接執行安裝腳本
./.githooks/install-hooks.sh
```

### Hook 執行範例

```bash
$ git push origin branch-name

╔════════════════════════════════════════════════════════════════╗
║  Pre-Push Checks - Claude Code omystatusline              ║
╚════════════════════════════════════════════════════════════════╝

ℹ 檢查 1/5: 工作目錄狀態
✓ 工作目錄乾淨

ℹ 檢查 2/5: Go 代碼編譯
✓ Go 代碼編譯成功

ℹ 檢查 3/5: Go 測試執行
✓ Go 測試通過

ℹ 檢查 4/5: Shell 腳本語法檢查
✓ install.sh 語法正確
✓ statusline-wrapper.sh 語法正確
✓ statusline.sh 語法正確

ℹ 檢查 5/5: 安裝腳本乾跑測試
✓ 安裝腳本載入測試通過

════════════════════════════════════════════════════════════════
✓ 所有檢查通過！
════════════════════════════════════════════════════════════════
```

### 檢查失敗範例

```bash
════════════════════════════════════════════════════════════════
✗ 有 2 項檢查失敗
════════════════════════════════════════════════════════════════

請修正上述問題後再推送

是否強制推送？ [y/N]:
```

## 🎨 功能特色

✅ **自動化品質檢查**：推送前自動執行
✅ **友善的 UI**：彩色輸出和清晰的進度顯示
✅ **智慧跳過**：自動跳過不適用的檢查
✅ **強制推送選項**：允許在必要時強制推送
✅ **詳細文件**：完整的 README 和使用說明
✅ **易於安裝**：一行指令即可安裝
✅ **可擴展**：容易新增更多檢查項目

## 📝 檔案變更

### 新增檔案
- `.githooks/pre-push` - Pre-push hook 腳本
- `.githooks/install-hooks.sh` - Hook 安裝腳本
- `.githooks/README.md` - 完整使用文件

### 修改檔案
- `Makefile` - 新增 test、install-hooks、uninstall-hooks targets
- `install.sh` - 修正編譯檢查邏輯

## 🧪 測試計畫

### Hook 功能測試
- [x] 工作目錄乾淨時通過檢查
- [x] 有未提交變更時顯示警告
- [x] Go 編譯成功時通過
- [x] Go 編譯失敗時阻止推送
- [x] Shell 腳本語法正確時通過
- [x] 可以強制推送（需確認）

### 安裝測試
- [x] make install-hooks 正常運作
- [x] 手動安裝腳本正常運作
- [x] Hook 正確安裝到 .git/hooks/
- [x] Hook 有執行權限

### 編譯修正測試
- [x] 修正後編譯檢查正常運作
- [x] 編譯成功時不會誤判失敗

## 🔗 相關連結

- [Git Hooks 文件](.githooks/README.md)
- [Makefile 使用說明](Makefile)

## ⚠️ 破壞性變更

無破壞性變更。

- Git hooks 是選用功能，不影響現有工作流程
- 使用者需要手動執行 `make install-hooks` 安裝
- 可隨時使用 `make uninstall-hooks` 卸載

## 💡 最佳實務建議

1. **安裝 hooks**：建議所有開發者安裝 pre-push hook
2. **定期更新**：fetch 後重新安裝 hooks 以獲取最新版本
3. **不要繞過**：避免使用 `--no-verify` 繞過檢查
4. **修正問題**：檢查失敗時應修正問題而非強制推送

## 📊 Commits 摘要

1. `6f6c50e` - Fix compilation check in installer
   - 修正 install.sh 中的編譯檢查邏輯
   - 移除導致誤判的 grep 過濾

2. `363c71e` - Add pre-push Git hook for quality checks
   - 新增完整的 pre-push hook 系統
   - 實作五項品質檢查
   - 更新 Makefile 新增相關指令

## 🙏 後續改進

- [ ] 新增 commit-msg hook 驗證 commit 訊息格式
- [ ] 新增 pre-commit hook 進行程式碼格式檢查
- [ ] 整合 shellcheck 進行更嚴格的 shell 腳本檢查
- [ ] 新增 CI/CD 整合測試

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
