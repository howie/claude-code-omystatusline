#!/bin/bash

# Wrapper script to properly handle ANSI escape codes for Claude Code statusline
# Calls the Go binary with JSON input and ensures ANSI codes are interpreted correctly

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Execute the Go statusline binary with JSON input and use printf to interpret ANSI codes
# Set STATUSLINE_DUMP_STDIN=/tmp/statusline-stdin.json to capture the raw stdin for debugging
if [ -n "$STATUSLINE_DUMP_STDIN" ]; then
  printf "%b" "$(tee "$STATUSLINE_DUMP_STDIN" | "$SCRIPT_DIR/statusline-go")"
else
  printf "%b" "$(cat | "$SCRIPT_DIR/statusline-go")"
fi