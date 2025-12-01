# Changelog

All notable changes to the Matrix MUD project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.32.0] - 2025-12-01 - Phase 4 Deployment Ready

### Added
- **fly.toml**: Complete Fly.io deployment configuration
  - TCP service for telnet (port 2323)
  - HTTP service for web client (ports 80/443 with TLS)
  - Persistent volume for game data
  - Health checks for both services
- **scripts/deploy.sh**: Automated deployment script
  - Supports local, Docker, and Fly.io deployments
  - Pre-flight checks (tests, build verification)
  - Color-coded output with clear instructions
- **.env.production.example**: Production configuration guide
  - Security checklist
  - Docker and Fly.io deployment examples
  - Secret management instructions
- **phase4_test.go**: Deployment configuration tests
- **GET /health**: Health check endpoint for load balancers
  - Returns JSON: `{"status":"healthy","version":"1.31.0","service":"matrix-mud"}`

### Changed
- **Dockerfile**: Complete rewrite with security best practices
  - Multi-stage build for minimal image size
  - Non-root user (mud:mud) for security
  - Built-in health check
  - Optimized layer caching
- **web.go**: Added /health endpoint handler
- Version bumped to v1.32

### Security
- Docker container runs as non-root user
- Health endpoint enables proper load balancer integration
- Production documentation includes security checklist

### Deployment Options
1. **Local**: `./scripts/deploy.sh local`
2. **Docker**: `./scripts/deploy.sh docker`
3. **Fly.io**: `./scripts/deploy.sh fly`

### Tests
All 25 tests passing (3 new deployment tests)

---

## [1.31.0] - 2025-12-01 - Phase 3 Enhancements

### Added
- **phase3_test.go**: Comprehensive test suite for Phase 3 enhancements (8 tests)
  - TestIACEchoConstants: Validates telnet IAC codes
  - TestConnectionTimeoutValues: Validates sensible timeout ranges
  - TestDownCommandAlias: Documents command design
  - TestBroadcastNilSafety: Validates nil handling in broadcast
  - TestXtermJSVersion: Validates xterm.js 5.x usage
  - TestWorldJSONIntegrity: Validates world data structure
  - TestGameVersion: Documents version tracking

### Changed
- **main.go**: Complete Phase 3 implementation
  - P3-ENH-19: IAC echo suppression for secure password input
  - P3-ENH-20: Connection timeouts now enforced (30s initial, 30min idle)
  - P3-ENH-21: 'dn' alias added for 'down' command
  - P3-ENH-22: Broadcast function handles nil sender safely
- **web.go**: xterm.js updated from 3.14.5 to 5.3.0
  - Modern terminal emulation with fit addon
  - Better mobile support
- Version bumped to v1.31

### Security
- Password input now suppresses echo via telnet IAC commands
- Idle connections automatically disconnected after 30 minutes
- Login timeout prevents hanging connections (30 seconds)

### Tests
All 22 tests passing:
```
=== PHASE 1 (Security) ===
TestConfigEnvironmentVariables     ✅ PASS
TestConfigDefaultPorts             ✅ PASS
TestAllowedOriginsConfig           ✅ PASS
TestGetEnvFunction                 ✅ PASS
TestAdminBindAddressNotExposed     ✅ PASS

=== PHASE 2 (Bug Fixes) ===
TestPhase1_Inventory               ✅ PASS
TestNilRoomAccessNoPanic           ✅ PASS
TestInventorySizeLimit             ✅ PASS
TestDownAlias                      ✅ PASS
TestNPCHPValues                    ✅ PASS
TestJSONLoadErrorHandling          ✅ PASS
TestConnectionTimeoutConfig        ✅ PASS
TestWorldInitialization            ✅ PASS
TestPhase2_NPCs                    ✅ PASS

=== PHASE 3 (Enhancements) ===
TestIACEchoConstants               ✅ PASS
TestConnectionTimeoutValues        ✅ PASS
TestDownCommandAlias               ✅ PASS
TestBroadcastNilSafety             ✅ PASS
TestXtermJSVersion                 ✅ PASS
TestWorldJSONIntegrity             ✅ PASS
TestGameVersion                    ✅ PASS

TOTAL: 22/22 PASSING
```

---

## [1.30.0] - 2025-11-28 - Phase 2 Bug Fixes

### Added
- **phase2_fixes_test.go**: Comprehensive test suite for Phase 2 bug fixes
  - TestNilRoomAccessNoPanic: Validates nil room handling
  - TestInventorySizeLimit: Validates inventory cap at 20 items
  - TestDownAlias: Documents command alias design
  - TestNPCHPValues: Validates all NPCs have valid HP/MaxHP
  - TestJSONLoadErrorHandling: Validates graceful JSON error recovery
  - TestConnectionTimeoutConfig: Validates timeout constants
  - TestWorldInitialization: Validates world setup
- **TASKS.md**: Comprehensive task tracking for multi-phase development
- Game balance constants in config.go:
  - MaxInventorySize = 20 items
  - ConnectionTimeout = 30 seconds
  - IdleTimeout = 30 minutes
  - DefaultNPCHP = 50
  - DefaultNPCMaxHP = 50

### Fixed
- **Issue #4**: Nil room access no longer causes panic
  - world.go Look() now returns error message for invalid rooms
  - Players in void can use 'recall' to return to safety
- **Issue #7**: JSON unmarshal errors now handled gracefully
  - loadWorldData() logs warnings and creates default world on error
  - createDefaultWorld() provides minimal fallback world
- **Issue #8**: Inventory size now limited to 20 items
  - GetItem() checks inventory size before adding
  - Returns "inventory full" message when at capacity
- **Issue #13-14**: NPC HP values validated during world load
  - NPCs with HP <= 0 get DefaultNPCHP (50)
  - NPCs with MaxHP <= 0 or MaxHP < HP get corrected
  - Warnings logged for each correction

### Changed
- **world.go**: Added 'log' import for error logging
- **config.go**: Added time import and game balance constants
- Version bumped to v1.30

### Tests
All 15 tests passing:
```
=== RUN   TestPhase1_Inventory               --- PASS
=== RUN   TestNilRoomAccessNoPanic           --- PASS
=== RUN   TestInventorySizeLimit             --- PASS
=== RUN   TestDownAlias                      --- PASS
=== RUN   TestNPCHPValues                    --- PASS
=== RUN   TestJSONLoadErrorHandling          --- PASS
=== RUN   TestConnectionTimeoutConfig        --- PASS
=== RUN   TestWorldInitialization            --- PASS
=== RUN   TestPhase2_NPCs                    --- PASS
=== RUN   TestConfigEnvironmentVariables     --- PASS
=== RUN   TestConfigDefaultPorts             --- PASS
=== RUN   TestAllowedOriginsConfig           --- PASS
=== RUN   TestGetEnvFunction                 --- PASS
=== RUN   TestAdminBindAddressNotExposed     --- PASS
PASS ok github.com/yourusername/matrix-mud 0.201s
```

---

## [1.29.0] - 2025-11-28 - ULTRATHINK Security Overhaul (Phase 1)

### Added
- **config.go**: New centralized configuration management system
  - All sensitive values loaded from environment variables
  - Auto-generates secure 32-character random passwords when ADMIN_PASS not set
  - Clear warning messages for production deployment
- **security_test.go**: Comprehensive security test suite
  - TestConfigEnvironmentVariables: Validates env-based configuration
  - TestConfigDefaultPorts: Validates default port settings
  - TestAllowedOriginsConfig: Validates WebSocket origin settings
  - TestGetEnvFunction: Validates helper function
  - TestAdminBindAddressNotExposed: Validates admin security
- **.env.example**: Documentation of all configurable environment variables
- Rate limiter cleanup goroutine (hourly) to prevent memory leaks

### Changed
- **admin.go**: Complete security overhaul
  - Admin credentials now loaded from Config (environment variables)
  - Default bind address changed from `0.0.0.0:9090` to `127.0.0.1:9090` (localhost only)
  - Added `checkAdminAuth()` helper function for DRY authentication
  - Warning displayed in admin panel when using auto-generated password
  - Better error handling and HTTP status codes
  - Logging for admin actions (kicks)
- **web.go**: WebSocket security improvements
  - Added `checkWebSocketOrigin()` function with configurable origin whitelist
  - ALLOWED_ORIGINS environment variable support ("*" for dev, comma-separated domains for prod)
  - Improved error logging for WebSocket and telnet connection failures
  - Added Content-Type header for HTML responses
- **main.go**: Server configuration improvements
  - Telnet port now configurable via TELNET_PORT environment variable
  - Added init() function to start rate limiter cleanup goroutine
  - Improved startup logging with all service URLs
  - Better error handling for listener failures
  - Version bumped to v1.29
- **go.mod**: Fixed Go version from non-existent 1.24.0 to 1.21

### Security
- **CRITICAL**: Removed hardcoded admin credentials (was admin/admin)
- **CRITICAL**: Admin panel no longer exposed to internet by default
- **HIGH**: WebSocket origin checking now configurable (was allowing all origins)
- **MEDIUM**: Rate limiter memory leak fixed with hourly cleanup
- All 7 security tests passing

### Tests
```
=== RUN   TestConfigEnvironmentVariables     --- PASS
=== RUN   TestConfigDefaultPorts             --- PASS  
=== RUN   TestAllowedOriginsConfig           --- PASS
=== RUN   TestGetEnvFunction                 --- PASS
=== RUN   TestAdminBindAddressNotExposed     --- PASS
=== RUN   TestPhase1_Inventory               --- PASS
=== RUN   TestPhase2_NPCs                    --- PASS
PASS ok github.com/yourusername/matrix-mud 0.563s
```

---

## [Unreleased-Post-1.29]

### Added
- Multi-agent development workflow documentation
- Comprehensive documentation suite (CHANGELOG, DEVLOG, CLAUDE, AGENTS)
- Custom slash commands for Claude Code integration
- Advanced architecture documentation
- Comprehensive godoc comments for world.go (25+ functions documented)
- Comprehensive godoc comments for web.go and admin.go
- Input validation package (pkg/validation/) with username, command, and room ID validation
- Input sanitization to prevent terminal escape sequence injection
- Rate limiting package (pkg/ratelimit/) with token bucket algorithm
- Development scripts for common workflows:
  - dev.sh: Development environment startup script
  - load-test.sh: Concurrent player load testing script
  - backup-data.sh: Automated data backup with timestamps
- Git hooks for code quality:
  - pre-commit: Runs formatting, linting, and tests before commits
  - post-commit: Suggests devlog entries after code changes
- Air hot reload configuration (.air.toml) for development
- Nokia Phone item added to loading_program room for testing

### Changed
- Updated Makefile dev target to use correct Air module path (github.com/air-verse/air)
- Updated go.mod to Go 1.24.0
- Added golang.org/x/crypto dependency for bcrypt password hashing
- Increased minimum password length requirement from 3 to 8 characters
- File permissions for sensitive data files changed from 0644 to 0600 (owner read/write only)

### Deprecated
- Plaintext password storage (replaced with bcrypt hashing)

### Removed
- TBD

### Fixed
- Test compilation errors in phase1_test.go and phase2_test.go
- Tests now correctly use ItemMap and NPCMap instead of Items and NPCs slices
- All tests now passing (TestPhase1_Inventory, TestPhase2_NPCs)

### Security
- **CRITICAL**: Implemented bcrypt password hashing to replace plaintext password storage
  - All new passwords are hashed with bcrypt.DefaultCost (cost factor 10)
  - Existing plaintext passwords will require users to create new accounts
- Implemented rate limiting for authentication (5 attempts per minute per user)
- Added 3-second delay for rate-limited clients to slow down brute force attacks
- Changed file permissions to 0600 for all sensitive data files:
  - data/users.json (user credentials)
  - data/players/*.json (player data)
  - data/world.json (world state)
- Added comprehensive input validation:
  - Username validation: 3-20 alphanumeric characters and underscores
  - Command validation: lowercase letters and spaces (1-100 chars)
  - Room ID validation: alphanumeric with underscores and hyphens (1-50 chars)
- Added input sanitization to remove control characters and prevent injection attacks
- Comprehensive error handling and logging for security events

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
