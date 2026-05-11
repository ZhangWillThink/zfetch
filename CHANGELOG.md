# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2026-05-11

### Added

- Auto-load `~/.config/zfetch/config.jsonc` when present (no `-c` required).
- `zfetch upgrade` verifies `SHA256SUMS` when published for a release (warns only if absent).
- `config.ListPresetNames()` and dynamic output for `--list-presets`.
- `scripts/build.sh` stamps `internal/upgrade.CurrentVersion` via `-ldflags`; release CI passes `ZFETCH_VERSION` from the tag.
- Installation script default `ZFETCH_VERSION=latest`.
- CI path filters so workflows run only when relevant paths change.

### Changed

- Module execution uses a small worker pool (up to 8) instead of unbounded goroutines.
- Windows: batch PowerShell for resolution/host/locale; merged battery WMI query; GPU/package commands use detected `pwsh`/`powershell`.

### Documentation

- README / AGENTS: static binary wording, build/version notes, preset search paths.

## [0.5.1] - 2026-05-11

### Added

- GitHub Actions `release` workflow: cross-build artifacts, `install.sh`, and `SHA256SUMS` on `v*` tags.
- Correct Windows upgrade asset name (`*.exe`).
