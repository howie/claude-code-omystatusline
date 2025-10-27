# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Claude Code Custom Status Line** is a rich, context-aware status line for Claude Code that displays:
- Current git branch and worktree detection
- Token consumption tracking with visual progress bar
- Session time tracking and multi-session awareness
- Last user message for context recall
- Optional audio notifications via the voice-reminder plugin

The project uses Go 1.21+ with a modular architecture separating concerns into reusable packages in `pkg/`.

## Quick Commands

### Build & Installation
```bash
make build                # Compile Go binaries
make install              # Interactive install (recommended, includes audio setup)
make install-simple       # Simple install (status line only, no audio)
make clean                # Remove compiled output
```

### Development
```bash
make test                 # Run tests
make fmt                  # Format code with gofmt
make lint                 # Run gofmt + golangci-lint checks
make install-hooks        # Install Git pre-commit/pre-push hooks
make uninstall            # Completely remove all installations
```

### Single Test
```bash
go test -v -run TestNamePattern ./path/to/package
# Example: go test -v -run TestGetBranch ./pkg/git
```

## Project Architecture

The codebase follows a clear separation of concerns:

### Key Directories

- **`cmd/`** - Application entry points
  - `cmd/statusline/` - Main status line binary (reads JSON from Claude Code via stdin)
  - `cmd/voice-reminder/` - Voice reminder plugin binary

- **`pkg/`** - Reusable packages (the core logic)
  - `pkg/git/` - Git branch detection, worktree checking, 5-second caching
  - `pkg/context/` - Token usage analysis and progress bar generation
  - `pkg/session/` - Time tracking, multi-session detection, daily accumulation
  - `pkg/statusline/` - Core formatting, color definitions, message extraction
  - `pkg/voicereminder/` - Audio notification handling (optional plugin)

- **`scripts/`** - Installation and wrapper scripts
  - `install.sh` - Interactive installer
  - `statusline-wrapper.sh` - Wrapper script called by Claude Code
  - `statusline.sh` - Bash fallback implementation

- **`docs/`** - Documentation with feature guides and architecture details

### Data Flow

```
Claude Code
    ↓
JSON Input (stdin) containing session metadata
    ↓
cmd/statusline/main.go parses and coordinates:
    ├─ pkg/git.GetBranch() [cached 5 sec]
    ├─ pkg/session.CalculateTotalHours()
    ├─ pkg/context.Analyze()
    └─ pkg/statusline.ExtractUserMessage()
    ↓
Results collected via channels
    ↓
Formatted status line output to stdout
```

### Design Principles

1. **Concurrent processing**: Git, context, and session data fetched in parallel goroutines
2. **Smart caching**: Git branch cached to avoid performance overhead
3. **Graceful degradation**: Missing data (failed git calls, missing files) doesn't break output
4. **Efficient I/O**:
   - Only last 100-200 lines of transcript read
   - Uses structured JSON parsing
5. **Modular testing**: Each package independently testable with unit tests

## Important Implementation Details

### Git Module (`pkg/git/`)
- Uses 5-second cache with mutex to reduce `git` command calls
- Detects worktrees via `.worktrees/` directory structure
- Gracefully handles non-git directories by returning empty string

### Context Module (`pkg/context/`)
- Analyzes transcript to estimate token count
- Generates visual progress bar (██████░░░░ format)
- Color-coded warnings: green (<60%), gold (60-80%), red (≥80%)
- Based on 200k token limit (Claude's session context window)

### Session Module (`pkg/session/`)
- Stores session data in `~/.claude/session-tracker/`
- Tracks daily accumulated time across multiple sessions
- Intelligently groups sessions: gaps >10 minutes create new intervals
- Returns formatted time like "2h45m" and session count like "[3 sessions]"

### Statusline Module (`pkg/statusline/`)
- `builder.go`: Core formatting logic, message extraction, system message filtering
- `color.go`: ANSI color constants (gold for Opus, cyan for Sonnet, pink for Haiku)
- `model.go`: Data structures (Input, Result, ContextInfo)

### Voice Reminder Plugin
- Separate Go binary in `cmd/voice-reminder/`
- Runs as a Claude Code hook to play audio notifications
- Can be enabled/disabled via slash commands
- Supports system sounds, custom audio files, and text-to-speech

## Common Development Tasks

### Adding a New Information Source
1. Create a new package in `pkg/` (e.g., `pkg/diagnostics/`)
2. Implement the data collection function
3. Add a goroutine in `cmd/statusline/main.go` that calls it
4. Update the result switch statement to handle the new type
5. Modify the output formatting in the printf statement

### Testing a Specific Feature
```bash
# Test git branch detection
go test -v ./pkg/git

# Test context tracking
go test -v ./pkg/context

# Run all tests with verbose output
go test -v -count=1 ./...
```

### Code Quality
- All code must pass `gofmt -s` (checked by `make lint`)
- Use `golangci-lint` for linting (configured in CI)
- Comment exported functions and non-obvious logic
- Write unit tests for new functionality

## Installation File Structure

After `make install`, files are organized under `~/.claude/omystatusline/`:
```
~/.claude/omystatusline/
├── bin/
│   ├── statusline-go              # Main status line binary
│   └── statusline-wrapper.sh      # Wrapper script
├── scripts/
│   └── statusline.sh             # Bash fallback
└── plugins/
    └── voice-reminder/           # Optional plugin (if installed)
```

The project uses symlinks in `~/.claude/commands/` to integrate voice-reminder slash commands with Claude Code's native command system.

## Requirements

- **Go 1.21+** (for compilation)
- **Git** (for branch detection)
- **golangci-lint** (for linting, installed via `brew install golangci-lint` on macOS)
- Terminal with ANSI color support

## Key Files to Know

- `README.md` - Complete user-facing documentation with bilingual content
- `Makefile` - All build, test, and installation automation
- `cmd/statusline/main.go` - Entry point and orchestration logic
- `docs/guides/architecture.md` - Detailed architecture explanation
- `CHANGELOG.md` - Version history and feature additions

## Performance Considerations

- Status line updates should complete in <100ms
- Memory usage typically <10MB
- Git operations cached for 5 seconds to avoid repeated calls
- Only reads necessary portions of transcript files (last 100-200 lines)

## Error Handling Patterns

All external operations (git commands, file reads) use graceful fallback:
- Missing git repo → return empty string (status line still works)
- Missing transcript file → return empty message
- Malformed JSON → exit with error (upstream issue)
- File permission errors → continue with reduced functionality

This ensures the status line never breaks Claude Code's operation even if individual components fail.
