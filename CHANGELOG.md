# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

- Added support for `tags` on a scorecard.

### Changed

- Added more validation around level and check group keys within the checks map.
  - Depending on the scorecard's type, either `scorecard_level_key` or `scorecard_check_group_key` must be present
  - The value of the check's grouping key must match one of the defined groupings

## [0.2.2] - 2025-06-12

### Changed

- Fixed handling of associations when importing. When there is no previous key in state, we fallback to _generating_ one by converting the item's name to snake case.

## [0.2.1] - 2025-06-11

### Changed

- Updated TF examples to recommend the new version string `~> 0.2.0`

## [0.2.0] - 2025-06-11

### Changed

- BREAKING: Restructured the `dx_scorecard` resource so the following attributes are now maps instead of lists:

  - `levels`
  - `check_groups`
  - `checks`

  This allows the provider to keep track of the identity of these objects and pass their IDs to updates, so check history is preserved.

## [0.1.2] - 2025-06-10

### Changed

- Fixed handlings of points-based fields and output fields

## [0.1.1] - 2025-06-06

### Changed

- Updated plugin address to `registry.terraform.io/get-dx/dx` in code and example docs

## [0.1.0] - 2025-06-06

Initial published release.

### Added

- Provider
- `dx_scorecard` resource

[0.2.2]: https://github.com/get-dx/terraform-provider-dx/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/get-dx/terraform-provider-dx/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/get-dx/terraform-provider-dx/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/get-dx/terraform-provider-dx/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/get-dx/terraform-provider-dx/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/get-dx/terraform-provider-dx/releases/tag/v0.1.0
