# rrule-go roadmap

Last updated: October 25, 2026

rrule-go is a maintained Go library for evaluating calendar recurrence rules. Our first priority is predictable, safe behavior under [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545.html). We also intend to add selected scheduling conveniences inspired by [python-dateutil](https://github.com/dateutil/dateutil), where they solve practical problems for Go applications.

This roadmap gives the order of priorities. Near-term work has a defined direction; later work needs more user feedback and design. It is not a release calendar or a promise to implement every feature listed. Release notes and tagged releases identify what has actually shipped.

## How we choose work

We distinguish three kinds of capability:

| Category | Meaning | Examples |
| --- | --- | --- |
| RFC behavior | Parsing and evaluating the recurrence features we support according to RFC 5545. | Skipping nonexistent generated local times; applying the required DTSTART/UNTIL value types. |
| Scheduling extensions and conveniences | Useful application APIs that go beyond the standard or make it easier to use. They are documented separately from RFC behavior. | Combining independent schedules, retrieving the next N occurrences, optional caching. |
| Broader date/time utilities | Functionality outside recurrence evaluation. Each addition needs its own justification and package boundary. | Calendar arithmetic resembling dateutil's `relativedelta`. |

The RFC is authoritative for standard behavior. Dateutil supplies useful designs, examples, and comparison cases, but matching Python output alone does not establish correctness. Existing extensions, such as BYEASTER, remain identified as extensions.

The project already supports all seven recurrence frequencies, the principal BY selectors, numbered weekdays, explicit inclusion and exclusion dates, lazy iteration, and Before/After/Between queries. The goal is to strengthen and extend that foundation, rather than port all of dateutil at once.

## Priority 1 — Complete the current correctness work

**Direction: planned and well defined. Correctness takes precedence over feature expansion.**

The immediate focus is the handling of civil times and recurrence value types, tracked in public issues [#5](https://github.com/kentonium3/rrule-go/issues/5) and [#6](https://github.com/kentonium3/rrule-go/issues/6).

### Daylight-saving transitions and civil-time schedules

- Skip generated local times that do not exist, before applying COUNT and BYSETPOS.
- Resolve ambiguous parsed local times to their first occurrence, as required by the RFC.
- Preserve wall-clock selector meanings through transitions, including half-hour changes.
- Distinguish explicit DTSTART interpretation from the treatment of generated candidates. Preserve the intended recurrence cadence when a parsed DTSTART lies in a gap.
- Keep iteration, queries, and sets consistent, with ordered results and no duplicate instants caused by transition handling.

### DTSTART and UNTIL forms

- Preserve floating DATE-TIME and DATE forms through parsing and serialization.
- Reject the specified forbidden DTSTART/UNTIL combinations with useful errors.
- Keep conformant TZID DTSTART plus UTC UNTIL working.
- Validate the relevant forms independently of input order, including shared DTSTART handling in sets.
- Preserve documented constructor and fragment conventions, and describe their boundaries clearly.

These changes intentionally alter some previously accepted inputs and transition-day results. Release notes must identify those changes and explain migration. Existing contributor fixes, including duplicate time-selector handling, remain part of the compatibility baseline.

**Compatibility note:** the planned v1.9.0 time-form implementation adds unexported metadata to `ROption`. Exported signatures and keyed literals remain available, but external positional `ROption` literals must migrate to keyed literals. For example:

```go
rrule.ROption{Freq: rrule.DAILY, Count: 3}
```

Completion means regression evidence for the corrected behavior, compatibility comparisons outside intentional corrections, minimum-Go validation, and independent review. New conveniences should not delay shipping these corrections.

## Priority 2 — Improve interoperability and everyday APIs

**Direction: planned. Exact API names and release assignments will be settled during design.**

These are relatively small additions to the existing recurrence model. Parser improvements come first, followed by query and construction helpers; independently reviewable changes can ship separately.

| Work item | Intended outcome | Boundaries | Initial effort estimate |
| --- | --- | --- | --- |
| Calendar-text interoperability | Accept folded recurrence content lines and handle permitted vendor parameters, including examples found in calendar exports. | Follow the RFC's parameter rules; do not silently discard unsupported recurrence semantics. This is not a full VCALENDAR parser. | 3–5 engineer-days |
| Caller-provided TZID resolution | Let applications supply a mapping or resolver for timezone identifiers that Go's normal location lookup cannot resolve. | Preserve existing location-based APIs. Define fallback and unknown-ID errors; make no automatic network lookup necessary. | 2–5 engineer-days |
| Bounded lazy queries | Provide a convenient way to retrieve up to N occurrences after a date without repeatedly starting a new query. | Build on existing iteration. Do not promise constant-time seeking into arbitrary schedules. Define inclusivity and exhaustion clearly. | 3–5 engineer-days |
| Safe copy and replacement | Make it straightforward to derive a revised rule while preserving its other settings. | Define ownership of option slices and parsed-time metadata. Revising one rule must not unexpectedly alter another. | 3–7 engineer-days |

These capabilities are inspired by dateutil's [recurrence and parsing APIs](https://dateutil.readthedocs.io/en/stable/rrule.html). The Go API should be idiomatic and explicit rather than reproduce every Python flag or operator.

Each item should include a public usage example, error and boundary tests, compatibility checks, and documentation. API decisions should preserve current callers wherever practical. The current Go 1.16 minimum remains a constraint unless a separately documented proposal changes it.

## Priority 3 — Explore schedule composition and repeated-query performance

**Direction: candidate work. User examples and design feedback are needed before commitments.**

### Compose independent schedules

Python's `rruleset` combines several inclusion rules. Go's current `Set` stores one inclusion rule, and assigning another replaces it. A future composition API could merge independently defined schedules into one ordered stream, remove duplicate instants, and apply explicit exclusions.

This is useful when one recurrence rule cannot express a schedule naturally, or when an application needs to combine schedules owned by different parts of the system. Simple patterns that already fit one rule should continue to use it.

Before implementation, we need to decide:

- Whether composition belongs in a new type or an additive API, while retaining existing `Set.RRule` replacement behavior.
- How component rules retain their own DTSTART, COUNT, UNTIL, and timezone semantics.
- What is copied or shared, and how later mutations affect an existing composite.
- How serialization distinguishes application-level composition from a standard calendar component.
- How large, overlapping, or never-matching component streams are bounded and tested.

**RFC boundary:** this is an application-level extension. RFC 5545 leaves the recurrence set from multiple RRULE properties in one component undefined. We should not present composition as repairing an RFC requirement. See [RFC 5545 §3.8.5.3](https://www.rfc-editor.org/rfc/rfc5545.html#section-3.8.5.3).

Initial estimate: **2–3 engineer-weeks**, subject to the API and ownership decisions above.

### Evaluate optional caching

Repeated queries currently begin with fresh iteration. Dateutil offers opt-in result caching; a Go equivalent may help workloads that repeatedly query the same long-lived schedule.

The first step is a benchmark study using realistic workloads. We need evidence about repeated-query latency, retained memory, and the effect of distant query dates before choosing a cache strategy. Any design must address memory limits, invalidation, mutation, and concurrent use explicitly. Caching must not change recurrence results or become a prerequisite for ordinary iteration.

Initial implementation estimate, if the benchmarks justify it: **2–4 engineer-weeks**.

### Consider recurring exclusions only with demonstrated demand

A composition API could eventually subtract another recurrence stream, supporting recurring blackout patterns. Python offers this through `exrule`; Go currently evaluates EXDATEs but does not apply EXRULE exclusions.

This would be an explicitly documented extension or legacy interoperability feature. RFC 5545 removed EXRULE from calendar components; it is not a missing strict-RFC requirement. See [Appendix A.3](https://www.rfc-editor.org/rfc/rfc5545.html#appendix-A.3).

Initial estimate: **1–2 additional engineer-weeks after composition**, with no commitment to implement it yet. In the meantime, documentation and parser behavior must make unsupported exclusion semantics clear.

## Priority 4 — Assess selected utilities beyond recurrence

**Direction: exploratory. These are proposals for feedback, not an implementation queue.**

The strongest broader candidate is calendar arithmetic similar to [`relativedelta`](https://dateutil.readthedocs.io/en/stable/relativedelta.html). It supports operations such as adding months with end-of-month clamping, replacing date components, adjusting to an nth weekday, and calculating calendar differences between dates.

That is meaningfully different from duration arithmetic and Go's normalization rules. For example, adding a month to January 31 can clamp to February's last day rather than normalize into March. A useful Go design needs explicit semantics and likely a separate package. A full implementation is initially estimated at **4–8 engineer-weeks**; a smaller subset should be defined from concrete needs first.

Other possibilities remain uncommitted:

| Candidate | What would justify it | Initial effort estimate |
| --- | --- | --- |
| Public gap/fold utilities | Applications need to inspect ambiguous or nonexistent wall times before constructing a schedule. Civil-time inputs and resolution policies need a clear API. | 1–2 weeks |
| Broader ISO-8601 parsing | Consumers need week dates or other defined forms beyond the current recurrence parser and Go's existing layouts. | 1–3 weeks |
| Standalone Easter calculations | Consumers need public Western, Orthodox, or Julian calculations beyond the existing internal Western BYEASTER support. | 3–7 days |
| VTIMEZONE/custom timezone support | Real calendar imports require timezone definitions embedded in files rather than resolvable IANA names. This needs a substantial design around Go's timezone representation. | 6–12+ weeks |
| General or fuzzy date parsing | A clear product need justifies a broad, configurable parser and its ambiguity rules. This remains deferred. | 8–16+ weeks |

For context, see dateutil's [timezone utilities](https://dateutil.readthedocs.io/en/stable/tz.html), [parsers](https://dateutil.readthedocs.io/en/stable/parser.html), and [Easter algorithms](https://dateutil.readthedocs.io/en/stable/easter.html).

Go already provides location loading, fixed offsets, and optional [embedded timezone data](https://pkg.go.dev/time/tzdata). Recreating those facilities, or reaching full dateutil parity, is not a project goal.

## Milestones and release sequencing

| Milestone | Intended role |
| --- | --- |
| [1.8.4](https://github.com/kentonium3/rrule-go/milestone/1) | Selected backward-compatible maintenance fixes. A patch release is not a bucket for every planned behavior or API change. |
| [1.9.0](https://github.com/kentonium3/rrule-go/milestone/2) | The current recurrence-correctness release, including documented time-form and compatibility changes. Priority 2 work should not hold up its release. |
| [2.0.0](https://github.com/kentonium3/rrule-go/milestone/3) | A tentative planning bucket for substantial API evolution and selected extensions. Its scope will be refined through feedback and design; it does not promise all exploratory items. |

Priority 2 additions can follow in appropriate minor releases. Further 1.x patch and minor milestones will be created when their scope is known. Additive functions do not automatically require a major version; a major release should communicate a deliberate compatibility boundary. See [Go module versioning guidance](https://go.dev/doc/modules/version-numbers).

The estimates above are rough effort ranges for one experienced maintainer, including design, implementation, tests, documentation, and review. They are not due dates, do not imply staffing, and should not be added into a release promise. The later items have substantially more uncertainty and will be re-estimated after discovery.

## Evidence and feedback

Behavior changes need focused regressions and a clear RFC rationale. Changes to recurrence generation also need compatibility comparisons, race testing, fuzzing, and review independent of the implementer. Performance features need reproducible benchmarks as well as correctness tests. Python behavior is a comparison input, not a substitute for the RFC or independent expected values.

Feedback is most useful when it includes a concrete schedule or input, the desired output, the application's constraints, and the current workaround. For composition, show the component schedules and intended union or exclusions. For caching, describe query frequency, schedule age, latency, and memory constraints. For broader utilities, show why existing Go facilities do not meet the need.

Open a [project issue](https://github.com/kentonium3/rrule-go/issues) to contribute an example or discuss a candidate. Confirmed work, exploratory proposals, and reported defects will remain distinguishable as the roadmap evolves.
