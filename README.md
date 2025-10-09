# Claude Code Custom Status Line

A rich, informative status line for [Claude Code](https://claude.com/claude-code) that displays real-time session information, git status, context usage, and more.

## Features

- **Model Display**: Shows the current Claude model (Opus 💛, Sonnet 💠, or Haiku 🌸) with color-coded icons
- **Project Info**: Displays current project name from the working directory
- **Git Integration**:
  - Shows current branch with ⚡ icon
  - Detects and displays git worktree status with 🔀 icon
  - Smart caching for improved performance
- **Context Usage Tracking**:
  - Visual progress bar showing token usage
  - Color-coded percentage (green < 60%, gold < 80%, red >= 80%)
  - Formatted token count (e.g., 45k, 120k)
  - Based on 200k token context window
- **Session Time Tracking**:
  - Tracks total active time across all sessions today
  - Displays multiple active sessions when detected
  - Automatic session archiving for past days
- **User Message Display**: Shows the last user message for context

## Installation

### Using Go Implementation (Recommended)

1. **Build the Go binary:**
   ```bash
   go build -o ~/.claude/statusline-go statusline.go
   ```

2. **Copy the wrapper script:**
   ```bash
   cp statusline-wrapper.sh ~/.claude/statusline
   chmod +x ~/.claude/statusline
   ```

3. **Configure Claude Code:**

   Add to your `~/.claude/config.json`:
   ```json
   {
     "statusLineCommand": "~/.claude/statusline"
   }
   ```

### Using Bash Implementation

1. **Copy the bash script:**
   ```bash
   cp statusline.sh ~/.claude/statusline
   chmod +x ~/.claude/statusline
   ```

2. **Configure Claude Code** (same as above)

## How It Works

The status line receives JSON input from Claude Code containing:
- Current model information
- Session ID
- Workspace directory
- Transcript path (for context analysis)

It outputs a formatted status line with ANSI color codes showing:
```
[💠 Sonnet 4.5] 📂 my-project ⚡ main | ██████░░░░ 65% 130k | 2h45m [2 sessions]
｜Your last message appears here...
```

### Status Line Components

1. **Model Badge**: `[💠 Sonnet 4.5]` - Shows current model with icon
2. **Project Name**: `📂 my-project` - Current working directory name
3. **Git Branch**: `⚡ main` - Current branch (🔀 if in worktree)
4. **Context Usage**: `██████░░░░ 65% 130k` - Visual progress bar with percentage and token count
5. **Session Time**: `2h45m` - Total active time today
6. **Multi-Session**: `[2 sessions]` - Shown when multiple sessions are active
7. **User Message**: Shows your last message for quick reference

### Session Tracking

Sessions are tracked in `~/.claude/session-tracker/sessions/` with automatic:
- Time interval recording (gaps > 10 minutes create new intervals)
- Daily totals calculation
- Old session archiving to `~/.claude/session-tracker/archive/`

## Performance Optimizations

The Go implementation includes several optimizations:
- **Concurrent Processing**: Git, context, and session data fetched in parallel using goroutines
- **Smart Caching**: Git branch cached for 5 seconds to reduce git command overhead
- **Efficient Parsing**: JSON parsing with structured types for fast data extraction
- **Minimal I/O**: Reads only last 100-200 lines of transcript for context analysis

## Requirements

### Go Implementation
- Go 1.16 or later
- Git (for branch detection)
- Claude Code CLI

### Bash Implementation
- Bash 4.0+
- jq (JSON processor)
- Git (for branch detection)
- Claude Code CLI

## Development

### Project Structure

```
.
├── statusline.go           # Go implementation (recommended)
├── statusline.sh           # Bash implementation (alternative)
├── statusline-wrapper.sh   # Wrapper for Go binary
└── README.md
```

### Building from Source

```bash
# Build the Go binary
go build -o statusline-go statusline.go

# Run tests (if available)
go test ./...
```

## Troubleshooting

### Status line not appearing
- Check that `~/.claude/statusline` is executable: `chmod +x ~/.claude/statusline`
- Verify the config.json path is correct
- Test the script manually: `echo '{"model":{"display_name":"Sonnet 4.5"},"session_id":"test","workspace":{"current_dir":"'$(pwd)'"}}' | ~/.claude/statusline`

### Colors not displaying correctly
- Ensure your terminal supports ANSI color codes
- Try the wrapper script for proper escape code handling

### Session tracking issues
- Check permissions on `~/.claude/session-tracker/`
- Verify session JSON files are valid: `cat ~/.claude/session-tracker/sessions/*.json | jq`

## License

MIT License - feel free to customize and extend!

## Contributing

Contributions welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## Credits

Created for the Claude Code community to enhance the CLI experience with rich status information.
