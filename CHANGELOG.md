# Changelog

Notable changes to this maintained fork are recorded here, newest first. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Released sections describe tagged code; Unreleased describes changes on main that
are not yet tagged. Future intentions belong in [ROADMAP.md](ROADMAP.md).

## [Unreleased]

## [1.9.0] - 2026-10-29

### Breaking

- **The v1.9.0 module path changes to `github.com/kentonium3/rrule-go`.**
  Change source imports from `github.com/teambition/rrule-go`; the package qualifier
  remains `rrule`. Follow the [migration guide](MIGRATION.md) for dependency and
  replacement changes.

- **External positional `ROption` literals must migrate to keyed literals.**
  Before: all fields could be supplied positionally. After: unexported parsing
  metadata makes external positional literals invalid. Use, for example,
  `rrule.ROption{Freq: rrule.DAILY, Count: 3}`. Exported signatures and keyed
  callers remain supported. This compatibility exception targets v1.9.0
  ([#6](https://github.com/kentonium3/rrule-go/issues/6)).

### Upgrade Notes

- **Check retained old-path dependencies before removing a replacement.** Removing
  the protective replacement can restore upstream code without this fork's
  panic/hang fixes, even if no rrule types cross APIs. Coordinate dependency
  migration first, or keep the pinned v1.8.3 arrangement. See the
  [dependency preflight](MIGRATION.md#check-dependencies-before-changing-replacements).

- **Explicit DTSTART/UNTIL values must use a permitted pair.** DATE starts require
  DATE UNTIL; floating starts require floating UNTIL; UTC or TZID starts require
  UTC UNTIL. Mismatches, TZID on DATE or UTC values, inconsistent VALUE parameters,
  and conflicting inline/shared starts now return errors. Update those inputs
  before upgrading ([#6](https://github.com/kentonium3/rrule-go/issues/6)).

### Added

- **Issue submission forms separate bug reports, enhancement requests, and questions.**
  Forms collect recurrence examples and optional area/impact information without
  adding kind or priority labels. Security reports use the private reporting channel.
- **Project priorities and ways to help are documented.** The [roadmap](ROADMAP.md)
  separates RFC 5545 behavior from scheduling extensions. The README identifies
  the maintainer and welcomes co-maintainer and issue-triage contributions
  ([#11](https://github.com/kentonium3/rrule-go/pull/11)).
- **A changelog records this fork's release history and unreleased changes.**
- **Maintainers can validate the changelog and extract finalized release notes**
  with `go run ./tools/changelog`. CI checks the format on pull requests and main;
  version tags also require matching finalized notes.

### Changed

- **Parsed floating and DATE values retain their forms in serialized rules.**
  Before: floating values could acquire UTC syntax. After: floating values omit
  `Z`, DATE starts use canonical `VALUE=DATE`, and parsed gap starts retain their
  original nominal text. Ordinary constructors and bare fragments retain their
  documented UTC-UNTIL serialization conventions
  ([#6](https://github.com/kentonium3/rrule-go/issues/6)).

- **RFC 5545 is the documented authority for recurrence behavior.** Where the
  original port or python-dateutil disagrees with it, this fork follows the RFC
  ([#10](https://github.com/kentonium3/rrule-go/pull/10)).
- **The default branch is now `main`.** New pull requests target `main`
  ([#8](https://github.com/kentonium3/rrule-go/pull/8)).

### Fixed

- **Generated DST-gap instances are skipped before COUNT and BYSETPOS.**
  Before: nonexistent wall times could be normalized or duplicated and folds
  could choose the later offset. After: all frequencies skip generated gaps and
  choose the first fold occurrence, preserving wall-clock selectors. Parsed gap
  DTSTART uses the pre-gap offset while preserving nominal cadence, with one
  positional rank per distinct instant. A supplied `time.Time` remains an
  authoritative anchor ([#5](https://github.com/kentonium3/rrule-go/issues/5)).
  Retained transition scenarios from
  [teambition PR 66](https://github.com/teambition/rrule-go/pull/66) and TZID-on-UTC
  rejection from [PR 68](https://github.com/teambition/rrule-go/pull/68) are credited;
  expected behavior was independently derived from RFC 5545.

- **Repeated BYHOUR, BYMINUTE, and BYSECOND values no longer duplicate occurrences
  or consume extra COUNT positions.** Previously, repeated time components could
  shorten the distinct recurrence set. Thanks to
  [@ect2000](https://github.com/ect2000)
  ([#9](https://github.com/kentonium3/rrule-go/pull/9),
  [#7](https://github.com/kentonium3/rrule-go/issues/7)).

## [1.8.3] - 2026-10-25

First release of this maintained fork, based on teambition/rrule-go. It retains
the module path `github.com/teambition/rrule-go`, the public API, and Go 1.16 support.

### Fixed

- **Accepted recurrence rules no longer panic on iterator mask bounds or calendar
  advancement.** Out-of-range BYDAY ordinals, BYEASTER offsets, and very large
  INTERVAL values previously could trigger index errors. Impossible candidates
  are skipped or iteration stops safely
  ([9340b1e](https://github.com/kentonium3/rrule-go/commit/9340b1e)).
- **Unreachable sub-daily clock selections terminate.** HOURLY, MINUTELY, and
  SECONDLY rules previously could loop forever when their INTERVAL could not
  reach the requested BYHOUR, BYMINUTE, or BYSECOND values
  ([aaf7495](https://github.com/kentonium3/rrule-go/commit/aaf7495)).
- **UTC rules that never match stop at their effective UNTIL boundary.** Previously,
  they could scan to the year 9999. This correction is limited to UTC DTSTART;
  other locations retain their previous scan behavior
  ([fadad75](https://github.com/kentonium3/rrule-go/commit/fadad75)).

Earlier history belongs to the [original project](https://github.com/teambition/rrule-go).

[Unreleased]: https://github.com/kentonium3/rrule-go/compare/v1.9.0...main
[1.9.0]: https://github.com/kentonium3/rrule-go/compare/v1.8.3...v1.9.0
[1.8.3]: https://github.com/kentonium3/rrule-go/tree/v1.8.3
