# Changelog

All notable changes to the Matrix MUD project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Multi-agent development workflow documentation
- Comprehensive documentation suite (CHANGELOG, DEVLOG, CLAUDE, AGENTS)
- Custom slash commands for Claude Code integration
- Advanced architecture documentation

### Changed
- TBD

### Deprecated
- TBD

### Removed
- TBD

### Fixed
- TBD

### Security
- TBD

## [1.0.0] - 2025-11-24

### Added
- Initial project structure with proper Go organization (cmd/, pkg/, internal/, tests/)
- Comprehensive README.md with installation and usage instructions
- CONTRIBUTING.md with development guidelines and workflow documentation
- MIT License
- Professional Makefile with 20+ development commands
- GitHub Actions CI/CD pipeline for automated testing
  - Multi-platform testing (Ubuntu, macOS, Windows)
  - Multi-version Go testing (1.21, 1.22)
  - Code coverage reporting
  - golangci-lint integration
  - Security scanning with gosec
- Release workflow for automated multi-platform binary builds
  - Linux AMD64/ARM64
  - macOS AMD64/ARM64 (Apple Silicon)
  - Windows AMD64
- Dependabot configuration for automated dependency updates
- EditorConfig for consistent coding standards
- Comprehensive .gitignore for Go projects
- golangci-lint configuration with 15+ linters
- Unit and integration test infrastructure
- Docker support with Dockerfile
- Core game features:
  - Multi-user telnet server (port 2323)
  - Web interface for world monitoring (port 8080)
  - Admin console (port 9090)
  - Three character classes: Hacker, Rebel, Operator
  - Real-time combat system
  - Item system with rarity tiers (Common, Uncommon, Rare, Legendary)
  - Procedural city generation
  - Quest system
  - Banking/storage system (The Archive)
  - Player authentication and persistence
  - NPC dialogue system
  - Global chat and private messaging
  - Builder commands for world creation

### Changed
- Updated go.mod with proper module path (github.com/yourusername/matrix-mud)
- Fixed missing color constants (Magenta, Cyan) in terminal.go
- Removed unused imports from world.go
- Formatted all code with gofmt

### Fixed
- Build errors related to undefined color constants
- Unused import warnings

## [0.28.0] - Pre-Release

### Added
- Phase 28 implementation (internal version)
- Basic MUD functionality
- World generation capabilities
- Combat system
- Item management
- Player progression system

---

## Changelog Guidelines

When adding entries, please follow these guidelines:

### Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security fixes and improvements

### Format

```markdown
### Added
- Brief description of the feature [#PR_NUMBER](link-to-pr)
```

### Example Entry

```markdown
## [1.1.0] - 2025-12-01

### Added
- REST API for player management [#15](https://github.com/user/matrix-mud/pull/15)
- PostgreSQL database backend [#16](https://github.com/user/matrix-mud/pull/16)

### Changed
- Improved combat balance [#17](https://github.com/user/matrix-mud/pull/17)

### Fixed
- Race condition in player state updates [#18](https://github.com/user/matrix-mud/pull/18)

### Security
- Implemented bcrypt password hashing [#19](https://github.com/user/matrix-mud/pull/19)
```

## Version Links

[Unreleased]: https://github.com/yourusername/matrix-mud/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/matrix-mud/releases/tag/v1.0.0
[0.28.0]: https://github.com/yourusername/matrix-mud/releases/tag/v0.28.0
