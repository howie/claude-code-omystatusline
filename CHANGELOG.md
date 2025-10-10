# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/yourusername/claude-code-omystatusline/releases/tag/v0.1.0
