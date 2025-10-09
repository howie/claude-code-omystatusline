#!/bin/zsh

# Wrapper script to properly handle ANSI escape codes for Claude Code statusline
# Calls the Go binary with JSON input and ensures ANSI codes are interpreted correctly

# Execute the Go statusline binary with JSON input and use printf to interpret ANSI codes
printf "%b" "$(cat | "$HOME/.claude/statusline-go")"