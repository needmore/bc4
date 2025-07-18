# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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