# 遷移指南 / Migration Guide

## 重大變更：新目錄結構 / Breaking Change: New Directory Structure

### 版本 2.0.0

從 2.0.0 版本開始，所有 omystatusline 相關檔案已遷移到獨立的目錄結構。

**舊目錄結構 (v1.x):**
```
~/.claude/
├── statusline-go
├── statusline-wrapper.sh
├── statusline.sh
├── voice-reminder
├── voice-reminder-config.json
├── voice-reminder-enabled
├── voice-reminder-stats.json
├── voice-reminder-debug.log
├── play-notification.sh
├── scripts/
│   ├── toggle-voice-reminder.sh
│   └── test-voice-reminder.sh
└── commands/
    └── voice-reminder-*.md
```

**新目錄結構 (v2.0+):**
```
~/.claude/
├── omystatusline/                    # ← 新的主目錄
│   ├── bin/
│   │   ├── statusline-go
│   │   └── statusline-wrapper.sh
│   ├── scripts/
│   │   ├── statusline.sh
│   │   └── play-notification.sh
│   └── plugins/
│       └── voice-reminder/
│           ├── bin/voice-reminder
│           ├── config/voice-reminder-config.json
│           ├── scripts/...
│           ├── data/...
│           └── commands/...
├── commands/                         # ← 符號連結
│   └── voice-reminder-*.md -> ../omystatusline/plugins/...
└── config.json
```

## 如何升級 / How to Upgrade

### 自動升級 (推薦)

全新安裝會自動使用新目錄結構：

```bash
cd claude-code-omystatusline
git pull
make install
```

安裝程式會：
1. ✅ 自動在新位置安裝所有檔案
2. ✅ 建立符號連結以保持 slash commands 功能
3. ✅ 更新 `~/.claude/config.json` 使用新路徑

### 手動清理舊檔案

安裝後，你可以手動清理舊位置的檔案：

```bash
# 備份舊配置（如果你有自訂設定）
cp ~/.claude/voice-reminder-config.json ~/voice-reminder-config.backup.json

# 刪除舊檔案
rm -f ~/.claude/statusline-go
rm -f ~/.claude/statusline-wrapper.sh
rm -f ~/.claude/statusline.sh
rm -f ~/.claude/voice-reminder
rm -f ~/.claude/voice-reminder-config.json
rm -f ~/.claude/voice-reminder-enabled
rm -f ~/.claude/voice-reminder-stats.json
rm -f ~/.claude/voice-reminder-debug.log
rm -f ~/.claude/play-notification.sh
rm -rf ~/.claude/scripts
```

**注意**: 不要刪除 `~/.claude/commands/` 中的符號連結，它們現在指向新位置。

## 向後相容性 / Backward Compatibility

### 程式碼相容性

所有 Go 程式碼已包含向後相容邏輯：

- **配置讀取**: 優先使用新路徑，如果不存在則回退到舊路徑
- **資料儲存**: 新安裝使用新路徑，舊安裝繼續使用舊路徑
- **腳本**: 自動偵測並使用正確的路徑

這意味著：
- ✅ 舊版安裝仍可正常運作（不升級也沒問題）
- ✅ 新版安裝會自動使用新結構
- ✅ 混合使用期間不會出錯

### Config.json 路徑

更新後你的 `~/.claude/config.json` 應該變成：

**舊版:**
```json
{
  "statusLineCommand": "/Users/xxx/.claude/statusline-wrapper.sh",
  "hooks": {
    "Notification": [{
      "hooks": [{"command": "/Users/xxx/.claude/voice-reminder"}]
    }]
  }
}
```

**新版:**
```json
{
  "statusLineCommand": "~/.claude/omystatusline/bin/statusline-wrapper.sh",
  "hooks": {
    "Notification": [{
      "hooks": [{"command": "~/.claude/omystatusline/plugins/voice-reminder/bin/voice-reminder"}]
    }]
  }
}
```

## 修改配置檔 / Editing Configuration

### 語音提醒配置

**舊位置:**
```bash
~/.claude/voice-reminder-config.json
```

**新位置:**
```bash
~/.claude/omystatusline/plugins/voice-reminder/config/voice-reminder-config.json
```

編輯語音內容：
```bash
# 使用你喜歡的編輯器
vim ~/.claude/omystatusline/plugins/voice-reminder/config/voice-reminder-config.json

# 或使用 Claude Code 打開
code ~/.claude/omystatusline/plugins/voice-reminder/config/voice-reminder-config.json
```

## Slash Commands

所有 slash commands 仍然正常運作：

- `/voice-reminder-on` - 啟用語音提醒
- `/voice-reminder-off` - 停用語音提醒
- `/voice-reminder-stats` - 顯示統計
- `/voice-reminder-test` - 測試系統

它們透過符號連結繼續運作，無需任何改變。

## 疑難排解 / Troubleshooting

### 狀態列不顯示

檢查配置路徑：
```bash
cat ~/.claude/config.json | grep statusLineCommand
```

應該顯示：
```
"statusLineCommand": "~/.claude/omystatusline/bin/statusline-wrapper.sh"
```

### 語音不播放

檢查 binary 位置：
```bash
ls -la ~/.claude/omystatusline/plugins/voice-reminder/bin/voice-reminder
```

測試語音系統：
```bash
~/.claude/omystatusline/plugins/voice-reminder/scripts/test-voice-reminder.sh
```

### Slash commands 無法使用

檢查符號連結：
```bash
ls -la ~/.claude/commands/voice-reminder-*
```

應該看到指向 `../omystatusline/plugins/voice-reminder/commands/` 的符號連結。

## 需要協助？ / Need Help?

如果遇到問題：

1. 查看 [Voice Reminder Plugin README](docs/plugins/voice-reminder/README.md)
2. 提交 Issue: https://github.com/howie/claude-code-omystatusline/issues
3. 檢查 debug 日誌：`~/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-debug.log`

## 回滾到舊版 / Rollback to Old Version

如果需要回到舊版本：

```bash
cd claude-code-omystatusline
git checkout v1.x.x  # 替換為你之前的版本
make uninstall
make install
```

**注意**: v2.0+ 的 `make uninstall` 會清理新目錄結構的檔案。
