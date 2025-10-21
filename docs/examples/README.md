# TTS 自訂範例 / TTS Customization Examples

這個目錄包含各種自訂 TTS 語音提醒的範例腳本。

This directory contains example scripts for customizing TTS voice notifications.

---

## 📁 範例檔案 / Example Files

### 1. `custom-tts-notification.sh`

**功能**：根據不同任務類型播報不同訊息

**涵蓋場景**：
- 編譯錯誤
- 測試失敗
- 部署成功
- Git 操作
- 需要確認
- Token 警告
- 一般錯誤
- 任務完成

**使用方式**：
```bash
# 複製到 Claude 設定目錄
cp custom-tts-notification.sh ~/.claude/play-notification.sh
chmod +x ~/.claude/play-notification.sh

# 或建立符號連結
ln -sf $(pwd)/custom-tts-notification.sh ~/.claude/play-notification.sh
```

### 2. `advanced-tts-voices.sh`

**功能**：使用不同聲音和語速

**特色**：
- **緊急情況**：快速語音（250 words/min）
- **警告訊息**：正常語速（200 words/min）
- **成功訊息**：稍快語速（220 words/min）
- **需要輸入**：慢速清晰（180 words/min）

**macOS 聲音選項**：
```bash
# 查看所有可用的中文聲音
say -v ? | grep zh

# 常用中文聲音：
# - Ting-Ting (繁體中文，自然，推薦)
# - Sin-ji (繁體中文，女性)
# - Mei-Jia (繁體中文，女性)
```

**使用方式**：
```bash
cp advanced-tts-voices.sh ~/.claude/play-notification.sh
chmod +x ~/.claude/play-notification.sh
```

---

## 🎨 自訂指南 / Customization Guide

### 修改語音內容

編輯腳本中的 `say` 或 `espeak` 命令後的文字：

```bash
# macOS - 修改中文訊息
say -v Ting-Ting "您想要的訊息內容"

# Linux - 修改英文訊息
espeak "Your custom message here"
```

### 調整語速 (macOS)

```bash
# 快速 (適合緊急訊息)
say -v Ting-Ting -r 250 "緊急訊息"

# 正常 (預設)
say -v Ting-Ting -r 200 "一般訊息"

# 慢速 (適合重要訊息)
say -v Ting-Ting -r 150 "重要訊息"
```

### 調整語速 (Linux espeak)

```bash
# 快速
espeak -s 180 "Fast message"

# 正常
espeak -s 150 "Normal message"

# 慢速
espeak -s 120 "Slow message"
```

### 使用不同聲音 (macOS)

```bash
# 查看所有可用聲音
say -v ?

# 使用特定聲音
say -v Ting-Ting "繁體中文女聲"
say -v Sin-ji "繁體中文女聲（另一個選項）"
say -v Alex "英文男聲"
say -v Samantha "英文女聲"
```

### 添加新的關鍵字檢測

在腳本中添加新的 `if` 或 `elif` 區塊：

```bash
# 檢測自訂關鍵字
elif echo "$INPUT" | grep -iE "您的關鍵字1|關鍵字2" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say -v Ting-Ting "自訂訊息"
    elif command -v espeak &> /dev/null; then
        espeak "Custom message" 2>/dev/null
    fi
```

---

## 🔧 測試 TTS 設定

### 測試 macOS 聲音

```bash
# 測試繁體中文聲音
say -v Ting-Ting "這是測試訊息"

# 測試不同語速
say -v Ting-Ting -r 150 "慢速測試"
say -v Ting-Ting -r 200 "正常測試"
say -v Ting-Ting -r 250 "快速測試"

# 列出所有中文聲音
say -v ? | grep zh
```

### 測試 Linux espeak

```bash
# 測試英文語音
espeak "This is a test message"

# 測試不同語速
espeak -s 120 "Slow test"
espeak -s 150 "Normal test"
espeak -s 180 "Fast test"

# 查看 espeak 版本和選項
espeak --help
```

### 測試整個腳本

```bash
# 建立測試輸入
echo "Task completed successfully" | ~/.claude/play-notification.sh

# 測試錯誤訊息
echo "Error: compilation failed" | ~/.claude/play-notification.sh

# 測試警告訊息
echo "Warning: token usage at 85%" | ~/.claude/play-notification.sh
```

---

## 💡 使用技巧 / Tips

### 1. 避免過長的訊息

語音播報應該簡短，5-10 個字最佳：

❌ **不好**：
```bash
say "任務已經完成執行，現在正在等待您的下一步指示，請確認是否要繼續進行"
```

✅ **良好**：
```bash
say "任務完成，請確認"
```

### 2. 在安靜環境測試音量

```bash
# macOS - 調整音量
osascript -e "set volume output volume 50"

# Linux - 調整音量
pactl set-sink-volume @DEFAULT_SINK@ 50%
```

### 3. 根據時間調整行為

```bash
# 夜間使用音效而非語音
HOUR=$(date +%H)
if [ $HOUR -ge 22 ] || [ $HOUR -le 7 ]; then
    # 播放輕柔音效而非語音
    afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
else
    # 白天使用語音
    say -v Ting-Ting "任務完成"
fi
```

### 4. 添加除錯模式

在腳本開頭添加：

```bash
#!/bin/bash
# 除錯模式：取消註解以查看偵測到的內容
# DEBUG=true

INPUT=$(cat)

if [ "$DEBUG" = "true" ]; then
    echo "DEBUG: Input received: $INPUT" >> ~/.claude/tts-debug.log
fi
```

---

## 🌍 多語言支援

### macOS 支援的語言

```bash
# 繁體中文
say -v Ting-Ting "繁體中文測試"

# 簡體中文
say -v Ting-Ting "简体中文测试"

# 英文
say -v Samantha "English test"

# 日文
say -v Kyoko "日本語テスト"

# 韓文
say -v Yuna "한국어 테스트"
```

### Linux espeak 多語言

```bash
# 安裝語言包
sudo apt-get install espeak-data

# 使用不同語言
espeak -v zh "中文測試"
espeak -v en "English test"
espeak -v ja "日本語テスト"
```

---

## 📚 參考資源

### macOS `say` 命令

- **查看幫助**：`man say`
- **聲音列表**：`say -v ?`
- **參數說明**：
  - `-v VOICE`：選擇聲音
  - `-r RATE`：語速（words/min，預設 200）
  - `-o FILE`：輸出到檔案而非播放

### Linux `espeak` 命令

- **查看幫助**：`espeak --help`
- **主要參數**：
  - `-v VOICE`：選擇聲音/語言
  - `-s SPEED`：語速（預設 150）
  - `-p PITCH`：音調（0-99，預設 50）
  - `-a AMPLITUDE`：音量（0-200，預設 100）

---

## 🔗 相關文件

- [音訊提醒完整文件](../docs/features/audio-notifications/README.md)
- [主要 README](../README.md)
- [安裝指南](../docs/installation.md)

---

## 📝 貢獻您的範例

如果您創建了有用的 TTS 腳本，歡迎貢獻！

1. Fork 這個專案
2. 在 `examples/` 目錄添加您的腳本
3. 在這個 README 中添加說明
4. 提交 Pull Request

---

## ⚠️ 注意事項

1. **隱私**：所有 TTS 處理都在本地完成，不會發送到雲端
2. **效能**：語音播報是異步的，不會阻塞 Claude Code
3. **相依性**：
   - macOS：無需額外安裝
   - Linux：需要安裝 `espeak` 或 `espeak-ng`
4. **測試**：修改腳本後務必測試，確保不會影響正常輸出

---

**祝您使用愉快！ / Enjoy!**
