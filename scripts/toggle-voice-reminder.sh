#!/bin/bash
# Helper script to enable/disable voice reminders
# Usage: toggle-voice-reminder.sh [on|off]

set -e

# 新路徑優先
NEW_ENABLED_FILE="$HOME/.claude/omystatusline/plugins/voice-reminder/data/voice-reminder-enabled"
# 舊路徑作為備援
OLD_ENABLED_FILE="$HOME/.claude/voice-reminder-enabled"

# 如果新目錄存在，使用新路徑；否則使用舊路徑
if [ -d "$HOME/.claude/omystatusline/plugins/voice-reminder/data" ]; then
    ENABLED_FILE="$NEW_ENABLED_FILE"
else
    ENABLED_FILE="$OLD_ENABLED_FILE"
fi

case "$1" in
  on)
    echo "true" > "$ENABLED_FILE"
    echo "✅ 語音提醒已啟用 / Voice reminders enabled"
    ;;
  off)
    echo "false" > "$ENABLED_FILE"
    echo "🔇 語音提醒已關閉 / Voice reminders disabled (muted)"
    ;;
  *)
    echo "Usage: $0 [on|off]"
    exit 1
    ;;
esac
