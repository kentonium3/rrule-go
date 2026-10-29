package rrule

import (
	"sort"
	"time"
)

// civilFields is a wall-clock tuple, independent of Go's gap normalization.
type civilFields struct {
	year                      int
	month                     time.Month
	day, hour, minute, second int
}

func civilFromTime(t time.Time) civilFields {
	y, m, d := t.Date()
	h, n, s := t.Clock()
	return civilFields{y, m, d, h, n, s}
}
func (c civilFields) utc() time.Time {
	return time.Date(c.year, c.month, c.day, c.hour, c.minute, c.second, 0, time.UTC)
}
func (c civilFields) clock() int { return c.hour*3600 + c.minute*60 + c.second }

// Parser descriptors are immutable. ROption copies can share them; replacing a
// part allocates a new container and never changes another copy's descriptor.
type timeForm uint8

const (
	timeFormDate timeForm = iota
	timeFormFloating
	timeFormUTC
	timeFormTZID
)

type parsedTime struct {
	form     timeForm
	nominal  civilFields
	location *time.Location
	snapshot time.Time
}
type parseMetadata struct{ dtstart, until *parsedTime }

func (p *parsedTime) valid(t time.Time) bool {
	return p != nil && t.Location() == p.location && t.Equal(p.snapshot)
}
func (o ROption) parsedStart() *parsedTime {
	if o.parse != nil && o.parse.dtstart.valid(o.Dtstart) {
		return o.parse.dtstart
	}
	return nil
}
func (o ROption) parsedUntil() *parsedTime {
	if o.parse != nil && o.parse.until.valid(o.Until) {
		return o.parse.until
	}
	return nil
}
func (o *ROption) setParsedStart(p *parsedTime) {
	until := o.parsedUntil()
	if p == nil && until == nil {
		o.parse = nil
		return
	}
	o.parse = &parseMetadata{dtstart: p, until: until}
}
func (o *ROption) setParsedUntil(p *parsedTime) {
	start := o.parsedStart()
	if start == nil && p == nil {
		o.parse = nil
		return
	}
	o.parse = &parseMetadata{dtstart: start, until: p}
}

// Modern DST transitions are separated by more than 72 hours, their offsets
// are within 24 hours of UTC, and their jumps are shorter than a day. Sampling
// the ends of this 72-hour window finds both offsets for any transition in the
// requested civil day. A binary search then locates its exact second. This is
// deliberately bounded, not a claim to handle every historical discontinuity.
// The profile belongs to an iterator/day, never a global timezone cache.
type civilDay struct {
	year             int
	month            time.Month
	day              int
	location         *time.Location
	midnight         int64
	before, after    int
	gapStart, gapEnd int
}

func newCivilDay(year int, month time.Month, day int, loc *time.Location) civilDay {
	midnight := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	year, month, day = midnight.Date()
	p := civilDay{year: year, month: month, day: day, location: loc, midnight: midnight.Unix()}
	if loc == time.UTC {
		return p
	}
	low, high := p.midnight-86400, p.midnight+2*86400
	_, p.before = time.Unix(low, 0).In(loc).Zone()
	_, p.after = time.Unix(high, 0).In(loc).Zone()
	if p.before == p.after {
		return p
	}
	for high-low > 1 {
		mid := low + (high-low)/2
		_, off := time.Unix(mid, 0).In(loc).Zone()
		if off == p.before {
			low = mid
		} else {
			high = mid
		}
	}
	if p.after > p.before {
		p.gapStart = int(high + int64(p.before) - p.midnight)
		p.gapEnd = int(high + int64(p.after) - p.midnight)
		if p.gapStart < 0 {
			p.gapStart = 0
		}
		if p.gapEnd > 86400 {
			p.gapEnd = 86400
		}
		if p.gapEnd < p.gapStart {
			p.gapEnd = p.gapStart
		}
	}
	return p
}
func (p civilDay) fields(clock int) civilFields {
	return civilFields{p.year, p.month, p.day, clock / 3600, clock / 60 % 60, clock % 60}
}
func (p civilDay) resolve(clock int) (time.Time, bool) {
	nominal := p.midnight + int64(clock)
	first := time.Unix(nominal-int64(p.before), 0).In(p.location)
	if p.before == p.after {
		return first, civilFromTime(first) == p.fields(clock)
	}
	second := time.Unix(nominal-int64(p.after), 0).In(p.location)
	want := p.fields(clock)
	validFirst, validSecond := civilFromTime(first) == want, civilFromTime(second) == want
	if validFirst && (!validSecond || first.Before(second)) {
		return first, true
	}
	if validSecond {
		return second, true
	}
	return time.Time{}, false
}
func (c civilFields) explicit(loc *time.Location) time.Time {
	p := newCivilDay(c.year, c.month, c.day, loc)
	if t, ok := p.resolve(c.clock()); ok {
		return t
	}
	// RFC5545 section3.3.5: an explicit gap uses the pre-gap offset.
	return time.Unix(c.utc().Unix()-int64(p.before), 0).In(loc)
}

// indexRange is a half-open missing interval in the sorted nominal timeset.
type indexRange struct{ start, end int }
type civilDayRanks struct {
	profile civilDay
	holes   []indexRange
	count   int
}

func nominalClock(t time.Time) int { h, m, s := t.Clock(); return h*3600 + m*60 + s }
func (r *RRule) dayRanks(p civilDay, timeset []time.Time, anchorAlreadyPresent bool) civilDayRanks {
	d := civilDayRanks{profile: p, count: len(timeset)}
	if p.gapStart >= p.gapEnd {
		return d
	}
	lo := sort.Search(len(timeset), func(i int) bool { return nominalClock(timeset[i]) >= p.gapStart })
	hi := sort.Search(len(timeset), func(i int) bool { return nominalClock(timeset[i]) >= p.gapEnd })
	if lo == hi {
		return d
	}
	// Retain the explicit gap anchor only if no normal slot in this period
	// already represents its resolved instant. Deduplication after positional
	// selection would have counted that instant twice before choosing ranks.
	anchor := -1
	if !anchorAlreadyPresent && r.anchor.year == p.year && r.anchor.month == p.month && r.anchor.day == p.day {
		i := sort.Search(len(timeset), func(i int) bool { return nominalClock(timeset[i]) >= r.anchor.clock() })
		if i < hi && i >= lo && nominalClock(timeset[i]) == r.anchor.clock() {
			anchor = i
		}
	}
	add := func(a, b int) {
		if a < b {
			d.holes = append(d.holes, indexRange{a, b})
			d.count -= b - a
		}
	}
	if anchor >= 0 {
		add(lo, anchor)
		add(anchor+1, hi)
	} else {
		add(lo, hi)
	}
	return d
}
func (d civilDayRanks) selectIndex(rank int) int {
	for _, hole := range d.holes {
		if rank < hole.start {
			break
		}
		rank += hole.end - hole.start
	}
	return rank
}
func (r *RRule) civilCandidate(p civilDay, clock int) (time.Time, bool) {
	if p.fields(clock) == r.anchor {
		return r.dtstart, true
	}
	return p.resolve(clock)
}
