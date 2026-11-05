# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-11-05

### Added
- Deterministic CI log investigator with structured JSON report
- Classification for formatting, go vet, compile, assertion mismatches, generic test failures, and unknown failures
- Structured source locations (file/line/column) and actionable remediation
- New report contract: `deterministic_findings` + `ai_analysis`
- AI investigation layer interface with noop/mock providers
- AI validation, redaction, sanitization, size limits, and resilience (timeouts/errors)
- Optional AI enrichment via environment variables (disabled by default)
- CLI flags: `--version`, `--help`
- Comprehensive tests for deterministic and AI components

### Changed
- Migrated from `failures` to `deterministic_findings` contract

### Security
- Secret redaction before any AI context preparation
- Prompt-injection sanitization
- No autonomous command execution or code changes
- AI disabled by default
