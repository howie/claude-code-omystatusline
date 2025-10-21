#!/bin/bash

# Wrapper script to properly handle ANSI escape codes for Claude Code statusline
# Calls the Go binary with JSON input and ensures ANSI codes are interpreted correctly

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Execute the Go statusline binary with JSON input and use printf to interpret ANSI codes
printf "%b" "$(cat | "$SCRIPT_DIR/statusline-go")"