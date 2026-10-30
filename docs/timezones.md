# Time zones and date/time forms

The maintained fork follows [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545).
The distinction between a parsed civil time and a supplied Go `time.Time` matters
at clock transitions. For upgrade steps, use the [migration guide](../MIGRATION.md).

## Choose a permitted pair

A complete parsed rule with an explicit DTSTART validates its UNTIL against the
start's textual form:

| DTSTART form | Example | Permitted UNTIL form |
|---|---|---|
| DATE | `DTSTART;VALUE=DATE:20260105` | DATE, such as `UNTIL=20260131` |
| Floating DATE-TIME | `DTSTART:20260105T100000` | Floating DATE-TIME, without `Z` |
| UTC DATE-TIME | `DTSTART:20260105T100000Z` | UTC DATE-TIME, with `Z` |
| DATE-TIME with TZID | `DTSTART;TZID=America/New_York:20260105T100000` | UTC DATE-TIME, with `Z` |

Do not combine TZID with a DATE or a UTC (`Z`) value. VALUE parameters must agree
with their values. Conflicting inline and component DTSTART declarations are
rejected. UNTIL is inclusive; it is not a count of occurrences.

## Floating values need an evaluation location

Floating text contains no zone identifier. The default parser evaluates it in
UTC. Use `StrToROptionInLocation`, followed by `NewRRule`, to select an evaluation
location. Set parsing also has `StrSliceToRRuleSetInLoc` for a default location.
Floating serialization still omits `Z` and TZID; the evaluation location is not
encoded by turning the text into a zoned rule.

Load required IANA locations with Go's `time.LoadLocation` and handle errors.
Deployment environments need usable timezone data. A caller-supplied custom
location is retained as that actual location, not reloaded solely by its name.

## Generated gaps and folds

For every frequency, generated nonexistent local times are skipped before COUNT
and BYSETPOS selection. A skipped gap does not consume COUNT. A generated
ambiguous time in a backward transition uses its first occurrence. Distinct
selection positions do not emit the same instant twice.

This preserves wall-clock selectors rather than treating every daily interval
as exactly 24 elapsed hours. Half-hour transitions are covered as well as
one-hour changes. See the [regression/source map](design.md) and
[implementation limits](limitations.md).

## An explicit parsed gap anchor is different

An explicit DTSTART in a gap is interpreted using the pre-gap offset and retains
its nominal civil cadence. Later generated gap candidates are still skipped.
For example:

```go
package main

import (
	"fmt"
	"time"

	"github.com/kentonium3/rrule-go"
)

func main() {
	r, err := rrule.StrToRRule("DTSTART;TZID=America/New_York:20230312T023000\nRRULE:FREQ=DAILY;COUNT=2")
	if err != nil {
		panic(err)
	}
	for _, occurrence := range r.All() {
		fmt.Println(occurrence.UTC().Format(time.RFC3339))
	}
}
```

Output:

```text
2023-03-12T07:30:00Z
2023-03-13T06:30:00Z
```

The serialized parsed DTSTART keeps the original `20230312T023000` civil text.
It does not replace the intended next-day 02:30 cadence with 03:30.

## Supplied Go times are authoritative instants

`NewRRule` accepts a `time.Time`, including an explicitly supplied later fold
instant. The library does not reinterpret that anchor as the first fold instant.
Go may already normalize a nonexistent civil time when `time.Date` constructs
it; the library cannot recover a discarded textual intention from that value.
Parse explicit DTSTART text when its original civil form is important.

Copied parsed `ROption` values retain their parsing descriptors while the relevant
instant and actual location remain unchanged. Replacing DTSTART/UNTIL values, or
calling their public setters, replaces the corresponding parsed semantics. Do
not use a setter merely to preserve textual metadata.

## Full rules, fragments and serialization

Parsed floating and DATE values retain their forms in `String()`. DATE starts
use canonical `VALUE=DATE`; floating values omit `Z`. A bare RRULE fragment without
DTSTART keeps the historical UTC-with-`Z` UNTIL serialization convention. Ordinary
programmatic options use that convention too.

Validation happens while parsing the complete text. Replacing `opt.Dtstart`
afterward cannot rescue an invalid parsed DTSTART/UNTIL pair. Programmatic
`NewRRule` options are not retroactively checked as a textual pair. Removing
DTSTART is not a general migration technique: it can lose the anchor, implicit
selectors and nominal cadence. See the [parsing migration caveat](../MIGRATION.md#check-dtstart-parsing-and-serialization).
