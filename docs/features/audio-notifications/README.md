# 工作完成音訊提醒功能 / Audio Notification for Work Completion

[English](#english) | [繁體中文](#chinese)

---

<a name="chinese"></a>

## 📢 功能概述

當使用 Claude Code 進行長時間工作時，你可能會切換到其他視窗或應用程式。此功能可以在以下情況發生時播放音訊提醒，讓你及時知道需要介入：

- ✅ **任務完成**：Claude 完成任務等待你的下一步指示
- ⚠️ **遇到錯誤**：需要你處理的錯誤或例外狀況
- 🔴 **接近限制**：Session 時間或 Token 使用量接近限制
- 💬 **等待輸入**：Claude 提出問題等待你的回應

## 🎯 為什麼需要這個功能？

在多工作業環境中：
- 你可能同時執行多個 Claude Code session
- 在等待 Claude 處理任務時切換到其他工作
- 長時間的程式碼產生或分析過程中離開螢幕
- 需要及時回應 Claude 的問題或確認請求

**音訊提醒確保你不會錯過任何需要介入的時刻。**

## 🔧 安裝與設定

### 快速安裝（推薦）

使用 omystatusline 的互動式安裝程式，可以輕鬆設定音訊提醒：

```bash
# 執行安裝程式
make install

# 或直接執行安裝腳本
./install.sh
```

安裝程式會詢問你：
1. ✅ 是否要安裝音訊提醒功能
2. 🔊 使用系統預設音效或自訂音訊檔案
3. 🗣️ 是否要開啟語音播報功能（TTS）

### 方案一：使用 Claude Code Hooks（推薦）

Claude Code 支援使用 hooks 在特定事件發生時執行自訂腳本。這是最簡單和最整合的方案。

#### 步驟 1: 建立音訊腳本

在 `~/.claude/` 目錄下建立一個播放音訊的腳本：

```bash
# 建立腳本檔案
cat > ~/.claude/play-notification.sh << 'EOF'
#!/bin/bash

# 根據作業系統選擇音訊播放工具
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    afplay /System/Library/Sounds/Glass.aiff
elif command -v paplay &> /dev/null; then
    # Linux with PulseAudio
    paplay /usr/share/sounds/freedesktop/stereo/complete.oga
elif command -v aplay &> /dev/null; then
    # Linux with ALSA
    aplay /usr/share/sounds/alsa/Front_Center.wav
elif command -v beep &> /dev/null; then
    # 使用系統蜂鳴器
    beep -f 800 -l 200
else
    # 使用終端機鈴聲作為備援方案
    echo -e '\a'
fi
EOF

# 新增執行權限
chmod +x ~/.claude/play-notification.sh
```

#### 步驟 2: 設定 Claude Code Hooks

編輯 `~/.claude/config.json` 新增 hook 設定：

```json
{
  "statusLineCommand": "~/.claude/statusline-wrapper.sh",
  "hooks": {
    "assistantMessageEnd": "~/.claude/play-notification.sh"
  }
}
```

**說明：**
- `assistantMessageEnd`：當 Claude 完成回覆時觸發
- 這樣每次 Claude 完成任務等待你的輸入時，都會播放音訊

#### 步驟 3: 測試

重新啟動 Claude Code 或開始新的對話，當 Claude 完成回覆時應該會聽到提示音。

### 方案二：智慧音訊提醒（進階）

如果你只想在特定情況下播放音訊（如遇到錯誤、接近限制等），可以建立一個更智慧的腳本：

#### 建立智慧提醒腳本

```bash
cat > ~/.claude/smart-notification.sh << 'EOF'
#!/bin/bash

# 讀取 Claude 的輸出
INPUT=$(cat)

# 檢查是否包含需要提醒的關鍵字
NEEDS_ATTENTION=false

# 檢查錯誤相關關鍵字
if echo "$INPUT" | grep -iE "error|failed|exception|cannot|unable|blocked" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 檢查問題或等待確認
if echo "$INPUT" | grep -iE "would you like|do you want|should I|please confirm|waiting for" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 檢查限制警告
if echo "$INPUT" | grep -E "🔴|🚨|⏰.*[0-9]+m" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 如果需要注意，播放音訊
if [ "$NEEDS_ATTENTION" = true ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga
    elif command -v aplay &> /dev/null; then
        aplay /usr/share/sounds/alsa/Front_Center.wav
    else
        echo -e '\a'
    fi
fi

# 將輸入原樣輸出（不影響正常流程）
echo "$INPUT"
EOF

chmod +x ~/.claude/smart-notification.sh
```

在 `~/.claude/config.json` 中使用：

```json
{
  "hooks": {
    "assistantMessageEnd": "~/.claude/smart-notification.sh"
  }
}
```

### 方案三：結合狀態列的進階提醒

你也可以修改狀態列腳本，在偵測到警告狀態時播放音訊。

#### 修改 statusline.go 新增音訊提醒

在 `~/.claude/statusline.go` 中新增音訊播放功能：

```go
// 在檔案開頭新增
import (
    "os/exec"
    // ... 其他 imports
)

// 新增播放音訊函式
func playNotificationSound() {
    // 非同步播放，不阻塞狀態列輸出
    go func() {
        var cmd *exec.Cmd

        // 根據系統選擇播放工具
        if _, err := exec.LookPath("afplay"); err == nil {
            cmd = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff")
        } else if _, err := exec.LookPath("paplay"); err == nil {
            cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
        } else if _, err := exec.LookPath("aplay"); err == nil {
            cmd = exec.Command("aplay", "/usr/share/sounds/alsa/Front_Center.wav")
        } else {
            // 終端機鈴聲
            fmt.Print("\a")
            return
        }

        if cmd != nil {
            _ = cmd.Run()
        }
    }()
}

// 在 main() 函式中，偵測到警告時呼叫
func main() {
    // ... 現有程式碼 ...

    // 在輸出狀態列之前檢查是否需要提醒
    needsAlert := false

    // 檢查 context 使用率
    if percentage >= 80 {
        needsAlert = true
    }

    // 檢查 session 時間（如果實作了限制警告功能）
    // if sessionTimeRemaining < 30 {
    //     needsAlert = true
    // }

    if needsAlert {
        playNotificationSound()
    }

    // ... 輸出狀態列 ...
}
```

## 🎵 自訂音訊檔案

### 使用自訂音訊檔案

你可以使用任何音訊檔案作為提醒音：

```bash
# 下載或準備你喜歡的音訊檔案（.wav, .mp3, .ogg, .aiff 等）
# 例如：
curl -o ~/.claude/notification.mp3 "https://example.com/your-sound.mp3"

# 修改腳本使用自訂檔案
cat > ~/.claude/play-notification.sh << 'EOF'
#!/bin/bash

SOUND_FILE="$HOME/.claude/notification.mp3"

if [[ "$OSTYPE" == "darwin"* ]]; then
    afplay "$SOUND_FILE"
elif command -v ffplay &> /dev/null; then
    ffplay -nodisp -autoexit "$SOUND_FILE" 2>/dev/null
elif command -v mpg123 &> /dev/null; then
    mpg123 -q "$SOUND_FILE"
elif command -v paplay &> /dev/null && command -v ffmpeg &> /dev/null; then
    ffmpeg -i "$SOUND_FILE" -f wav - 2>/dev/null | paplay
else
    echo -e '\a'
fi
EOF

chmod +x ~/.claude/play-notification.sh
```

### 推薦的音訊檔案來源

1. **系統內建音訊**（已包含在範例腳本中）
   - macOS: `/System/Library/Sounds/`
   - Linux: `/usr/share/sounds/`

2. **免費音效網站**
   - [FreeSound.org](https://freesound.org/)
   - [Notification Sounds](https://notificationsounds.com/)
   - [Zapsplat](https://www.zapsplat.com/)

3. **建立自己的提示音**
   - 使用 Audacity 等工具錄製或編輯
   - 保持簡短（1-2 秒）
   - 音量適中，不刺耳

## 🎚️ 音量控制

### 調整系統音量

確保你的系統音量設定合適：

```bash
# macOS
osascript -e "set volume output volume 50"  # 設定為 50%

# Linux (PulseAudio)
pactl set-sink-volume @DEFAULT_SINK@ 50%

# Linux (ALSA)
amixer set Master 50%
```

### 在腳本中控制音量

```bash
# macOS - 使用 afplay 時暫時調整音量
osascript -e "set volume output volume 30"
afplay /System/Library/Sounds/Glass.aiff
osascript -e "set volume output volume 50"  # 還原原音量

# Linux - 使用 paplay 時調整音量
paplay --volume=32768 /usr/share/sounds/freedesktop/stereo/complete.oga
# 注意：32768 是 50% 音量（最大值是 65536）
```

## 🔍 疑難排解

### 問題：沒有聽到音訊

1. **檢查音訊工具是否安裝**
   ```bash
   # 檢查可用的播放工具
   which afplay paplay aplay beep ffplay mpg123
   ```

2. **測試音訊檔案**
   ```bash
   # 手動執行腳本測試
   ~/.claude/play-notification.sh
   ```

3. **檢查系統音量**
   ```bash
   # 確保沒有靜音
   # macOS: 檢查「系統偏好設定」>「聲音」
   # Linux: alsamixer 或系統音量設定
   ```

4. **檢查檔案權限**
   ```bash
   ls -l ~/.claude/play-notification.sh
   # 應該顯示 -rwxr-xr-x（可執行）
   ```

### 問題：音訊播放但很刺耳

- 降低系統音量或在腳本中調整音量
- 選擇更柔和的音訊檔案
- 使用漸進式音效（fade in）

### 問題：Hook 沒有觸發

1. **驗證 config.json 格式**
   ```bash
   cat ~/.claude/config.json | python3 -m json.tool
   # 應該沒有語法錯誤
   ```

2. **檢查 Claude Code 版本**
   ```bash
   claude --version
   # 確保支援 hooks 功能
   ```

3. **查看日誌**
   ```bash
   # 檢查 Claude Code 的日誌輸出
   # 可能在 ~/.claude/logs/ 或終端機輸出中
   ```

### 問題：Linux 下沒有可用的音訊播放工具

安裝音訊播放工具：

```bash
# Ubuntu/Debian
sudo apt-get install pulseaudio-utils alsa-utils beep

# Fedora/RHEL
sudo dnf install pulseaudio-utils alsa-utils beep

# Arch Linux
sudo pacman -S pulseaudio alsa-utils beep
```

## 🎨 進階自訂

### 不同事件使用不同音訊

建立一個更複雜的腳本，根據訊息內容播放不同的音訊：

```bash
cat > ~/.claude/smart-sounds.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# 預設音訊
SOUND="default"

# 偵測錯誤
if echo "$INPUT" | grep -iE "error|failed|exception" > /dev/null; then
    SOUND="error"
fi

# 偵測完成
if echo "$INPUT" | grep -iE "completed|finished|done|success" > /dev/null; then
    SOUND="success"
fi

# 偵測警告
if echo "$INPUT" | grep -E "🔴|🚨|⚠️" > /dev/null; then
    SOUND="warning"
fi

# 播放對應的音訊
case "$SOUND" in
    error)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Basso.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/dialog-error.oga
        fi
        ;;
    success)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Glass.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/complete.oga
        fi
        ;;
    warning)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Ping.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/dialog-warning.oga
        fi
        ;;
    *)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Hero.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/message.oga
        fi
        ;;
esac

echo "$INPUT"
EOF

chmod +x ~/.claude/smart-sounds.sh
```

### 新增語音播報（Text-to-Speech）

```bash
cat > ~/.claude/voice-notification.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# 提取關鍵資訊並語音播報
if echo "$INPUT" | grep -iE "error|failed" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "任務失敗，請檢查"
    elif command -v espeak &> /dev/null; then
        espeak "Task failed, please check" 2>/dev/null
    fi
elif echo "$INPUT" | grep -iE "completed|finished" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "任務完成"
    elif command -v espeak &> /dev/null; then
        espeak "Task completed" 2>/dev/null
    fi
fi

echo "$INPUT"
EOF

chmod +x ~/.claude/voice-notification.sh
```

## 📊 最佳實務

1. **避免過度提醒**
   - 不要在每個小任務完成時都播放音訊
   - 只在真正需要注意的情況下提醒

2. **選擇合適的音量**
   - 足夠大以引起注意，但不要打擾他人
   - 在開放辦公環境考慮使用耳機

3. **使用不同的音訊**
   - 錯誤用低沉的音效
   - 完成用愉快的音效
   - 警告用中性的音效

4. **考慮工作時間**
   - 可以新增時間檢查，夜間自動靜音
   ```bash
   HOUR=$(date +%H)
   if [ $HOUR -ge 22 ] || [ $HOUR -le 7 ]; then
       # 夜間不播放音訊
       exit 0
   fi
   ```

5. **提供關閉開關**
   ```bash
   # 建立設定檔
   NOTIFICATION_ENABLED=$(cat ~/.claude/notification-enabled 2>/dev/null || echo "true")
   if [ "$NOTIFICATION_ENABLED" != "true" ]; then
       exit 0
   fi

   # 快速關閉/開啟
   # echo "false" > ~/.claude/notification-enabled  # 關閉
   # echo "true" > ~/.claude/notification-enabled   # 開啟
   ```

## 🔗 相關資源

- [Claude Code 官方文件](https://docs.anthropic.com/claude/docs)
- [Claude Code Hooks 文件](https://docs.anthropic.com/claude/docs/hooks)
- [Linux 音訊系統指南](https://wiki.archlinux.org/title/Sound_system)
- [macOS 命令列音訊播放](https://ss64.com/osx/afplay.html)

---

<a name="english"></a>

## 📢 Feature Overview

When working with Claude Code for extended periods, you might switch to other windows or applications. This feature can play audio notifications in the following situations to ensure you're promptly notified:

- ✅ **Task Completed**: Claude has finished a task and is waiting for your next instruction
- ⚠️ **Error Encountered**: An error or exception that requires your attention
- 🔴 **Approaching Limits**: Session time or token usage approaching limits
- 💬 **Awaiting Input**: Claude has asked a question and is waiting for your response

## 🎯 Why Do You Need This?

In a multitasking work environment:
- You may run multiple Claude Code sessions simultaneously
- You switch to other work while waiting for Claude to process tasks
- You step away during long code generation or analysis processes
- You need to respond promptly to Claude's questions or confirmation requests

**Audio notifications ensure you never miss a moment that requires your intervention.**

## 🔧 Installation & Configuration

### Quick Install (Recommended)

Use omystatusline's interactive installer to easily set up audio notifications:

```bash
# Run the installer
make install

# Or run the install script directly
./install.sh
```

The installer will ask you:
1. ✅ Whether to install audio notification features
2. 🔊 Use system default sounds or custom audio files
3. 🗣️ Whether to enable text-to-speech (TTS) functionality

### Option 1: Using Claude Code Hooks (Recommended)

Claude Code supports hooks to execute custom scripts when specific events occur. This is the simplest and most integrated approach.

#### Step 1: Create Sound Script

Create a sound playback script in `~/.claude/`:

```bash
# Create script file
cat > ~/.claude/play-notification.sh << 'EOF'
#!/bin/bash

# Choose sound playback tool based on OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    afplay /System/Library/Sounds/Glass.aiff
elif command -v paplay &> /dev/null; then
    # Linux with PulseAudio
    paplay /usr/share/sounds/freedesktop/stereo/complete.oga
elif command -v aplay &> /dev/null; then
    # Linux with ALSA
    aplay /usr/share/sounds/alsa/Front_Center.wav
elif command -v beep &> /dev/null; then
    # Use system beeper
    beep -f 800 -l 200
else
    # Fall back to terminal bell
    echo -e '\a'
fi
EOF

# Add execute permission
chmod +x ~/.claude/play-notification.sh
```

#### Step 2: Configure Claude Code Hooks

Edit `~/.claude/config.json` to add hook configuration:

```json
{
  "statusLineCommand": "~/.claude/statusline-wrapper.sh",
  "hooks": {
    "assistantMessageEnd": "~/.claude/play-notification.sh"
  }
}
```

**Explanation:**
- `assistantMessageEnd`: Triggers when Claude completes a response
- This plays a sound every time Claude finishes a task and waits for your input

#### Step 3: Test

Restart Claude Code or start a new conversation. You should hear a notification sound when Claude completes a response.

### Option 2: Smart Audio Notifications (Advanced)

If you only want to play sounds in specific situations (errors, approaching limits, etc.), create a more intelligent script:

#### Create Smart Notification Script

```bash
cat > ~/.claude/smart-notification.sh << 'EOF'
#!/bin/bash

# Read Claude's output
INPUT=$(cat)

# Check if notification is needed
NEEDS_ATTENTION=false

# Check for error-related keywords
if echo "$INPUT" | grep -iE "error|failed|exception|cannot|unable|blocked" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# Check for questions or confirmation requests
if echo "$INPUT" | grep -iE "would you like|do you want|should I|please confirm|waiting for" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# Check for limit warnings
if echo "$INPUT" | grep -E "🔴|🚨|⏰.*[0-9]+m" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# Play sound if attention needed
if [ "$NEEDS_ATTENTION" = true ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga
    elif command -v aplay &> /dev/null; then
        aplay /usr/share/sounds/alsa/Front_Center.wav
    else
        echo -e '\a'
    fi
fi

# Pass through input unchanged
echo "$INPUT"
EOF

chmod +x ~/.claude/smart-notification.sh
```

Use in `~/.claude/config.json`:

```json
{
  "hooks": {
    "assistantMessageEnd": "~/.claude/smart-notification.sh"
  }
}
```

### Option 3: Advanced Notifications with Status Line Integration

You can also modify the status line script to play sounds when warning states are detected.

#### Modify statusline.go to Add Audio Notifications

Add sound playback functionality to `~/.claude/statusline.go`:

```go
// Add at the beginning of the file
import (
    "os/exec"
    // ... other imports
)

// Add sound playback function
func playNotificationSound() {
    // Play asynchronously, don't block status line output
    go func() {
        var cmd *exec.Cmd

        // Choose playback tool based on system
        if _, err := exec.LookPath("afplay"); err == nil {
            cmd = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff")
        } else if _, err := exec.LookPath("paplay"); err == nil {
            cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
        } else if _, err := exec.LookPath("aplay"); err == nil {
            cmd = exec.Command("aplay", "/usr/share/sounds/alsa/Front_Center.wav")
        } else {
            // Terminal bell
            fmt.Print("\a")
            return
        }

        if cmd != nil {
            _ = cmd.Run()
        }
    }()
}

// In main() function, call when warning detected
func main() {
    // ... existing code ...

    // Check if alert needed before outputting status line
    needsAlert := false

    // Check context usage
    if percentage >= 80 {
        needsAlert = true
    }

    // Check session time (if limit warning feature is implemented)
    // if sessionTimeRemaining < 30 {
    //     needsAlert = true
    // }

    if needsAlert {
        playNotificationSound()
    }

    // ... output status line ...
}
```

## 🎵 Custom Sound Files

### Using Custom Audio Files

You can use any audio file as a notification sound:

```bash
# Download or prepare your preferred audio file (.wav, .mp3, .ogg, .aiff, etc.)
# For example:
curl -o ~/.claude/notification.mp3 "https://example.com/your-sound.mp3"

# Modify script to use custom file
cat > ~/.claude/play-notification.sh << 'EOF'
#!/bin/bash

SOUND_FILE="$HOME/.claude/notification.mp3"

if [[ "$OSTYPE" == "darwin"* ]]; then
    afplay "$SOUND_FILE"
elif command -v ffplay &> /dev/null; then
    ffplay -nodisp -autoexit "$SOUND_FILE" 2>/dev/null
elif command -v mpg123 &> /dev/null; then
    mpg123 -q "$SOUND_FILE"
elif command -v paplay &> /dev/null && command -v ffmpeg &> /dev/null; then
    ffmpeg -i "$SOUND_FILE" -f wav - 2>/dev/null | paplay
else
    echo -e '\a'
fi
EOF

chmod +x ~/.claude/play-notification.sh
```

### Recommended Sound File Sources

1. **Built-in System Sounds** (included in example scripts)
   - macOS: `/System/Library/Sounds/`
   - Linux: `/usr/share/sounds/`

2. **Free Sound Effect Websites**
   - [FreeSound.org](https://freesound.org/)
   - [Notification Sounds](https://notificationsounds.com/)
   - [Zapsplat](https://www.zapsplat.com/)

3. **Create Your Own**
   - Use tools like Audacity to record or edit
   - Keep it short (1-2 seconds)
   - Moderate volume, not jarring

## 🎚️ Volume Control

### Adjust System Volume

Ensure your system volume is set appropriately:

```bash
# macOS
osascript -e "set volume output volume 50"  # Set to 50%

# Linux (PulseAudio)
pactl set-sink-volume @DEFAULT_SINK@ 50%

# Linux (ALSA)
amixer set Master 50%
```

### Control Volume in Script

```bash
# macOS - temporarily adjust volume with afplay
osascript -e "set volume output volume 30"
afplay /System/Library/Sounds/Glass.aiff
osascript -e "set volume output volume 50"  # Restore original volume

# Linux - adjust volume with paplay
paplay --volume=32768 /usr/share/sounds/freedesktop/stereo/complete.oga
# Note: 32768 is 50% volume (max is 65536)
```

## 🔍 Troubleshooting

### Issue: No Sound

1. **Check if audio tools are installed**
   ```bash
   # Check available playback tools
   which afplay paplay aplay beep ffplay mpg123
   ```

2. **Test sound file**
   ```bash
   # Run script manually to test
   ~/.claude/play-notification.sh
   ```

3. **Check system volume**
   ```bash
   # Make sure not muted
   # macOS: System Preferences > Sound
   # Linux: alsamixer or system volume settings
   ```

4. **Check file permissions**
   ```bash
   ls -l ~/.claude/play-notification.sh
   # Should show -rwxr-xr-x (executable)
   ```

### Issue: Sound Plays But Is Jarring

- Lower system volume or adjust volume in script
- Choose a softer sound file
- Use sounds with fade-in effect

### Issue: Hook Not Triggering

1. **Verify config.json format**
   ```bash
   cat ~/.claude/config.json | python3 -m json.tool
   # Should have no syntax errors
   ```

2. **Check Claude Code version**
   ```bash
   claude --version
   # Ensure hooks feature is supported
   ```

3. **Check logs**
   ```bash
   # Check Claude Code log output
   # May be in ~/.claude/logs/ or terminal output
   ```

### Issue: No Audio Playback Tools on Linux

Install audio playback tools:

```bash
# Ubuntu/Debian
sudo apt-get install pulseaudio-utils alsa-utils beep

# Fedora/RHEL
sudo dnf install pulseaudio-utils alsa-utils beep

# Arch Linux
sudo pacman -S pulseaudio alsa-utils beep
```

## 🎨 Advanced Customization

### Different Sounds for Different Events

Create a more complex script that plays different sounds based on message content:

```bash
cat > ~/.claude/smart-sounds.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# Default sound
SOUND="default"

# Detect errors
if echo "$INPUT" | grep -iE "error|failed|exception" > /dev/null; then
    SOUND="error"
fi

# Detect completion
if echo "$INPUT" | grep -iE "completed|finished|done|success" > /dev/null; then
    SOUND="success"
fi

# Detect warnings
if echo "$INPUT" | grep -E "🔴|🚨|⚠️" > /dev/null; then
    SOUND="warning"
fi

# Play corresponding sound
case "$SOUND" in
    error)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Basso.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/dialog-error.oga
        fi
        ;;
    success)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Glass.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/complete.oga
        fi
        ;;
    warning)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Ping.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/dialog-warning.oga
        fi
        ;;
    *)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            afplay /System/Library/Sounds/Hero.aiff
        else
            paplay /usr/share/sounds/freedesktop/stereo/message.oga
        fi
        ;;
esac

echo "$INPUT"
EOF

chmod +x ~/.claude/smart-sounds.sh
```

### Add Voice Announcements (Text-to-Speech)

```bash
cat > ~/.claude/voice-notification.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# Extract key information and announce
if echo "$INPUT" | grep -iE "error|failed" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "Task failed, please check"
    elif command -v espeak &> /dev/null; then
        espeak "Task failed, please check" 2>/dev/null
    fi
elif echo "$INPUT" | grep -iE "completed|finished" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "Task completed"
    elif command -v espeak &> /dev/null; then
        espeak "Task completed" 2>/dev/null
    fi
fi

echo "$INPUT"
EOF

chmod +x ~/.claude/voice-notification.sh
```

## 📊 Best Practices

1. **Avoid Over-Notification**
   - Don't play sounds for every small task completion
   - Only notify when attention is truly needed

2. **Choose Appropriate Volume**
   - Loud enough to get attention, but not disturb others
   - Consider using headphones in open office environments

3. **Use Different Sounds**
   - Low-pitched sounds for errors
   - Pleasant sounds for completion
   - Neutral sounds for warnings

4. **Consider Working Hours**
   - Add time checks to auto-mute at night
   ```bash
   HOUR=$(date +%H)
   if [ $HOUR -ge 22 ] || [ $HOUR -le 7 ]; then
       # Don't play sound at night
       exit 0
   fi
   ```

5. **Provide On/Off Switch**
   ```bash
   # Create config file
   NOTIFICATION_ENABLED=$(cat ~/.claude/notification-enabled 2>/dev/null || echo "true")
   if [ "$NOTIFICATION_ENABLED" != "true" ]; then
       exit 0
   fi

   # Quick toggle
   # echo "false" > ~/.claude/notification-enabled  # Disable
   # echo "true" > ~/.claude/notification-enabled   # Enable
   ```

## 🔗 Related Resources

- [Claude Code Official Documentation](https://docs.anthropic.com/claude/docs)
- [Claude Code Hooks Documentation](https://docs.anthropic.com/claude/docs/hooks)
- [Linux Audio System Guide](https://wiki.archlinux.org/title/Sound_system)
- [macOS Command Line Audio Playback](https://ss64.com/osx/afplay.html)
