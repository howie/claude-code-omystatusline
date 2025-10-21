#!/bin/bash
# Helper script to enable/disable voice reminders
# Usage: toggle-voice-reminder.sh [on|off]

set -e

ENABLED_FILE="$HOME/.claude/voice-reminder-enabled"

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
