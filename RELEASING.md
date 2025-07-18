# Releasing bc4

This document describes the process for releasing new versions of bc4.

## Prerequisites

- You must have write access to the bc4 repository
- Go 1.21 or later installed
- GoReleaser installed locally (for testing): `brew install goreleaser`
- GPG key for signing (optional)

## Release Process

### 1. Prepare the Release

1. Ensure all changes for the release are merged to `main`
2. Run tests locally: `make test`
3. Update version references if needed (though most are automated)

### 2. Create and Push a Tag

```bash
# Fetch latest changes
git checkout main
git pull origin main

# Create a new tag following semantic versioning
# For a new release:
git tag -a v1.0.0 -m "Release v1.0.0"

# For a pre-release:
git tag -a v1.0.0-beta.1 -m "Pre-release v1.0.0-beta.1"

# Push the tag
git push origin v1.0.0
```

### 3. Monitor the Release

1. Go to the [Actions tab](https://github.com/needmore/bc4/actions)
2. Watch the "Release" workflow
3. Once completed, check the [Releases page](https://github.com/needmore/bc4/releases)

### 4. Verify Homebrew Formula Update

GoReleaser will automatically update the Homebrew formula in the `Formula` directory of this repository. After the release workflow completes:

1. Check that the formula was updated:
   ```bash
   git pull origin main
   cat Formula/bc4.rb
   ```

2. The formula should have been automatically updated with the new version and SHA256 checksums.

3. If you need to make manual adjustments, you can edit `Formula/bc4.rb` and commit the changes.

### 5. Test Installation

Test that users can install the new version:

```bash
# For first-time installation
brew install needmore/bc4/bc4

# For updates
brew update
brew upgrade bc4

# Verify
bc4 --version
```

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- Increment MAJOR for breaking changes
- Increment MINOR for new features
- Increment PATCH for bug fixes

Pre-release versions:
- Alpha: `v1.0.0-alpha.1`
- Beta: `v1.0.0-beta.1`
- Release Candidate: `v1.0.0-rc.1`

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
4. Start the process again

### GoReleaser Issues

Test GoReleaser locally before pushing a tag:

```bash
# Dry run
goreleaser release --snapshot --clean

# Check the dist/ directory for output
ls -la dist/
```

## Release Checklist

- [ ] All tests passing
- [ ] CHANGELOG.md updated (if maintaining one)
- [ ] Version tag created and pushed
- [ ] GitHub release created successfully
- [ ] Binaries uploaded to release
- [ ] Homebrew formula updated
- [ ] Installation tested via Homebrew
- [ ] Announcement made (if applicable)