# Documentation

Start with the root [README](../README.md) for installation. These guides describe
the maintained library's v1.9.0 behavior.

| Your task | Read |
|---|---|
| Construct or parse a rule, combine explicit dates, or query occurrences | [Usage and examples](usage.md) |
| Choose a DTSTART/UNTIL form or understand a daylight-saving transition | [Time zones and date/time forms](timezones.md) |
| Upgrade imports, replacements or `ROption` literals | [Migration guide](../MIGRATION.md) |
| Find the implementation and its regression coverage | [Design and source map](design.md) |
| Check supported boundaries and known limitations | [Limitations](limitations.md) |
| Look up exported declarations and executable examples | [Go API reference](https://pkg.go.dev/github.com/kentonium3/rrule-go) |
| Find release-specific changes | [Changelog](../CHANGELOG.md) |

The migration guide owns upgrade instructions; the changelog owns release
history. The guides explain how to use the library without maintaining a second
API reference. Go source comments and `example_test.go` supply that reference.
