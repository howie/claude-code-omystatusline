#!/bin/bash
# 自訂 TTS 語音提醒範例
# Custom TTS Voice Notification Example

# 讀取 Claude 的輸出
INPUT=$(cat)

# ============================================================================
# JSON 解析 - 提取 message 欄位
# ============================================================================

# 從 JSON 提取 message 欄位
MESSAGE=$(echo "$INPUT" | jq -r '.message // ""' 2>/dev/null)

# 如果 jq 不可用或失敗，使用 grep/sed 作為備援
if [ -z "$MESSAGE" ] || [ "$MESSAGE" = "null" ]; then
    MESSAGE=$(echo "$INPUT" | grep -o '"message"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"message"[[:space:]]*:[[:space:]]*"\(.*\)"/\1/' 2>/dev/null)
fi

# 如果還是提取不到，使用整個輸入
if [ -z "$MESSAGE" ]; then
    MESSAGE="$INPUT"
fi

# ============================================================================
# 自訂語音訊息 - 根據不同情境播報不同內容
# ============================================================================

# 檢測需要確認的問題（優先級最高）
if echo "$MESSAGE" | grep "?" > /dev/null || \
   echo "$MESSAGE" | grep -iE "permission|confirm|approve" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "Claude 需要您的確認"
    elif command -v espeak &> /dev/null; then
        espeak "Claude needs your confirmation" 2>/dev/null
    fi

# 檢測編譯錯誤
elif echo "$MESSAGE" | grep -iE "compilation error|build failed" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "編譯失敗，請檢查程式碼"
    elif command -v espeak &> /dev/null; then
        espeak "Compilation failed, please check the code" 2>/dev/null
    fi

# 檢測測試失敗
elif echo "$MESSAGE" | grep -iE "test failed|tests? failing" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "測試未通過，需要修正"
    elif command -v espeak &> /dev/null; then
        espeak "Tests failed, need to fix" 2>/dev/null
    fi

# 檢測部署成功
elif echo "$MESSAGE" | grep -iE "deployed|deployment successful" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "部署完成，上線成功"
    elif command -v espeak &> /dev/null; then
        espeak "Deployment successful, now live" 2>/dev/null
    fi

# 檢測 Git 操作
elif echo "$MESSAGE" | grep -iE "pushed to|committed|pull request created" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "Git 操作完成"
    elif command -v espeak &> /dev/null; then
        espeak "Git operation completed" 2>/dev/null
    fi

# 檢測 Token 警告（80% 以上）
elif echo "$MESSAGE" | grep -E "🔴.*[8-9][0-9]%|🔴.*100%" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "注意！Token 使用量接近上限"
    elif command -v espeak &> /dev/null; then
        espeak "Warning! Token usage approaching limit" 2>/dev/null
    fi

# 一般錯誤
elif echo "$MESSAGE" | grep -iE "error|failed" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "發生錯誤"
    elif command -v espeak &> /dev/null; then
        espeak "Error occurred" 2>/dev/null
    fi

# 任務完成
elif echo "$MESSAGE" | grep -iE "completed|finished|done" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "任務完成"
    elif command -v espeak &> /dev/null; then
        espeak "Task completed" 2>/dev/null
    fi

# 其他情況 - 播放音效而非語音
else
    if [[ "$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null
    else
        echo -e '\a'
    fi
fi

# 必須將原始輸出傳遞回去，否則 Claude Code 無法顯示內容
echo "$INPUT"
