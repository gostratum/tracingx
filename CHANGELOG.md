# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Config.Sanitize()** now returns `any` instead of `Config` to implement `logx.Sanitizable` interface
  - This enables automatic secret sanitization when logging with `logx.Any()`
  - Secret-like OTLP headers (token, key, secret, authorization) are automatically redacted when configs are logged
  - **BREAKING:** If you were using the return value directly, add type assertion: `sanitized := cfg.Sanitize().(Config)`
  - Tests added to verify `Sanitizable` interface compliance
- **NewConfig()** no longer calls `Sanitize()` before returning; sanitization happens automatically via `logx.Any()` when logging

## [0.2.0] - 2025-10-29

### Added
- None user-facing; release packaging and version bump for v0.2.0.

### Changed
- Update `github.com/gostratum/core` dependency to `v0.2.0`.

### Fixed
- Corrected module naming inconsistencies (module name aligned to `tracingx`).

### Refactored
- Internal type safety: replace `interface{}` with `any` across the module for clearer code and modern Go usage.

### Tests
- Added and updated tests; refactored logger initialization in tests to match new logging helpers.

### Build / CI
- Makefile and release scripts improved to support version management and streamlined releases.



## [0.1.5] - 2025-10-26

### Added

- Enhance Makefile and add version management scripts for improved release process

## [0.1.4] - 2025-10-26

### Fixed

- Update github.com/gostratum/core dependency to v0.1.8

### Added

- Add Sanitize and ConfigSummary methods to Config for improved security and diagnostics

## [0.1.3] - 2025-10-26

### Fixed

- Update github.com/gostratum/core dependency to v0.1.7

## [0.1.2] - 2025-10-26

### Changed

- Replace 'interface{}' with 'any' for improved type safety

### Fixed

- Update core dependency to v0.1.5

## [0.1.1] - 2025-10-26

### Fixed

- Correct module name from "tracingx" to "tracing"
- Update core dependency to v0.1.4 and refactor logger initialization in tests

### Added

- Add comprehensive tests for tracing module and update .gitignore

## [0.1.0] - 2025-10-26

### Added

- Add tracing module with OTLP and noop providers