# 工作完成声音提醒功能 / Audio Notification for Work Completion

[English](#english) | [中文](#chinese)

---

<a name="chinese"></a>

## 📢 功能概述

当使用 Claude Code 进行长时间工作时，你可能会切换到其他窗口或应用程序。此功能可以在以下情况发生时播放声音提醒，让你及时知道需要介入：

- ✅ **任务完成**：Claude 完成任务等待你的下一步指示
- ⚠️ **遇到错误**：需要你处理的错误或异常情况
- 🔴 **接近限制**：Session 时间或 Token 使用量接近限制
- 💬 **等待输入**：Claude 提出问题等待你的回应

## 🎯 为什么需要这个功能？

在多任务工作环境中：
- 你可能同时运行多个 Claude Code session
- 在等待 Claude 处理任务时切换到其他工作
- 长时间的代码生成或分析过程中离开屏幕
- 需要及时响应 Claude 的问题或确认请求

**声音提醒确保你不会错过任何需要介入的时刻。**

## 🔧 安装与配置

### 方案一：使用 Claude Code Hooks（推荐）

Claude Code 支持使用 hooks 在特定事件发生时执行自定义脚本。这是最简单和最集成的方案。

#### 步骤 1: 创建声音脚本

在 `~/.claude/` 目录下创建一个播放声音的脚本：

```bash
# 创建脚本文件
cat > ~/.claude/play-notification.sh << 'EOF'
#!/bin/bash

# 根据操作系统选择声音播放工具
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
    # 使用系统蜂鸣器
    beep -f 800 -l 200
else
    # 使用终端铃声作为后备方案
    echo -e '\a'
fi
EOF

# 添加执行权限
chmod +x ~/.claude/play-notification.sh
```

#### 步骤 2: 配置 Claude Code Hooks

编辑 `~/.claude/config.json` 添加 hook 配置：

```json
{
  "statusLineCommand": "~/.claude/statusline-wrapper.sh",
  "hooks": {
    "assistantMessageEnd": "~/.claude/play-notification.sh"
  }
}
```

**说明：**
- `assistantMessageEnd`：当 Claude 完成回复时触发
- 这样每次 Claude 完成任务等待你的输入时，都会播放声音

#### 步骤 3: 测试

重启 Claude Code 或开始新的对话，当 Claude 完成回复时应该会听到提示音。

### 方案二：智能声音提醒（高级）

如果你只想在特定情况下播放声音（如遇到错误、接近限制等），可以创建一个更智能的脚本：

#### 创建智能提醒脚本

```bash
cat > ~/.claude/smart-notification.sh << 'EOF'
#!/bin/bash

# 读取 Claude 的输出
INPUT=$(cat)

# 检查是否包含需要提醒的关键词
NEEDS_ATTENTION=false

# 检查错误相关关键词
if echo "$INPUT" | grep -iE "error|failed|exception|cannot|unable|blocked" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 检查问题或等待确认
if echo "$INPUT" | grep -iE "would you like|do you want|should I|please confirm|waiting for" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 检查限制警告
if echo "$INPUT" | grep -E "🔴|🚨|⏰.*[0-9]+m" > /dev/null; then
    NEEDS_ATTENTION=true
fi

# 如果需要注意，播放声音
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

# 将输入原样输出（不影响正常流程）
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

### 方案三：结合状态栏的高级提醒

你也可以修改状态栏脚本，在检测到警告状态时播放声音。

#### 修改 statusline.go 添加声音提醒

在 `~/.claude/statusline.go` 中添加声音播放功能：

```go
// 在文件开头添加
import (
    "os/exec"
    // ... 其他 imports
)

// 添加播放声音函数
func playNotificationSound() {
    // 异步播放，不阻塞状态栏输出
    go func() {
        var cmd *exec.Cmd

        // 根据系统选择播放工具
        if _, err := exec.LookPath("afplay"); err == nil {
            cmd = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff")
        } else if _, err := exec.LookPath("paplay"); err == nil {
            cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
        } else if _, err := exec.LookPath("aplay"); err == nil {
            cmd = exec.Command("aplay", "/usr/share/sounds/alsa/Front_Center.wav")
        } else {
            // 终端铃声
            fmt.Print("\a")
            return
        }

        if cmd != nil {
            _ = cmd.Run()
        }
    }()
}

// 在 main() 函数中，检测到警告时调用
func main() {
    // ... 现有代码 ...

    // 在输出状态栏之前检查是否需要提醒
    needsAlert := false

    // 检查 context 使用率
    if percentage >= 80 {
        needsAlert = true
    }

    // 检查 session 时间（如果实现了限制警告功能）
    // if sessionTimeRemaining < 30 {
    //     needsAlert = true
    // }

    if needsAlert {
        playNotificationSound()
    }

    // ... 输出状态栏 ...
}
```

## 🎵 自定义声音文件

### 使用自定义音频文件

你可以使用任何音频文件作为提醒音：

```bash
# 下载或准备你喜欢的音频文件（.wav, .mp3, .ogg, .aiff 等）
# 例如：
curl -o ~/.claude/notification.mp3 "https://example.com/your-sound.mp3"

# 修改脚本使用自定义文件
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
EOF
```

### 推荐的声音文件来源

1. **系统内置声音**（已包含在示例脚本中）
   - macOS: `/System/Library/Sounds/`
   - Linux: `/usr/share/sounds/`

2. **免费音效网站**
   - [FreeSound.org](https://freesound.org/)
   - [Notification Sounds](https://notificationsounds.com/)
   - [Zapsplat](https://www.zapsplat.com/)

3. **创建自己的提示音**
   - 使用 Audacity 等工具录制或编辑
   - 保持简短（1-2 秒）
   - 音量适中，不刺耳

## 🎚️ 音量控制

### 调整系统音量

确保你的系统音量设置合适：

```bash
# macOS
osascript -e "set volume output volume 50"  # 设置为 50%

# Linux (PulseAudio)
pactl set-sink-volume @DEFAULT_SINK@ 50%

# Linux (ALSA)
amixer set Master 50%
```

### 在脚本中控制音量

```bash
# macOS - 使用 afplay 时临时调整音量
osascript -e "set volume output volume 30"
afplay /System/Library/Sounds/Glass.aiff
osascript -e "set volume output volume 50"  # 恢复原音量

# Linux - 使用 paplay 时调整音量
paplay --volume=32768 /usr/share/sounds/freedesktop/stereo/complete.oga
# 注意：32768 是 50% 音量（最大值是 65536）
```

## 🔍 故障排除

### 问题：没有听到声音

1. **检查音频工具是否安装**
   ```bash
   # 检查可用的播放工具
   which afplay paplay aplay beep ffplay mpg123
   ```

2. **测试声音文件**
   ```bash
   # 手动运行脚本测试
   ~/.claude/play-notification.sh
   ```

3. **检查系统音量**
   ```bash
   # 确保没有静音
   # macOS: 检查系统偏好设置 > 声音
   # Linux: alsamixer 或系统音量设置
   ```

4. **检查文件权限**
   ```bash
   ls -l ~/.claude/play-notification.sh
   # 应该显示 -rwxr-xr-x（可执行）
   ```

### 问题：声音播放但很刺耳

- 降低系统音量或在脚本中调整音量
- 选择更柔和的声音文件
- 使用渐进式音效（fade in）

### 问题：Hook 没有触发

1. **验证 config.json 格式**
   ```bash
   cat ~/.claude/config.json | python3 -m json.tool
   # 应该没有语法错误
   ```

2. **检查 Claude Code 版本**
   ```bash
   claude --version
   # 确保支持 hooks 功能
   ```

3. **查看日志**
   ```bash
   # 检查 Claude Code 的日志输出
   # 可能在 ~/.claude/logs/ 或终端输出中
   ```

### 问题：Linux 下没有可用的声音播放工具

安装音频播放工具：

```bash
# Ubuntu/Debian
sudo apt-get install pulseaudio-utils alsa-utils beep

# Fedora/RHEL
sudo dnf install pulseaudio-utils alsa-utils beep

# Arch Linux
sudo pacman -S pulseaudio alsa-utils beep
```

## 🎨 高级自定义

### 不同事件使用不同声音

创建一个更复杂的脚本，根据消息内容播放不同的声音：

```bash
cat > ~/.claude/smart-sounds.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# 默认声音
SOUND="default"

# 检测错误
if echo "$INPUT" | grep -iE "error|failed|exception" > /dev/null; then
    SOUND="error"
fi

# 检测完成
if echo "$INPUT" | grep -iE "completed|finished|done|success" > /dev/null; then
    SOUND="success"
fi

# 检测警告
if echo "$INPUT" | grep -E "🔴|🚨|⚠️" > /dev/null; then
    SOUND="warning"
fi

# 播放对应的声音
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

### 添加语音播报（Text-to-Speech）

```bash
cat > ~/.claude/voice-notification.sh << 'EOF'
#!/bin/bash

INPUT=$(cat)

# 提取关键信息并语音播报
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

## 📊 最佳实践

1. **避免过度提醒**
   - 不要在每个小任务完成时都播放声音
   - 只在真正需要注意的情况下提醒

2. **选择合适的音量**
   - 足够大以引起注意，但不要打扰他人
   - 在开放办公环境考虑使用耳机

3. **使用不同的声音**
   - 错误用低沉的音效
   - 完成用愉快的音效
   - 警告用中性的音效

4. **考虑工作时间**
   - 可以添加时间检查，夜间自动静音
   ```bash
   HOUR=$(date +%H)
   if [ $HOUR -ge 22 ] || [ $HOUR -le 7 ]; then
       # 夜间不播放声音
       exit 0
   fi
   ```

5. **提供关闭开关**
   ```bash
   # 创建配置文件
   NOTIFICATION_ENABLED=$(cat ~/.claude/notification-enabled 2>/dev/null || echo "true")
   if [ "$NOTIFICATION_ENABLED" != "true" ]; then
       exit 0
   fi

   # 快速关闭/开启
   # echo "false" > ~/.claude/notification-enabled  # 关闭
   # echo "true" > ~/.claude/notification-enabled   # 开启
   ```

## 🔗 相关资源

- [Claude Code 官方文档](https://docs.anthropic.com/claude/docs)
- [Claude Code Hooks 文档](https://docs.anthropic.com/claude/docs/hooks)
- [Linux 音频系统指南](https://wiki.archlinux.org/title/Sound_system)
- [macOS 命令行音频播放](https://ss64.com/osx/afplay.html)

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
