> **This is a maintained fork of [teambition/rrule-go](https://github.com/teambition/rrule-go).**
> The original project is no longer maintained. This is a modified copy with Go 1.16 support.
> Version v1.9.0 uses module path `github.com/kentonium3/rrule-go`.
> Existing callers must change imports, and external positional `ROption` literals must use
> keyed fields. See [MIGRATION.md](MIGRATION.md) before upgrading.
>
> [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545) is the reference for how rules behave. Where
> the original code or python-dateutil differs from it, this fork follows the RFC.

## Changes in v1.9.0

Generated daylight-saving gap times are skipped before COUNT and BYSETPOS, and
folds choose the first occurrence. Parsed floating and DATE values retain their
forms; explicit DTSTART/UNTIL pairs are validated. Repeated time selectors no
longer duplicate occurrences. The [changelog](CHANGELOG.md) describes the behavior
changes and credits, and the [migration guide](MIGRATION.md) covers imports,
replacements, keyed `ROption` literals and parsing compatibility.

[v1.9.0 is released](https://github.com/kentonium3/rrule-go/releases/tag/v1.9.0).

## Fixes released in v1.8.3

- **Panics on accepted rules.** BYDAY ordinals outside the month, BYEASTER offsets outside the year,
  and very large INTERVAL values could make the iterator panic with "index out of range". They now
  either yield no occurrence for the impossible parts or stop safely.
- **A hang.** HOURLY, MINUTELY and SECONDLY rules whose BYHOUR, BYMINUTE or BYSECOND values can't be
  reached from DTSTART at the given INTERVAL looped forever. They now finish with no occurrence.
- **Long scans with UNTIL (UTC only).** When DTSTART is in UTC, a rule that never matches now stops
  at UNTIL instead of scanning to the year 9999. Other locations behave as before.

The v1.8.3 fixes preserve occurrences for rules that worked before that release.
They come with regression tests and a fuzz test.

See [CHANGELOG.md](CHANGELOG.md) for released and unreleased changes.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for project priorities, sequencing, and the distinction
between RFC 5545 behavior and useful scheduling extensions.

## Maintainer and help wanted

This fork is maintained by [Kent Gale (@kentonium3)](https://github.com/kentonium3).

**Co-maintainer wanted.** Kent is looking for an experienced Go developer to help
share ongoing maintenance: reviewing changes, investigating bugs, improving tests,
and shaping releases. Experience with recurrence rules, calendar standards, or
timezones would be especially useful.

**Issue-triage contributors are welcome, too.** If a full co-maintainer role is
more than you want to take on, help reproducing reports, reducing test cases,
identifying duplicates, clarifying expected behavior, and proposing labels or
milestones would be valuable. This can be a smaller, focused contribution;
responsibilities and repository access can be agreed individually.

To express interest in either role, [open an issue](https://github.com/kentonium3/rrule-go/issues/new)
with a short introduction, your relevant experience, and the kind of help you
would like to offer. Small contributions are welcome.

## Using it

For v1.9.0, import `github.com/kentonium3/rrule-go` (package name `rrule`).
Install the released version:

```bash
go get github.com/kentonium3/rrule-go@v1.9.0
go mod tidy
```

Existing users must follow [MIGRATION.md](MIGRATION.md) before removing old-path
replacements. Consumers staying on v1.8.3 can retain the older arrangement:

```bash
go mod edit -replace github.com/teambition/rrule-go=github.com/kentonium3/rrule-go@v1.8.3
go mod tidy
```

That v1.8.3 arrangement uses imports of `github.com/teambition/rrule-go`.

Report problems in this repository's issues. Licensed under MIT, as the original (© Teambition).

---

# rrule-go

Go library for working with recurrence rules for calendar dates.

[![CI](https://github.com/kentonium3/rrule-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kentonium3/rrule-go/actions/workflows/ci.yml)
[![CodeQL](https://github.com/kentonium3/rrule-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/kentonium3/rrule-go/actions/workflows/codeql.yml)
[![License](https://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://github.com/kentonium3/rrule-go/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kentonium3/rrule-go.svg)](https://pkg.go.dev/github.com/kentonium3/rrule-go)

The rrule module implements the recurrence rules of the iCalendar standard,
[RFC 5545](https://www.rfc-editor.org/rfc/rfc5545). It began as a partial port of the rrule module from
the excellent [python-dateutil](https://github.com/dateutil/dateutil) library.

## Documentation

- [Documentation index](docs/README.md): guides by task.
- [Usage and examples](docs/usage.md): rule construction, parsing, sets and queries.
- [Time zones](docs/timezones.md): DTSTART/UNTIL forms, gaps, folds and serialization.
- [Design](docs/design.md) and [limitations](docs/limitations.md): contributor map and supported boundaries.
- [Migration](MIGRATION.md) and [changelog](CHANGELOG.md): upgrades and release history.

## Demo

The full demonstrations moved to the [usage guide](docs/usage.md#demonstrations).
Existing section links remain below for compatibility.

### rrule.RRule

[Rule construction and queries](docs/usage.md#rrulerrule).

### rrule.Set

[Rule sets and explicit dates](docs/usage.md#rruleset).

### rrule.StrToRRule

[Parsing a rule](docs/usage.md#rrulestrtorrule).

### rrule.StrToRRuleSet

[Parsing a rule set](docs/usage.md#rrulestrtorruleset).

## License

rrule-go is licensed under the [MIT](LICENSE) license.
Copyright &copy; 2017-2023 [Teambition](https://www.teambition.com).
