# Voice Reminder Plugin

一個為 Claude Code 提供語音提醒功能的插件，當重要事件發生時（確認、錯誤、完成）播放語音通知。

## 功能特色

- ✅ **多語言支援**: 支援英文和中文語音播報
- ✅ **事件監聽**: 監聽三種 Claude Code hook 事件（Notification, Stop, SubagentStop）
- ✅ **超時保護**: 語音播放有 10 秒超時機制
- ✅ **自動重試**: 失敗時自動重試一次
- ✅ **備援方案**: 語音失敗時回退到系統提示音
- ✅ **統計追蹤**: 記錄每種事件的觸發次數
- ✅ **Debug 日誌**: 提供詳細的除錯日誌功能

## 安裝位置

插件安裝在 `~/.claude/omystatusline/plugins/voice-reminder/` 目錄下：

```
~/.claude/omystatusline/plugins/voice-reminder/
├── bin/
│   └── voice-reminder           # 主程式
├── config/
│   └── voice-reminder-config.json   # 配置檔案
├── scripts/
│   ├── toggle-voice-reminder.sh     # 啟用/停用腳本
│   └── test-voice-reminder.sh       # 測試腳本
├── data/                             # 運行時資料
│   ├── voice-reminder-enabled        # 啟用狀態
│   ├── voice-reminder-stats.json     # 統計資料
│   └── voice-reminder-debug.log      # Debug 日誌
└── commands/                         # Slash commands
    ├── voice-reminder-on.md
    ├── voice-reminder-off.md
    ├── voice-reminder-stats.md
    └── voice-reminder-test.md
```

## 配置檔案

編輯 `~/.claude/omystatusline/plugins/voice-reminder/config/voice-reminder-config.json` 來自訂設定：

```json
{
  "debug_mode": false,
  "language": "zh",
  "speed": 180,
  "messages": {
    "notification": {
      "confirmation": ["Claude 需要您的確認", "主人請確認", "需要您做個決定"],
      "error": "任務失敗，請檢查",
      "completed": ["任務完成", "工作完成了"],
      "default": "請注意"
    },
    "stop": {
      "default": ["Claude 回應完成", "我說完了", "等待您的指示"]
    },
    "subagent_stop": {
      "default": "子任務已完成"
    }
  }
}
```

### 配置選項說明

- **`debug_mode`**: 啟用 debug 日誌（預設：false）
- **`language`**: 語言設定（`"zh"` 中文或 `"en"` 英文）
- **`speed`**: 語音速度，每分鐘字數（預設：180）
- **`messages`**: 各事件的語音訊息
  - 陣列形式會隨機選擇一個播放
  - 字串形式會固定播放該訊息

## Slash Commands

### `/voice-reminder-on`
啟用語音提醒功能

### `/voice-reminder-off`
停用語音提醒功能（靜音）

### `/voice-reminder-stats`
顯示使用統計資訊

### `/voice-reminder-test`
測試語音系統是否正常運作

## 手動控制

### 啟用/停用
```bash
~/.claude/omystatusline/plugins/voice-reminder/scripts/toggle-voice-reminder.sh on   # 啟用
~/.claude/omystatusline/plugins/voice-reminder/scripts/toggle-voice-reminder.sh off  # 停用
```

### 測試語音
```bash
~/.claude/omystatusline/plugins/voice-reminder/scripts/test-voice-reminder.sh
```

## 統計資料

統計資料儲存在 `~/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-stats.json`：

```json
{
  "notification_count": 15,
  "stop_count": 42,
  "subagent_stop_count": 8,
  "last_triggered": "2025-10-21T21:49:01+08:00"
}
```

## Debug 日誌

啟用 debug 模式後，日誌會寫入 `~/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-debug.log`：

```
[2025-10-21 21:49:01.132] ========== Hook 觸發 ==========
[2025-10-21 21:49:01.132] 收到的原始 JSON: {"message": "..."}
[2025-10-21 21:49:01.132] 解析結果 - EventName: Notification
[2025-10-21 21:49:01.132] 開始播放語音...
[2025-10-21 21:49:01.132] 語音播放成功
[2025-10-21 21:49:01.132] ========== 處理完成 ==========
```

## 技術細節

### 支援的 Hook 事件

1. **Notification**: 當 Claude 需要用戶確認或回應時觸發
2. **Stop**: 當 Claude 完成回應時觸發
3. **SubagentStop**: 當子代理（subagent）任務完成時觸發

### 語音播放機制

- **macOS**: 使用內建的 `say` 命令
- **Linux**: 使用 `espeak` 或 `espeak-ng`（需要另行安裝）
- **備援**: 系統提示音（Glass.aiff）

### 超時與重試

- 語音播放有 10 秒超時保護
- 失敗時自動重試一次
- 兩次都失敗則播放備援提示音

## 疑難排解

### 語音不播放
1. 檢查是否已啟用：`cat ~/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-enabled`
2. 測試語音系統：`/voice-reminder-test`
3. 查看 debug 日誌：啟用 `debug_mode` 並檢查日誌檔案

### Linux 語音失敗
```bash
# Ubuntu/Debian
sudo apt-get install espeak espeak-ng

# Fedora
sudo dnf install espeak

# Arch
sudo pacman -S espeak
```

## 向後相容

此插件支援舊版安裝路徑（`~/.claude/voice-reminder*`）的自動回退機制，確保升級後仍能正常運作。

## 授權

Apache License 2.0
