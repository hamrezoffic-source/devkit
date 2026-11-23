# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `devkit --version` / `-v`, set at build time from the release tag.
- `go vet` and coverage reporting in CI.
- Contributing guide, security policy, issue and pull request templates.

## [0.1.0] - 2026-11-23

### Added
- `json format`, `json to-yaml`, `yaml to-json` and `port check` commands.
- Cross-platform release builds for Linux, macOS and Windows (amd64, arm64).

### Fixed
- JSON to YAML no longer turns numbers into quoted strings.
- macOS CI builds with current stable Go.
