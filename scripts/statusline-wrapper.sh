#!/bin/bash

# Wrapper script to properly handle ANSI escape codes for Claude Code statusline
# Calls the Go binary with JSON input and ensures ANSI codes are interpreted correctly

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Execute the Go statusline binary with JSON input and use printf to interpret ANSI codes
# Set STATUSLINE_DUMP_STDIN=/tmp/statusline-stdin.json to capture the raw stdin for debugging.
# Reads stdin once into a variable so the dump and the binary receive the same data.
if [ -n "$STATUSLINE_DUMP_STDIN" ]; then
  stdin_data=$(cat)
  if ! printf '%s\n' "$stdin_data" > "$STATUSLINE_DUMP_STDIN"; then
    printf 'statusline-wrapper: WARNING: could not write STATUSLINE_DUMP_STDIN to %s\n' "$STATUSLINE_DUMP_STDIN" >&2
  fi
  printf "%b" "$(printf '%s\n' "$stdin_data" | "$SCRIPT_DIR/statusline-go")"
else
  printf "%b" "$(cat | "$SCRIPT_DIR/statusline-go")"
fi
