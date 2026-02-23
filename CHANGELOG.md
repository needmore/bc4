# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `comment download-attachments` command for downloading attachments from comments
- `document download-attachments` command for downloading attachments from documents
- `--include-comments` flag on card/todo/message/document download-attachments to also download comment attachments
- Attachment display in `comment view` output

### Changed
- Refactored download logic into shared `internal/download` package, reducing code duplication

## [0.13.0] - 2026-01-19

### Added
- Schedules & calendar support for managing project schedules and events (#124)
- Project administration commands for managing project settings (#125)
- People management commands for managing project members (#126)
- Search functionality for finding resources across Basecamp (#127)
- CLAUDE.md for AI assistant guidance

### Fixed
- Merge conflict in root.go imports

## [0.12.0] - 2026-01-19

### Added
- `--group` parameter to `todo add` command for creating todos in specific groups (#123)

## [0.11.3] - 2026-01-18

### Added
- `--color` flag support for todo groups and card columns (#121)

### Fixed
- Unwanted HTML wrapping for todo titles with dashes

## [0.11.2] - 2026-01-16

### Fixed
- Card move command now ensures cards stay on the same board (#119)

## [0.11.1] - 2026-01-15

### Fixed
- CI workflow updates for golangci-lint v2 compatibility

## [0.11.0] - 2026-01-15

### Added
- Activity command for fetching project events (#108)
- `--attach` flag to todo and card commands for file attachments (#118)
- Exponential backoff with automatic retry for transient API errors (#116)
- `--project` and `--account` flags to all comment commands
- API implementation status tracking document (#114)

### Changed
- Remote auth and auth failure handling improvements (#117)

### Fixed
- Added `--account` and `--project` flags to check, uncheck, and comment create commands (#110)

## [0.10.0] - 2026-01-06

### Added
- Column on-hold status commands for card management (#87, #107)
- Message pin and unpin commands (#104)
- Profile command to view current user info (#102)
- Attachment upload support for comments (#100)
- Quick Start section to root help output (#103)

### Changed
- Pinned messages now show pinned status in list and sort first (#105)

### Fixed
- `--with-comments` output now renders in terminal, raw when piped (#106)
- CI workflow improvements and updates
- Updated README with missing commands and improved examples

## [0.9.0] - 2026-01-05

### Added
- Todo `edit-list` command to update todo lists (#98)
- Todo `move` command to reposition todos within a list (#97)
- Todo `edit` command for modifying existing todos (#96)
- Support for viewing and listing attachments on todos and cards (#80)

## [0.8.0] - 2026-01-01

### Added
- Improved CLI help and feedback to match GitHub CLI patterns (#81)

### Changed
- Updated GoReleaser action to v6 for v2 config support

## [0.7.2] - 2025-10-27

### Added
- AI-optimized markdown output with `--with-comments` flag (#78)
- Support for creating and managing grouped to-do lists (to-do list groups) via Basecamp API (#72)
- `--with-comments` flag to view commands for inline comment display (#73)

### Changed
- Updated GoReleaser config to v2 format

### Fixed
- Removed unsupported `--draft` flag from message post command (#75)
- Panic when truncating strings with multi-byte UTF-8 characters (emojis) (#69)
- Improved README (#67)

## [0.7.1] - 2025-10-01

### Fixed
- Pagination bug and improved UI/UX across multiple features

## [0.7.0] - 2025-09-30

### Fixed
- Handle non-numeric account IDs in pagination URL extraction

## [0.6.0] - 2025-08-19

### Added
- Comprehensive commenting system for Basecamp resources (#65)
- Document support for Basecamp API (#62)
- Linux and Windows builds (#64)

### Changed
- Improved Markdown to Basecamp Rich Text conversion fidelity (#60)

### Fixed
- API pagination to show complete todo lists (#61)

### Documentation
- Updated README with comment management and platform support

## [0.5.0] - 2025-07-28

### Added
- Interactive todo list selection with `bc4 todo select` command (#40)
- Message board functionality for posting and viewing Campfire messages (#36)
- Unified selection UI using list components for better user experience

### Changed
- Project and account selection now use consistent list-based UI with filtering capability
- Improved UI consistency across all interactive selection commands

### Fixed
- Resolved multiple linting issues for better code quality
- Fixed table rendering error in message list command

### Documentation
- Updated README to reflect new message board functionality
- Clarified that prerequisites are only needed for building from source

## [0.4.0] - 2025-07-18

### Fixed
- Homebrew installation instructions to use explicit GitHub URL when tapping
- This resolves the issue where Homebrew was looking for homebrew-bc4 repository instead of the main bc4 repository

## [0.3.0] - 2025-07-18

### Added
- golangci-lint configuration and GitHub Action for code quality enforcement (#34)
- Comprehensive test coverage across multiple packages (#30)
- GitHub Actions workflow for automated linting on pull requests

### Fixed
- Test expectations for IsFirstRun
- Various linting issues across the codebase

### Changed
- Updated Makefile lint target to run golangci-lint on all files
- Improved README with current features and contribution guidelines (#32)

## [0.2.0] - 2025-07-17

### Added
- Improved error handling with actionable advice (#28)
- Attribution for Needmore Designs in README

### Fixed
- Resolved all golangci-lint errors

### Changed
- Comprehensive release process documentation

### Removed
- AI assistant files from repository

## [0.1.0] - Initial Release

### Added
- Initial release of bc4 - Basecamp CLI tool
- Core authentication functionality
- Basic account and project management
- Command-line interface with version support