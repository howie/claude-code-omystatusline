#!/bin/bash
# Helper script to test voice reminders with debug logging
# Usage: test-voice-reminder.sh

set -e

VOICE_REMINDER="$HOME/.claude/voice-reminder"
DEBUG_LOG="$HOME/.claude/voice-reminder-debug.log"

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
    echo "No debug log found"
fi
