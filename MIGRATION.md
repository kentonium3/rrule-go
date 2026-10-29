# Migrating to the new module path

The candidate v1.9.0 changes the module from `github.com/teambition/rrule-go` to
`github.com/kentonium3/rrule-go`. The package name remains `rrule`. The steps below
prepare consumers for that release; this document does not establish that the
public v1.9.0 tag exists or can be downloaded. Existing immutable releases remain
available to consumers pinned to them.

## Check dependencies before changing replacements

Before removing an old-path replacement, inspect both module requirements and
package imports in your application and its dependencies:

```sh
go mod graph
go mod why -m github.com/teambition/rrule-go
go list -deps -test -f '{{.ImportPath}}: {{join .Imports " "}}' ./...
```

Look for `github.com/teambition/rrule-go` in the selected packages and their
imports. A module requirement alone does not prove which packages still use it;
a single `go mod why` chain is not a complete inventory. Repeat package inspection
for the build tags and platforms you actually support, including relevant tests.

If dependencies still import the old path, removing the replacement can select
original upstream code again and lose this fork's panic/hang fixes. This risk
exists even when no rrule types cross an API boundary. Coordinate those
dependencies' migration first, or keep the existing pinned v1.8.3 replacement
arrangement until ready. Passing application tests alone does not prove those
fixes remain active. Do not delete legitimate old transitive dependencies merely
to make the graph look clean.

Types declared under different import paths are distinct Go named types. If your
APIs exchange values such as `rrule.ROption` or `*rrule.RRule` with a dependency,
migrate both sides together. This module-path change adds no compatibility bridge.

## Change imports and select your dependency case

After the preflight, change application imports:

```go
// Before:
import "github.com/teambition/rrule-go"
// After:
import "github.com/kentonium3/rrule-go"
```

Keep calls such as `rrule.NewRRule` and `rrule.DAILY` unchanged. Select one case
below; the two replacement-removal commands are alternatives.

### Direct upstream dependency without a replacement

Change the imports and omit the replacement-removal step. Continue with
[install and test](#install-and-test-after-public-release).

### Replacement applying to all old-module versions

For a directive such as:

```go
replace github.com/teambition/rrule-go => github.com/kentonium3/rrule-go v1.8.3
```

Remove it only after the dependency preflight permits migration:

```sh
go mod edit -dropreplace=github.com/teambition/rrule-go
```

Then continue with install and test.

### Replacement with a version-specific left side

For a directive such as:

```go
replace github.com/teambition/rrule-go v1.8.2 => github.com/kentonium3/rrule-go v1.8.3
```

Use the actual version on the **left** side, not the replacement's v1.8.3:

```sh
go mod edit -dropreplace=github.com/teambition/rrule-go@v1.8.2
```

Then continue with install and test.

## Install and test after public release

Run these commands only once v1.9.0 is actually published:

```sh
go get github.com/kentonium3/rrule-go@v1.9.0
go mod tidy
go test ./...
```

Review `go.mod` and `go.sum`, check selected sources and imports again, and run
your normal application checks for supported configurations. Tidy removes unused
direct requirements; retain any old requirement that is legitimately still used,
subject to the protective-fix preflight above.

## Separate ROption literal migration

The candidate v1.9.0 also inherits an independent compatibility change:
unexported parse metadata makes external positional `ROption` literals invalid.
Convert them to keyed literals, for example:

```go
rrule.ROption{Freq: rrule.DAILY, Count: 3}
```

Exported signatures and keyed callers remain supported. See
[v1.9.0 changes](CHANGELOG.md) for the accompanying recurrence
corrections and input-validation changes.

## Check DTSTART parsing and serialization

Inventory whether calls to `StrToROption` or `StrToRRule` parse a complete
`DTSTART` plus `RRULE`, or a bare rule fragment. A parsed `DTSTART` enables
date/time-form validation: an invalid `DTSTART`/`UNTIL` pair can fail before
your code replaces `opt.Dtstart`. Ordinary programmatic `NewRRule` options are
not retroactively subject to that textual-pair validation. See the
[DTSTART/UNTIL pairing upgrade note](CHANGELOG.md#upgrade-notes).

Valid parsed floating and DATE values now retain their forms when serialized;
bare fragments keep their documented UTC-UNTIL convention. Review stored-text
and round-trip expectations against the
[floating and DATE serialization change](CHANGELOG.md#changed).
Keep `DTSTART` when it defines the recurrence: stripping it can lose the anchor,
implicit selectors and nominal cadence across a daylight-saving gap. Any adapter
for an application's separate anchor or legacy text contract needs its own
compatibility checks; removing `DTSTART` is not a general migration step.

## Temporary local candidate checks

A maintainer may verify an unpublished checkout in a disposable consumer module
using a local replacement. After changing imports and applying the appropriate
old-replacement removal above, set `candidate_dir` to that checkout's absolute
path and use:

```sh
GOWORK=off GOPROXY=off go mod edit -require=github.com/kentonium3/rrule-go@v1.9.0
GOWORK=off GOPROXY=off go mod edit -replace=github.com/kentonium3/rrule-go="$candidate_dir"
GOWORK=off GOPROXY=off go mod tidy
GOWORK=off GOPROXY=off go list -m -json all
GOWORK=off GOPROXY=off go test ./...
```

This replacement is temporary verification scaffolding, not a production
installation recommendation. It proves only local candidate behavior; it does
not verify a public tag, download, checksum, or package-documentation indexing.
Discard the disposable fixture after checking it. Verify the real public
installation separately after release using the commands above.
