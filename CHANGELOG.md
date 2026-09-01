# Changelog

Notable changes to YaYeet are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-09-01

### Changed

- Hash installed game archives with the native Go xxHash implementation instead of the external `xxhsum` command.
- Expand the README with installation, configuration, development, and version-filter documentation.

## [1.1.0] - 2026-09-01

### Added

- Configuration filters for unstable, test, and revision releases.
- Tests for revision filtering and display formatting.

### Changed

- Show the home screen immediately while version information loads in the background.
- Keep the Play button disabled until version loading and installed-version detection finish.
- Run configuration saving, installation detection, and game startup without blocking the UI.
- Reorganize the home screen around a 1280 by 720 default size.
- Select the latest stable release by default and list newer versions first.
- Mark stable, unstable, and test releases independently from their latest status.
- Display manifest revisions such as `0.9.0_0015_1` as `0.9.0-15.1`.
- Hide older revisions and their revision numbers unless `Show revisions` is enabled.

## [1.0.0] - 2026-09-01

### Added

- Initial Linux launcher for installing and running Voices of the Void through Wine or Proton.

[1.2.0]: https://github.com/Huijiro/YaYeet/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/Huijiro/YaYeet/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/Huijiro/YaYeet/releases/tag/v1.0.0
