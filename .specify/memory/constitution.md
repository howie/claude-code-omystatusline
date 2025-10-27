# Claude Code Custom Status Line Constitution

## Core Principles

### I. Modular Architecture (Separation of Concerns)
Each package in `pkg/` has a single, well-defined responsibility. Packages must be:
- **Self-contained**: No circular dependencies between packages
- **Independently testable**: Each package includes unit tests in `*_test.go` files
- **Clearly scoped**: One concern per package (git, context, session, statusline, voicereminder)
- **Documented**: Public functions include godoc comments

### II. Concurrent Processing via Goroutines
Performance-critical operations (git, context analysis, session tracking) run in parallel:
- Main orchestrator in `cmd/statusline/main.go` spawns 4 concurrent goroutines
- Results collected via buffered channel to prevent blocking
- WaitGroup ensures all goroutines complete before aggregating output
- Cache mechanisms prevent redundant I/O (e.g., 5-second git cache)

### III. Graceful Degradation (Resilience)
External operations must never break the status line:
- Missing git repo → return empty string (don't crash)
- Malformed transcript → return empty message
- Failed file I/O → continue with partial data
- All errors logged but don't prevent output to stdout

### IV. Efficient I/O Operations
Minimize filesystem and git command overhead:
- Git branch cached for 5 seconds with mutex protection
- Transcript analysis reads only last 100-200 lines (not entire file)
- Structured JSON parsing for input/output
- Buffered readers for file scanning

### V. Go-First Implementation with Bash Fallback
Primary implementation in Go 1.21+:
- Compiled binaries in `cmd/` for performance
- Bash scripts in `scripts/` as documented fallback
- Both must work identically (feature parity)
- Voice-reminder plugin uses same Go pattern as main statusline

### VI. Plugin Architecture
Voice-reminder extends functionality without bloating core:
- Separate binary in `cmd/voice-reminder/`
- Independent configuration in `.specify/voice-reminder-config.json`
- Integrated via Claude Code hooks (pass-through when disabled)
- Slash commands managed via symlinks in `~/.claude/commands/`

### VII. Quality Standards (Non-Negotiable)
All code must pass before merge:
- **Format**: `gofmt -s` (checked by `make lint`)
- **Lint**: `golangci-lint` with 5-minute timeout
- **Tests**: `go test -v -count=1 ./...` with 100% pass rate
- **Comments**: Exported functions have godoc; complex logic documented

## Performance Standards

- **Status line latency**: <100ms from input to output
- **Memory usage**: <10MB per invocation
- **Git operations**: 5-second cache TTL maximum
- **File I/O**: Never read entire transcript (max 200 lines)

## Testing Requirements

- **Unit tests**: Required for all new packages in `pkg/`
- **Test coverage**: Aim for 80%+ on critical paths (git, context, session)
- **Integration**: Test git caching behavior, transcript parsing edge cases
- **Running tests**: `make test` or `go test -v ./...`

## Development Workflow

1. **Planning**: Use TodoWrite to break complex tasks into steps
2. **Implementation**: Follow modular patterns; add tests first where possible
3. **Quality Check**: Run `make lint && make test` before committing
4. **Documentation**: Update CLAUDE.md if architecture changes; add godoc comments
5. **Git Hooks**: Pre-commit hook verifies lint, pre-push hook runs tests

## Governance

**Constitution supersedes all other practices.** All code changes must:
- Maintain the 7 core principles above
- Pass all quality gates (lint, format, tests)
- Include documentation updates if behavior changes
- Reference CLAUDE.md for architectural consistency

Complexity additions require:
1. Justification: What problem does it solve?
2. Testing: Unit and integration tests
3. Documentation: Update CLAUDE.md and/or godoc comments
4. Performance impact assessment

**Runtime Development Guidance**: Refer to CLAUDE.md for:
- Quick build/test commands
- Architecture overview with data flow
- Common development tasks and patterns
- File organization rationale

**Version**: 1.0.0 | **Ratified**: 2025-10-27 | **Last Amended**: 2025-10-27
