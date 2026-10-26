# Changelog

Notable changes to this maintained fork are recorded here, newest first. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Released sections describe tagged code; Unreleased describes changes on main that
are not yet tagged. Future intentions belong in [ROADMAP.md](ROADMAP.md).

## [Unreleased]

### Added

- **Project priorities and ways to help are documented.** The [roadmap](ROADMAP.md)
  separates RFC 5545 behavior from scheduling extensions. The README identifies
  the maintainer and welcomes co-maintainer and issue-triage contributions
  ([#11](https://github.com/kentonium3/rrule-go/pull/11)).
- **A changelog records this fork's release history and unreleased changes.**
- **Maintainers can validate the changelog and extract finalized release notes**
  with `go run ./tools/changelog`. CI checks the format on pull requests and main;
  version tags also require matching finalized notes.

### Changed

- **RFC 5545 is the documented authority for recurrence behavior.** Where the
  original port or python-dateutil disagrees with it, this fork follows the RFC
  ([#10](https://github.com/kentonium3/rrule-go/pull/10)).
- **The default branch is now `main`.** New pull requests target `main`
  ([#8](https://github.com/kentonium3/rrule-go/pull/8)).

### Fixed

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

[Unreleased]: https://github.com/kentonium3/rrule-go/compare/v1.8.3...main
[1.8.3]: https://github.com/kentonium3/rrule-go/tree/v1.8.3
