# Release Process

This document describes the automated release process for bc4.

## Overview

bc4 uses a fully automated release process powered by GitHub Actions and GoReleaser. When you push a version tag, the entire release process happens automatically:

1. Tests are run
2. Binaries are built for macOS (Intel & Apple Silicon)
3. A GitHub release is created with changelog
4. The Homebrew formula is updated
5. Users can immediately install or upgrade via Homebrew

## Creating a Release

### 1. Prepare Your Changes

Ensure all changes for the release are merged to the `main` branch:

```bash
git checkout main
git pull origin main
```

### 2. Create and Push a Version Tag

Create an annotated tag following [semantic versioning](https://semver.org/):

```bash
# For a patch release (bug fixes)
git tag -a v0.1.1 -m "Fix authentication bug"

# For a minor release (new features, backwards compatible)
git tag -a v0.2.0 -m "Add support for message boards"

# For a major release (breaking changes)
git tag -a v1.0.0 -m "First stable release with new API"

# Push the tag to trigger the release
git push origin v0.1.1
```

### 3. Monitor the Release

The release process takes approximately 2-3 minutes:

1. Go to the [Actions tab](https://github.com/needmore/bc4/actions)
2. Watch the "Release" workflow progress
3. Once complete, check the [Releases page](https://github.com/needmore/bc4/releases)

## What Happens During a Release

The automated process handles everything:

### GitHub Release
- Creates a new GitHub release
- Uploads the universal macOS binary
- Generates a changelog from commit messages
- Includes installation instructions

### Homebrew Formula
- Updates `Formula/bc4.rb` with the new version
- Updates SHA256 checksums automatically
- Commits the changes back to the repository
- Formula is immediately available for installation

### Binary Distribution
- Builds for both Intel (amd64) and Apple Silicon (arm64)
- Creates a universal binary that works on all macOS systems
- Includes all necessary files (LICENSE, README)

## Version Numbering

Follow [semantic versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backwards compatible
- **PATCH** (0.0.1): Bug fixes

Examples:
- `v0.1.0` → `v0.1.1`: Bug fix
- `v0.1.1` → `v0.2.0`: New feature added
- `v0.2.0` → `v1.0.0`: Breaking change

## Pre-releases

For beta or release candidate versions:

```bash
# Beta release
git tag -a v1.0.0-beta.1 -m "Beta release for v1.0.0"

# Release candidate
git tag -a v1.0.0-rc.1 -m "Release candidate for v1.0.0"
```

## Installation Methods

Once released, users can install bc4 via:

### Homebrew (Recommended)
```bash
# First time installation
brew tap needmore/bc4
brew install bc4

# Upgrading
brew update
brew upgrade bc4
```

### Direct Download
Users can download the binary directly from the GitHub releases page.

## Commit Message Format

For better changelogs, use conventional commit messages:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `chore:` Maintenance tasks
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `perf:` Performance improvements

Examples:
```bash
git commit -m "feat: add support for document uploads"
git commit -m "fix: resolve OAuth token refresh issue"
git commit -m "docs: update API authentication guide"
```

## Troubleshooting

### Release Failed

If the release workflow fails:

1. Check the [Actions logs](https://github.com/needmore/bc4/actions)
2. Fix any issues
3. Delete the tag locally and remotely:
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```
4. Create and push the tag again

### Testing Locally

Before creating a release, you can test the build:

```bash
# Test the build
make build

# Run tests
make test

# Check version injection
./build/bc4 version
```

## Release Checklist

Before creating a release:

- [ ] All tests passing (`make test`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No linting issues (`go vet ./...`)
- [ ] Version-specific documentation updated
- [ ] Any breaking changes clearly documented

## Rollback Procedure

If a release has critical issues:

1. Create a new patch release with the fix
2. If necessary, delete the GitHub release (but keep the tag for history)
3. Communicate with users about the issue

## Security Releases

For security fixes:
1. Fix the vulnerability
2. Create a patch release immediately
3. Update the GitHub Security Advisory (if applicable)
4. Mention security fix in release notes

## Questions?

For questions about the release process, please open an issue on GitHub.