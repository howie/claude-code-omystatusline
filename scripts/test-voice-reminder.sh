#!/bin/bash
# Helper script to test voice reminders with debug logging
# Usage: test-voice-reminder.sh

set -e

# 新路徑優先
NEW_VOICE_REMINDER="$HOME/.claude/omystatusline/plugins/voice-reminder/bin/voice-reminder"
NEW_DEBUG_LOG="$HOME/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-debug.log"
# 舊路徑作為備援
OLD_VOICE_REMINDER="$HOME/.claude/voice-reminder"
OLD_DEBUG_LOG="$HOME/.claude/voice-reminder-debug.log"

# 如果新的 binary 存在，使用新路徑；否則使用舊路徑
if [ -f "$NEW_VOICE_REMINDER" ]; then
    VOICE_REMINDER="$NEW_VOICE_REMINDER"
    DEBUG_LOG="$NEW_DEBUG_LOG"
else
    VOICE_REMINDER="$OLD_VOICE_REMINDER"
    DEBUG_LOG="$OLD_DEBUG_LOG"
fi

echo "Testing voice reminder system..."
echo ""

# Set debug mode and pipe test JSON to voice-reminder
export VOICE_REMINDER_DEBUG=true
echo '{"message": "測試訊息：Claude 需要您的確認？", "hook_event_name": "Notification", "session_id": "test-123"}' | "$VOICE_REMINDER"

echo ""
echo "📋 Debug log (最後 20 行):"
if [ -f "$DEBUG_LOG" ]; then
    tail -20 "$DEBUG_LOG"
else
    echo "No debug log found at $DEBUG_LOG"
fi
