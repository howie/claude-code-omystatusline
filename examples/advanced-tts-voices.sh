#!/bin/bash
# 進階 TTS 範例 - 使用不同聲音和語速
# Advanced TTS Example - Different Voices and Speech Rates

INPUT=$(cat)

# ============================================================================
# macOS 可用的中文聲音 (使用 say -v ? | grep zh 查看完整列表)
# ============================================================================
# Ting-Ting (繁體中文，女性，自然)
# Sin-ji (繁體中文，女性)
# Mei-Jia (繁體中文，女性)

# ============================================================================
# 語速控制：-r 參數
# ============================================================================
# 預設: 200 words/min
# 快速: 300-400 words/min
# 慢速: 100-150 words/min

# 檢測緊急錯誤 - 使用快速語音
if echo "$INPUT" | grep -iE "critical|urgent|emergency|fatal" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # 使用 Ting-Ting 聲音，快速語速
        say -v Ting-Ting -r 250 "緊急！發現嚴重錯誤，請立即處理"
    elif command -v espeak &> /dev/null; then
        espeak -s 180 "Critical error, immediate action required" 2>/dev/null
    fi

# 檢測警告 - 使用正常語速
elif echo "$INPUT" | grep -iE "warning|caution" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say -v Ting-Ting -r 200 "注意，發現警告訊息"
    elif command -v espeak &> /dev/null; then
        espeak -s 150 "Warning detected" 2>/dev/null
    fi

# 檢測成功部署 - 使用愉快的語調
elif echo "$INPUT" | grep -iE "deployed successfully|deployment complete" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # 使用稍快的語速表示興奮
        say -v Ting-Ting -r 220 "太好了！部署成功"
    elif command -v espeak &> /dev/null; then
        espeak -s 160 "Great! Deployment successful" 2>/dev/null
    fi

# 檢測測試通過 - 所有測試通過
elif echo "$INPUT" | grep -iE "all tests passed|tests? successful" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say -v Ting-Ting -r 210 "所有測試通過，做得好"
    elif command -v espeak &> /dev/null; then
        espeak -s 155 "All tests passed, well done" 2>/dev/null
    fi

# 檢測需要輸入 - 使用慢速清晰發音
elif echo "$INPUT" | grep -iE "please confirm|need your input|waiting for you" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # 使用較慢語速，確保清楚聽到
        say -v Ting-Ting -r 180 "Claude 正在等待您的輸入"
    elif command -v espeak &> /dev/null; then
        espeak -s 140 "Claude is waiting for your input" 2>/dev/null
    fi

# 檢測 Token 使用警告
elif echo "$INPUT" | grep -E "🔴.*[8-9][0-9]%|🔴.*100%" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say -v Ting-Ting -r 190 "注意！Context 使用量已超過百分之八十"
    elif command -v espeak &> /dev/null; then
        espeak -s 145 "Warning! Context usage over 80 percent" 2>/dev/null
    fi

# 一般任務完成
elif echo "$INPUT" | grep -iE "completed|finished" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say -v Ting-Ting -r 200 "任務完成"
    elif command -v espeak &> /dev/null; then
        espeak -s 150 "Task completed" 2>/dev/null
    fi

# 其他 - 簡短提示音
else
    if [[ "$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null
    fi
fi

echo "$INPUT"
