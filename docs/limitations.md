# Supported boundaries and known limitations

These describe the maintained v1.9.0 implementation. They are not a promise of
complete iCalendar component support. The [changelog](../CHANGELOG.md) owns
release-specific corrections, and the [time-zone guide](timezones.md) explains
date/time behavior.

## Representation and compatibility

- Recurrence calculations use second precision; sub-second components are
  truncated. BYSECOND accepts 0–59, not leap-second value 60.
- Calendar iteration is bounded by year 9999. When UNTIL is omitted, the effective
  default is DTSTART plus the largest Go `time.Duration`, approximately 290 years.
  Use an explicit DTSTART and an appropriate COUNT or UNTIL for predictable work.
- Go 1.16 remains supported. The fuzz-test file requires Go 1.18 or later.
- External positional `ROption` literals must use keyed fields in v1.9.0. See
  [migration](../MIGRATION.md) for this and the module-path change.

## Search cost and materialization

`All()` materializes the resulting slice. Large recurrence sets can therefore
use substantial memory. An iterator avoids accumulating the entire output, but
limiting emitted results alone does not bound the search for a never-matching
occurrence. The public iterator has no context/cancellation parameter.

Unreachable sub-daily clock selections terminate, and UTC rules that never match
stop at their effective UNTIL boundary. The corresponding scan optimization is
limited to UTC DTSTART; a never-matching rule in another location can still scan
to year 9999. Do not treat COUNT as a general CPU-cost limit. Applications handling
untrusted rules should retain their own validation and resource boundaries.

## Sets and text parsing

`Set` has one effective RRULE, with RDATE additions and EXDATE exclusions.
Duplicate instants are removed and EXDATE wins over an inclusion. Multiple parsed
RRULE lines retain their historical last-wins behavior. EXRULE is ignored for
evaluation, although its date/time parts are checked during parsing. Do not use
it as an exclusion-rule implementation.

RDATE and EXDATE parsing accepts DATE and DATE-TIME values. DATE inputs become
midnight instants in the applicable parsing location; do not assume their DATE
lexical form is preserved during serialization. PERIOD values are not supported.
This package is a recurrence library, not a complete VCALENDAR event parser or
an application scheduling/notification service.

## Timezone data and unusual transitions

Named zones require usable timezone data in the Go deployment environment.
Transition handling is tested against several real zones, half-hour changes and
synthetic timezone data. Its local transition profiling assumes isolated
transitions within a 72-hour window; it does not guarantee every possible custom
or historical sequence of discontinuities. Review unusual custom-zone use
against its actual rules rather than relying only on the zone's name.
