# Design and source map

rrule-go is one Go package. Exported declarations and executable examples are
documented in the [Go API reference](https://pkg.go.dev/github.com/kentonium3/rrule-go).
This page helps contributors find the implementation; it does not duplicate the
API or define new recurrence behavior.

## From input to occurrences

1. `str.go` parses text into options and date/time descriptors, validates complete
   DTSTART/UNTIL pairs, and serializes rules and sets.
2. `rrule.go` validates option bounds, applies defaults, prepares selectors and
   constructs the recurrence iterator.
3. `civil_time.go` preserves parsed nominal anchors and handles local transition
   candidates, including gap filtering and fold choice.
4. `rruleset.go` merges the effective rule with explicit dates and exclusions.
5. `util.go` supplies iterator/query helpers and calendar utilities.

`Iterator` produces one instant at a time. Convenience methods such as `All`,
`Between`, `Before` and `After` create iterators; they are not cached queries.
Parsing metadata is private implementation state. Callers should construct
`ROption` with keyed fields; see [migration](../MIGRATION.md).

## Where behavior is checked

| Files | Primary coverage |
|---|---|
| `example_test.go` | Executable public API examples |
| `rrule_test.go`, `str_test.go`, `rruleset_test.go` | Recurrence, parsing, serialization and set behavior |
| `rrule_safety_test.go`, `rrule_reachability_test.go` | Accepted-input safety and unreachable clock selections |
| `rrule_until_test.go` | UNTIL scan behavior and regression expectations |
| `rrule_dst_test.go` | Gaps, folds, nominal anchors and positional selection |
| `str_timeforms_test.go`, `str_timeforms_api_test.go` | Date/time forms and external callers |
| `rrule_fuzz_test.go` | Bounded fuzz coverage on Go 1.18 and later |
| `tools/changelog/` | Changelog validation and release-note extraction |

Keep the single package and its tests at the repository root. A `src/` or `lib/`
wrapper would add structure without improving the import API. `tools/` is for
maintainer utilities, not runtime dependencies.

## Documentation maintenance

Give each topic one canonical home: usage here in the guides, upgrade actions in
the root migration guide, release history in the changelog, and exported API
reference in Go comments/examples. Update the relevant page when behavior
changes, rather than copying a rule into several places.

Before proposing documentation changes, check examples against the current
library, verify relative links and heading anchors, and identify any intended
future behavior as a proposal. Keep existing root document URLs stable. Do not
turn private development records or application data into public examples.

## Fixture directories

Go's `testdata/` convention is suitable for reusable input/output fixtures read
by tests. It is not required for ordinary inline test tables, nor is it an API
source directory. This repository currently keeps its regression tables,
synthetic timezone data and fuzz seeds in Go tests.

Add fixture files only when they improve a real test. Record their schema,
provenance and expected semantics, keep them deterministic and public-safe, and
make missing/malformed data fail the test. A golden output must have a reviewed
reason to change; it must not be regenerated merely to make a failure disappear.
