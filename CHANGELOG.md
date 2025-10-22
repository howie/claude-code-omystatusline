# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-10-22

### Added
- **Plugin architecture** with organized directory structure for better modularity
- **PreToolUse hook support** for real-time notifications when Claude Code uses tools
  - Notifications for AskUserQuestion events
  - Configurable filtering to control which tool uses trigger notifications
  - Integration with voice-reminder plugin
- **Claude Code v2.0.25+ statusline format support** for compatibility with latest releases
- **Voice-reminder plugin documentation** with comprehensive usage guide
- CLI flag support for installer with `--help` and `--version` options
- gofmt check integrated into `make lint` for consistent code formatting
- Chinese voice support for TTS notifications
- TTS voice test during installation for user verification
- Interactive installer with audio notification options
- Pre-push Git hook for quality checks

### Fixed
- Installer updated to use settings.json and properly configure PreToolUse hook
- Audio notification hook to use correct Claude Code event name
- TTS installation error when selecting option 3
- Compilation check in installer
- CI compatibility issues for cross-platform support

### Changed
- Migrated to new plugin architecture with organized directory structure
- Output binaries to `output/` directory instead of project root for cleaner workspace
- Enhanced installer with language selection (English/Chinese)
- Improved statusline formatting with additional tests

### Removed
- Redundant MIGRATION.md documentation file
- Redundant QUICK_START.md documentation file

## [1.1.0] - 2025-10-21

### Added
- **Timeout mechanism** for voice playback (10-second timeout) to prevent `say` command from hanging indefinitely
- **Automatic retry logic** with single retry on first failure for improved reliability
- **Fallback to system sounds** when voice playback times out or fails after retries
- **CLI flags** for voice-reminder binary:
  - `--help` / `-h`: Display usage information and features
  - `--version` / `-v`: Show version information
- Enhanced debug logging for timeout events and retry attempts
- Improved error handling with detailed logging for troubleshooting

### Fixed
- Voice reminder hanging issue when macOS `say` command gets stuck (#7)
  - Previously 50% success rate (2/4 attempts failed)
  - Now protected by timeout mechanism with automatic recovery
- Zombie processes left behind when `say` command hangs
- Silent failures when audio device is locked or TTS engine encounters errors

### Changed
- Voice playback now uses goroutines with timeout channels for better reliability
- Updated speaker.go to include `SpeakWithLogger` function for better integration with debug logging
- Improved voice-reminder architecture with separate functions for macOS and Linux platforms

### Technical Details
- Timeout protection: 10 seconds maximum wait time for voice playback
- Retry delay: 500ms between retry attempts
- Maximum retries: 1 attempt before falling back to system sound
- Process cleanup: Automatic kill of hung processes after timeout
- Logger integration: Timeout events and retries logged to debug log

## [0.1.1] - 2025-10-11

### Added
- Professional badges to README (License, Stars, Forks, Go Report Card, Release, Go Version, CI, Last Commit)
- GitHub Actions CI workflow for automated testing and linting
- GitHub Actions Release workflow for automated binary releases
- Go module support with go.mod file
- Go Report Card integration with A+ rating
- Cross-platform binary builds in CI/CD (Linux, macOS, Windows)

### Fixed
- License discrepancy in README (changed from MIT to Apache 2.0 to match LICENSE file)
- Linting errors in statusline.go (unchecked error returns, ineffectual assignments)
- Shell wrapper compatibility (changed shebang from zsh to bash for Ubuntu support)
- CI test for shell wrapper to properly set up binary location

### Changed
- Improved code formatting with go fmt
- Updated CI workflow to test on multiple Go versions (1.21, 1.22, 1.23)
- Enhanced error handling in JSON unmarshaling and file operations

## [0.1.0] - 2025-10-10

### Added
- Initial release of Claude Code Custom Status Line
- Model display with emoji badges (Opus 💛, Sonnet 💠, Haiku 🌸)
- Project information showing current directory name
- Git integration with branch detection and worktree support
- Smart git caching (5-second cache to optimize performance)
- Token consumption tracking with visual progress bar
- Color-coded usage warnings (Green < 60%, Gold 60-80%, Red ≥ 80%)
- Session time tracking with multi-session awareness
- User message context display in status line
- Parallel processing with concurrent goroutines
- Makefile for automated build and installation
- Comprehensive bilingual README (English/Chinese)
- FAQ section explaining Claude usage status and reset time checking
- Limitation warning feature documentation
- Installation guide and quick install script

### Technical Details
- Go-based implementation with performance optimizations
- Sub-100ms status update performance
- Efficient transcript parsing (last 100-200 lines only)
- ANSI color support for terminal display
- JSON-based configuration integration with Claude Code

### Documentation
- English and Chinese documentation
- Installation guides (quick and manual)
- Feature documentation for limitation warning system
- Testing documentation for developers
- Screenshot examples

[1.2.0]: https://github.com/howie/claude-code-omystatusline/releases/tag/v1.2.0
[1.1.0]: https://github.com/howie/claude-code-omystatusline/releases/tag/v1.1.0
[0.1.1]: https://github.com/howie/claude-code-omystatusline/releases/tag/v0.1.1
[0.1.0]: https://github.com/howie/claude-code-omystatusline/releases/tag/v0.1.0
