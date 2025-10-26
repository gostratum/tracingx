# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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