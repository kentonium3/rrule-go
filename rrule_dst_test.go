package rrule

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"
)

func dstLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}
func dstUTC(s string) time.Time {
	v, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		panic(err)
	}
	return v
}
func dstCheck(t *testing.T, got []time.Time, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v (%d), want %v (%d)", got, len(got), want, len(want))
	}
	for i, s := range want {
		if !got[i].Equal(dstUTC(s)) {
			t.Errorf("[%d] got %s (%s), want %s", i, got[i], got[i].UTC(), s)
		}
		if i > 0 && !got[i].After(got[i-1]) {
			t.Errorf("nonascending instances: %v", got)
		}
	}
}
func dstRule(t *testing.T, o ROption) *RRule {
	t.Helper()
	r, err := NewRRule(o)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Sydney scenarios retain the useful transition fixtures from teambition PR66.
// Expected instants are re-derived from RFC5545 sections3.3.5/3.3.10:
// ignore generated gaps, and use the first offset for ambiguous local values.
func TestDSTSydney(t *testing.T) {
	loc := dstLocation(t, "Australia/Sydney")
	cases := []struct {
		name string
		opt  ROption
		want []string
	}{
		{"HourlySpring", ROption{Freq: HOURLY, Dtstart: time.Date(2022, 10, 2, 0, 0, 0, 0, loc), Count: 6}, []string{"2022-10-01T14:00:00Z", "2022-10-01T15:00:00Z", "2022-10-01T16:00:00Z", "2022-10-01T17:00:00Z", "2022-10-01T18:00:00Z", "2022-10-01T19:00:00Z"}},
		{"HourlyFold", ROption{Freq: HOURLY, Dtstart: time.Date(2023, 4, 2, 0, 0, 0, 0, loc), Count: 6}, []string{"2023-04-01T13:00:00Z", "2023-04-01T14:00:00Z", "2023-04-01T15:00:00Z", "2023-04-01T17:00:00Z", "2023-04-01T18:00:00Z", "2023-04-01T19:00:00Z"}},
		{"MinutelySpring", ROption{Freq: MINUTELY, Dtstart: time.Date(2022, 10, 2, 0, 0, 0, 0, loc), Byminute: []int{0}, Count: 4}, []string{"2022-10-01T14:00:00Z", "2022-10-01T15:00:00Z", "2022-10-01T16:00:00Z", "2022-10-01T17:00:00Z"}},
		{"DailyExpansionCompatibility", ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 1, 0, 0, 0, 0, loc), Byhour: []int{2, 3}, Count: 4}, []string{"2022-09-30T16:00:00Z", "2022-09-30T17:00:00Z", "2022-10-01T16:00:00Z", "2022-10-02T15:00:00Z"}},
		{"DailyExpansionAligned", ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 1, 2, 0, 0, 0, loc), Byhour: []int{2, 3}, Count: 4}, []string{"2022-09-30T16:00:00Z", "2022-09-30T17:00:00Z", "2022-10-01T16:00:00Z", "2022-10-02T15:00:00Z"}},
		{"DailyGap", ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 1, 2, 30, 0, 0, loc), Count: 3}, []string{"2022-09-30T16:30:00Z", "2022-10-02T15:30:00Z", "2022-10-03T15:30:00Z"}},
		{"DailyFold", ROption{Freq: DAILY, Dtstart: time.Date(2023, 4, 1, 2, 30, 0, 0, loc), Count: 3}, []string{"2023-03-31T15:30:00Z", "2023-04-01T15:30:00Z", "2023-04-02T16:30:00Z"}},
		{"BYHOUR9WallClockControl", ROption{Freq: HOURLY, Dtstart: time.Date(2023, 4, 1, 0, 0, 0, 0, loc), Byhour: []int{9}, Count: 2}, []string{"2023-03-31T22:00:00Z", "2023-04-01T23:00:00Z"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { dstCheck(t, dstRule(t, c.opt).All(), c.want...) })
	}
}

func TestDSTAllFrequencies(t *testing.T) {
	for f := YEARLY; f <= SECONDLY; f++ {
		t.Run(f.String()+"Gap", func(t *testing.T) {
			loc := dstLocation(t, "Australia/Sydney")
			o := ROption{Freq: f, Dtstart: time.Date(2022, 10, 1, 2, 30, 0, 0, loc), Count: 3}
			want := []string{"2022-09-30T16:30:00Z", "2022-10-02T15:30:00Z", "2022-10-03T15:30:00Z"}
			switch f {
			case YEARLY:
				o.Bymonth = []int{10}
				o.Bymonthday = []int{1, 2, 3}
				want[2] = "2023-10-01T15:30:00Z"
			case MONTHLY:
				o.Bymonthday = []int{1, 2, 3}
				want[2] = "2022-10-31T15:30:00Z"
			case WEEKLY:
				o.Byweekday = []Weekday{SA, SU, MO}
				want[2] = "2022-10-07T15:30:00Z"
			case HOURLY, MINUTELY, SECONDLY:
				o.Dtstart = time.Date(2022, 10, 2, 1, 30, 0, 0, loc)
				if f == MINUTELY {
					o.Interval = 60
				}
				if f == SECONDLY {
					o.Interval = 3600
				}
				want = []string{"2022-10-01T15:30:00Z", "2022-10-01T16:30:00Z", "2022-10-01T17:30:00Z"}
			}
			dstCheck(t, dstRule(t, o).All(), want...)
		})
		t.Run(f.String()+"Fold", func(t *testing.T) {
			loc := dstLocation(t, "Australia/Sydney")
			o := ROption{Freq: f, Dtstart: time.Date(2023, 4, 1, 2, 30, 0, 0, loc), Count: 3}
			want := []string{"2023-03-31T15:30:00Z", "2023-04-01T15:30:00Z", "2023-04-02T16:30:00Z"}
			switch f {
			case YEARLY:
				o.Bymonth = []int{4}
				o.Bymonthday = []int{1, 2, 3}
			case MONTHLY:
				o.Bymonthday = []int{1, 2, 3}
			case WEEKLY:
				o.Byweekday = []Weekday{SA, SU, MO}
			case HOURLY, MINUTELY, SECONDLY:
				o.Dtstart = time.Date(2023, 4, 2, 1, 30, 0, 0, loc)
				if f == MINUTELY {
					o.Interval = 60
				}
				if f == SECONDLY {
					o.Interval = 3600
				}
				want = []string{"2023-04-01T14:30:00Z", "2023-04-01T15:30:00Z", "2023-04-01T17:30:00Z"}
			}
			dstCheck(t, dstRule(t, o).All(), want...)
		})
	}
}

func TestDSTOtherZones(t *testing.T) {
	cases := []struct {
		name, zone string
		start      time.Time
		freq       Frequency
		interval   int
		want       []string
	}{
		{"NewYorkGap", "America/New_York", dstUTC("2023-03-12T06:30:00Z"), HOURLY, 1, []string{"2023-03-12T06:30:00Z", "2023-03-12T07:30:00Z", "2023-03-12T08:30:00Z"}},
		{"NewYorkFold", "America/New_York", dstUTC("2023-11-05T04:30:00Z"), HOURLY, 1, []string{"2023-11-05T04:30:00Z", "2023-11-05T05:30:00Z", "2023-11-05T07:30:00Z"}},
		{"LordHoweGap", "Australia/Lord_Howe", dstUTC("2023-09-30T15:00:00Z"), MINUTELY, 15, []string{"2023-09-30T15:00:00Z", "2023-09-30T15:15:00Z", "2023-09-30T15:30:00Z"}},
		{"LordHoweFold", "Australia/Lord_Howe", dstUTC("2023-04-01T14:00:00Z"), MINUTELY, 15, []string{"2023-04-01T14:00:00Z", "2023-04-01T14:15:00Z", "2023-04-01T14:30:00Z"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := ROption{Freq: c.freq, Dtstart: c.start.In(dstLocation(t, c.zone)), Interval: c.interval, Count: 3}
			dstCheck(t, dstRule(t, o).All(), c.want...)
		})
	}
}

func TestDSTBysetposRanks(t *testing.T) {
	loc := dstLocation(t, "Australia/Sydney")
	for _, pos := range []int{3, -3} {
		t.Run(fmt.Sprint(pos), func(t *testing.T) {
			o := ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 1, 1, 0, 0, 0, loc), Byhour: []int{1, 2, 3, 4}, Bysetpos: []int{pos}, Count: 3}
			want := []string{"2022-09-30T17:00:00Z", "2022-10-01T17:00:00Z", "2022-10-02T16:00:00Z"}
			if pos < 0 {
				want = []string{"2022-09-30T16:00:00Z", "2022-10-01T15:00:00Z", "2022-10-02T15:00:00Z"}
			}
			dstCheck(t, dstRule(t, o).All(), want...)
		})
	}
	for _, pos := range []int{1, 2, 3, -1, -2, -3} {
		t.Run("HalfHour"+fmt.Sprint(pos), func(t *testing.T) {
			o := ROption{Freq: HOURLY, Dtstart: time.Date(2023, 10, 1, 1, 0, 0, 0, dstLocation(t, "Australia/Lord_Howe")), Byminute: []int{0, 15, 30, 45}, Bysetpos: []int{pos}, Count: 2}
			minute := 0
			if pos == 2 || pos == -1 {
				minute = 15
			}
			first := fmt.Sprintf("2023-09-30T14:%02d:00Z", 30+minute)
			second := "2023-09-30T15:30:00Z"
			if pos == 2 || pos == -1 {
				second = "2023-09-30T15:45:00Z"
			}
			if pos == 3 {
				first = "2023-09-30T15:00:00Z"
				second = "2023-09-30T16:30:00Z"
			}
			if pos == -3 {
				first = "2023-09-30T14:45:00Z"
				second = "2023-09-30T16:15:00Z"
			}
			if pos == -1 {
				first = "2023-09-30T15:15:00Z"
			}
			if pos == -2 {
				first = "2023-09-30T15:00:00Z"
			}
			dstCheck(t, dstRule(t, o).All(), first, second)
		})
	}
	t.Run("SignedOverlapDedup", func(t *testing.T) {
		o := ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 2, 1, 0, 0, 0, loc), Byhour: []int{1, 2, 3, 4}, Bysetpos: []int{3, -1}, Count: 3}
		dstCheck(t, dstRule(t, o).All(), "2022-10-01T17:00:00Z", "2022-10-02T16:00:00Z", "2022-10-02T17:00:00Z")
	})
}

func TestDSTPublicQueries(t *testing.T) {
	r := dstRule(t, ROption{Freq: DAILY, Dtstart: time.Date(2022, 10, 1, 2, 30, 0, 0, dstLocation(t, "Australia/Sydney")), Count: 3})
	want := []string{"2022-09-30T16:30:00Z", "2022-10-02T15:30:00Z", "2022-10-03T15:30:00Z"}
	t.Run("Iterator", func(t *testing.T) {
		next := r.Iterator()
		var got []time.Time
		for i := 0; i < 4; i++ {
			v, ok := next()
			if !ok {
				break
			}
			got = append(got, v)
		}
		dstCheck(t, got, want...)
	})
	t.Run("Between", func(t *testing.T) { dstCheck(t, r.Between(dstUTC(want[0]), dstUTC(want[2]), true), want...) })
	t.Run("After", func(t *testing.T) { dstCheck(t, []time.Time{r.After(dstUTC(want[0]), false)}, want[1]) })
	t.Run("Before", func(t *testing.T) { dstCheck(t, []time.Time{r.Before(dstUTC(want[1]), false)}, want[0]) })
	t.Run("Set", func(t *testing.T) { s := Set{}; s.RRule(r); dstCheck(t, s.All(), want...) })
	for _, delta := range []int{-1, 0, 1} {
		t.Run(fmt.Sprint("Until", delta), func(t *testing.T) {
			o := r.OrigOptions
			o.Count = 0
			o.Until = dstUTC(want[1]).Add(time.Duration(delta) * time.Second)
			w := want[:2]
			if delta < 0 {
				w = want[:1]
			}
			dstCheck(t, dstRule(t, o).All(), w...)
		})
	}
}

func TestDSTParsedAnchors(t *testing.T) {
	cases := []struct {
		name, input string
		want        []string
	}{
		{"GapNominalCadence", "DTSTART;TZID=Australia/Sydney:20221002T023000\nRRULE:FREQ=DAILY;COUNT=3", []string{"2022-10-01T16:30:00Z", "2022-10-02T15:30:00Z", "2022-10-03T15:30:00Z"}},
		{"NewYorkPregapOffset", "DTSTART;TZID=America/New_York:20230312T023000\nRRULE:FREQ=DAILY;COUNT=2", []string{"2023-03-12T07:30:00Z", "2023-03-13T06:30:00Z"}},
		{"FoldFirst", "DTSTART;TZID=Australia/Sydney:20230402T023000\nRRULE:FREQ=DAILY;COUNT=2", []string{"2023-04-01T15:30:00Z", "2023-04-02T16:30:00Z"}},
		{"NewYorkGapSubdailyOverlap", "DTSTART;TZID=America/New_York:20230312T023000\nRRULE:FREQ=MINUTELY;INTERVAL=15;COUNT=4", []string{"2023-03-12T07:30:00Z", "2023-03-12T07:45:00Z", "2023-03-12T08:00:00Z", "2023-03-12T08:15:00Z"}},
		{"GapSubdailyOverlap", "DTSTART;TZID=Australia/Sydney:20221002T023000\nRRULE:FREQ=MINUTELY;INTERVAL=15;COUNT=4", []string{"2022-10-01T16:30:00Z", "2022-10-01T16:45:00Z", "2022-10-01T17:00:00Z", "2022-10-01T17:15:00Z"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := StrToRRule(c.input)
			if err != nil {
				t.Fatal(err)
			}
			dstCheck(t, r.All(), c.want...)
		})
	}
	t.Run("StandaloneGap", func(t *testing.T) {
		v, e := StrToDtStart("TZID=America/New_York:20230312T023000", time.UTC)
		if e != nil {
			t.Fatal(e)
		}
		dstCheck(t, []time.Time{v}, "2023-03-12T07:30:00Z")
	})
	t.Run("StandaloneFold", func(t *testing.T) {
		v, e := StrToDtStart("TZID=Australia/Sydney:20230402T023000", time.UTC)
		if e != nil {
			t.Fatal(e)
		}
		dstCheck(t, []time.Time{v}, "2023-04-01T15:30:00Z")
	})
}

func TestDSTAuthoritativeLaterFold(t *testing.T) {
	for _, c := range []struct{ zone, instant, next string }{{"Australia/Sydney", "2023-04-01T16:30:00Z", "2023-04-02T16:30:00Z"}, {"America/New_York", "2023-11-05T06:30:00Z", "2023-11-06T06:30:00Z"}, {"Australia/Lord_Howe", "2023-04-01T15:15:00Z", "2023-04-02T15:15:00Z"}} {
		t.Run(c.zone, func(t *testing.T) {
			v := dstUTC(c.instant).In(dstLocation(t, c.zone))
			o := ROption{Freq: DAILY, Dtstart: v, Count: 2}
			r := dstRule(t, o)
			dstCheck(t, r.All(), c.instant, c.next)
			r.DTStart(v)
			dstCheck(t, r.All(), c.instant, c.next)
		})
	}
}

func TestDSTMetadataCopiesAndLocation(t *testing.T) {
	o, err := StrToROption("DTSTART;TZID=Australia/Sydney:20221002T023000\nRRULE:FREQ=DAILY;COUNT=2")
	if err != nil {
		t.Fatal(err)
	}
	copy := *o
	t.Run("CopyRetainsNominal", func(t *testing.T) {
		dstCheck(t, dstRule(t, copy).All(), "2022-10-01T16:30:00Z", "2022-10-02T15:30:00Z")
	})
	t.Run("ChangedStartInvalidates", func(t *testing.T) {
		copy := copy
		copy.Dtstart = copy.Dtstart.Add(time.Hour)
		dstCheck(t, dstRule(t, copy).All(), "2022-10-01T17:30:00Z", "2022-10-02T17:30:00Z")
	})
	t.Run("LocationIdentityInvalidates", func(t *testing.T) {
		copy := copy
		loc := time.FixedZone("Australia/Sydney", 10*3600)
		copy.Dtstart = copy.Dtstart.In(loc)
		r := dstRule(t, copy)
		if r.GetDTStart().Location() != loc {
			t.Fatal("lost exact location")
		}
		dstCheck(t, r.All(), "2022-10-01T16:30:00Z", "2022-10-02T16:30:00Z")
	})
	t.Run("UntilReplacementRetainsStart", func(t *testing.T) {
		copy := copy
		copy.Count = 0
		r := dstRule(t, copy)
		r.Until(dstUTC("2022-10-03T00:00:00Z"))
		dstCheck(t, r.All(), "2022-10-01T16:30:00Z", "2022-10-02T15:30:00Z")
	})
	t.Run("PublicSetterReplacesNominal", func(t *testing.T) {
		r := dstRule(t, copy)
		r.DTStart(copy.Dtstart)
		dstCheck(t, r.All(), "2022-10-01T16:30:00Z", "2022-10-02T16:30:00Z")
	})
	t.Run("FixedZoneNameIsNotIANA", func(t *testing.T) {
		loc := time.FixedZone("America/New_York", -5*3600)
		op, e := StrToROptionInLocation("DTSTART:20230312T013000\nRRULE:FREQ=HOURLY;COUNT=3", loc)
		if e != nil {
			t.Fatal(e)
		}
		r := dstRule(t, *op)
		if op.Dtstart.Location() != loc || r.GetDTStart().Location() != loc {
			t.Fatal("lost exact location")
		}
		dstCheck(t, r.All(), "2023-03-12T06:30:00Z", "2023-03-12T07:30:00Z", "2023-03-12T08:30:00Z")
	})
}

func BenchmarkDSTAnnualBysetpos(b *testing.B) {
	values := func(n int) []int {
		v := make([]int, n)
		for i := range v {
			v[i] = i
		}
		return v
	}
	o := ROption{Freq: YEARLY, Dtstart: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Byhour: values(24), Byminute: values(60), Bysecond: values(60), Bysetpos: []int{1, -1}, Byweekday: []Weekday{MO, TU, WE, TH, FR, SA, SU}, Count: 2}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := NewRRule(o)
		r.All()
	}
}

// Each frequency reaches both transitions in every supported DST fixture zone.
// Coarse frequencies explicitly select adjacent dates, so a 16-value sample
// cannot accidentally stay away from the transition.
func TestDSTEveryFrequencyEveryZone(t *testing.T) {
	cases := []struct {
		name, zone  string
		y           int
		m           time.Month
		d, h, min   int
		fold        bool
		coarse, sub []string
	}{
		{"NYGap", "America/New_York", 2023, 3, 12, 2, 15, false, []string{"2023-03-11T07:15:00Z", "2023-03-13T06:15:00Z"}, []string{"2023-03-12T06:15:00Z", "2023-03-12T07:15:00Z"}},
		{"NYFold", "America/New_York", 2023, 11, 5, 1, 15, true, []string{"2023-11-04T05:15:00Z", "2023-11-05T05:15:00Z"}, []string{"2023-11-05T04:15:00Z", "2023-11-05T05:15:00Z"}},
		{"SydneyGap", "Australia/Sydney", 2022, 10, 2, 2, 15, false, []string{"2022-09-30T16:15:00Z", "2022-10-02T15:15:00Z"}, []string{"2022-10-01T15:15:00Z", "2022-10-01T16:15:00Z"}},
		{"SydneyFold", "Australia/Sydney", 2023, 4, 2, 2, 15, true, []string{"2023-03-31T15:15:00Z", "2023-04-01T15:15:00Z"}, []string{"2023-04-01T14:15:00Z", "2023-04-01T15:15:00Z"}},
		{"LordHoweGap", "Australia/Lord_Howe", 2023, 10, 1, 2, 15, false, []string{"2023-09-29T15:45:00Z", "2023-10-01T15:15:00Z"}, []string{"2023-09-30T14:45:00Z", "2023-09-30T16:15:00Z"}},
		{"LordHoweFold", "Australia/Lord_Howe", 2023, 4, 2, 1, 45, true, []string{"2023-03-31T14:45:00Z", "2023-04-01T14:45:00Z"}, []string{"2023-04-01T13:45:00Z", "2023-04-01T14:45:00Z"}},
	}
	for _, c := range cases {
		for f := YEARLY; f <= SECONDLY; f++ {
			t.Run(c.name+"/"+f.String(), func(t *testing.T) {
				loc := dstLocation(t, c.zone)
				prev := time.Date(c.y, c.m, c.d-1, c.h, c.min, 0, 0, loc)
				o := ROption{Freq: f, Dtstart: prev, Count: 2}
				want := c.coarse
				switch f {
				case YEARLY, MONTHLY:
					// Lord Howe's October1 fixture begins in September. Include both
					// months so its pre-transition positive control remains selected.
					o.Bymonthday = []int{prev.Day(), c.d, c.d + 1}
					if f == YEARLY {
						o.Bymonth = []int{int(prev.Month()), int(c.m)}
					}
				case WEEKLY:
					for i := -1; i <= 1; i++ {
						d := time.Date(c.y, c.m, c.d+i, 0, 0, 0, 0, time.UTC)
						o.Byweekday = append(o.Byweekday, Weekday{weekday: toPyWeekday(d.Weekday())})
					}
				case HOURLY, MINUTELY, SECONDLY:
					o.Dtstart = time.Date(c.y, c.m, c.d, c.h-1, c.min, 0, 0, loc)
					if f == MINUTELY {
						o.Interval = 60
					}
					if f == SECONDLY {
						o.Interval = 3600
					}
					want = c.sub
				}
				dstCheck(t, dstRule(t, o).All(), want...)
			})
		}
	}
}

func TestDSTCustomTZifIdentity(t *testing.T) {
	// A valid fixed-offset TZif with an IANA-colliding name. Its behavior and
	// pointer must survive parsing, option copies, construction, Set and setters.
	data := make([]byte, 54)
	copy(data, "TZif")
	binary.BigEndian.PutUint32(data[36:40], 1)
	binary.BigEndian.PutUint32(data[40:44], 4)
	binary.BigEndian.PutUint32(data[44:48], 2*3600)
	copy(data[50:], "XTZ\x00")
	loc, e := time.LoadLocationFromTZData("America/New_York", data)
	if e != nil {
		t.Fatal(e)
	}
	o, e := StrToROptionInLocation("DTSTART:20230312T023000\nRRULE:FREQ=DAILY;COUNT=2", loc)
	if e != nil {
		t.Fatal(e)
	}
	copy := *o
	r := dstRule(t, copy)
	set := Set{}
	set.RRule(r)
	if o.Dtstart.Location() != loc || r.GetDTStart().Location() != loc || set.GetDTStart().Location() != loc {
		t.Fatal("replaced custom TZif by its name")
	}
	dstCheck(t, set.All(), "2023-03-12T00:30:00Z", "2023-03-13T00:30:00Z")
	r.DTStart(copy.Dtstart)
	if r.GetDTStart().Location() != loc {
		t.Fatal("setter lost location")
	}
	dstCheck(t, r.All(), "2023-03-12T00:30:00Z", "2023-03-13T00:30:00Z")
}

func TestDSTAnnualTransitionRanks(t *testing.T) {
	for _, f := range []Frequency{YEARLY, MONTHLY} {
		for _, pos := range []int{7, -7} {
			t.Run(fmt.Sprint(f, "/", pos), func(t *testing.T) {
				o := ROption{Freq: f, Dtstart: time.Date(2022, 10, 1, 1, 0, 0, 0, dstLocation(t, "Australia/Sydney")), Bymonth: []int{10}, Bymonthday: []int{1, 2, 3}, Byhour: []int{1, 2, 3, 4}, Bysetpos: []int{pos}, Count: 1}
				want := "2022-10-01T17:00:00Z"
				if pos < 0 {
					want = "2022-10-01T15:00:00Z"
				}
				dstCheck(t, dstRule(t, o).All(), want)
			})
		}
	}
}

func TestDSTUTCShortUntilFloatingControl(t *testing.T) {
	// Default floating parsing evaluates numerically in UTC. Keep the existing
	// short UNTIL termination bound even after attaching private start metadata.
	r, e := StrToRRule("DTSTART:20260105T100000\nRRULE:FREQ=DAILY;BYMONTH=2;BYMONTHDAY=30;UNTIL=20260201T000000")
	if e != nil {
		t.Fatal(e)
	}
	if got := r.All(); len(got) != 0 {
		t.Fatal(got)
	}
}

func TestDSTDenseAnnualArithmetic(t *testing.T) {
	values := func(n int) []int {
		v := make([]int, n)
		for i := range v {
			v[i] = i
		}
		return v
	}
	// All weekdays suppress the YEARLY default month/day restriction. This
	// period contains365*86400=31,536,000 nominal candidates, selected by rank.
	o := ROption{Freq: YEARLY, Dtstart: dstUTC("2023-01-01T00:00:00Z"), Byweekday: []Weekday{MO, TU, WE, TH, FR, SA, SU}, Byhour: values(24), Byminute: values(60), Bysecond: values(60), Bysetpos: []int{1, 366, -366, -1}, Count: 4}
	dstCheck(t, dstRule(t, o).All(), "2023-01-01T00:00:00Z", "2023-01-01T00:06:05Z", "2023-12-31T23:53:54Z", "2023-12-31T23:59:59Z")
}

func TestDSTParsedSetAnchors(t *testing.T) {
	for _, c := range []struct {
		name, input string
		want        []string
	}{
		{"SydneyGap", "DTSTART;TZID=Australia/Sydney:20221002T023000\nRRULE:FREQ=DAILY;COUNT=2", []string{"2022-10-01T16:30:00Z", "2022-10-02T15:30:00Z"}},
		{"NYGap", "DTSTART;TZID=America/New_York:20230312T023000\nRRULE:FREQ=DAILY;COUNT=2", []string{"2023-03-12T07:30:00Z", "2023-03-13T06:30:00Z"}},
		{"SydneyFold", "DTSTART;TZID=Australia/Sydney:20230402T023000\nRRULE:FREQ=DAILY;COUNT=2", []string{"2023-04-01T15:30:00Z", "2023-04-02T16:30:00Z"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			set, e := StrToRRuleSet(c.input)
			if e != nil {
				t.Fatal(e)
			}
			dstCheck(t, set.All(), c.want...)
		})
	}
}

func TestDSTGapAnchorBysetposCollision(t *testing.T) {
	cases := []struct {
		name, positions string
		want            []string
	}{
		{"PositiveCollision", "1,2", []string{"2007-03-11T07:30:00Z", "2007-03-11T08:30:00Z", "2007-03-12T06:30:00Z", "2007-03-12T07:30:00Z"}},
		{"PositiveOutOfRange", "1,3", []string{"2007-03-11T07:30:00Z", "2007-03-12T06:30:00Z", "2007-03-12T08:30:00Z", "2007-03-13T06:30:00Z"}},
		{"SignedControl", "1,-1", []string{"2007-03-11T07:30:00Z", "2007-03-11T08:30:00Z", "2007-03-12T06:30:00Z", "2007-03-12T08:30:00Z"}},
		{"NegativeControl", "-2,-1", []string{"2007-03-11T07:30:00Z", "2007-03-11T08:30:00Z", "2007-03-12T07:30:00Z", "2007-03-12T08:30:00Z"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			input := "DTSTART;TZID=America/New_York:20070311T023000\nRRULE:FREQ=DAILY;COUNT=4;BYHOUR=2,3,4;BYMINUTE=30;BYSETPOS=" + c.positions
			r, e := StrToRRule(input)
			if e != nil {
				t.Fatal(e)
			}
			t.Run("All", func(t *testing.T) { dstCheck(t, r.All(), c.want...) })
			t.Run("Iterator", func(t *testing.T) {
				next := r.Iterator()
				var got []time.Time
				for i := 0; i < 5; i++ {
					v, ok := next()
					if !ok {
						break
					}
					got = append(got, v)
				}
				dstCheck(t, got, c.want...)
			})
			t.Run("Set", func(t *testing.T) { set := Set{}; set.RRule(r); dstCheck(t, set.All(), c.want...) })
			t.Run("Between", func(t *testing.T) { dstCheck(t, r.Between(dstUTC(c.want[0]), dstUTC(c.want[3]), true), c.want...) })
			t.Run("BetweenExclusive", func(t *testing.T) {
				dstCheck(t, r.Between(dstUTC(c.want[0]), dstUTC(c.want[3]), false), c.want[1:3]...)
			})
			t.Run("After", func(t *testing.T) { dstCheck(t, []time.Time{r.After(dstUTC(c.want[0]), false)}, c.want[1]) })
			t.Run("Before", func(t *testing.T) { dstCheck(t, []time.Time{r.Before(dstUTC(c.want[2]), false)}, c.want[1]) })
			for _, delta := range []int{-1, 0, 1} {
				t.Run(fmt.Sprint("Until", delta), func(t *testing.T) {
					options := r.OrigOptions
					options.Count = 0
					options.Until = dstUTC(c.want[1]).Add(time.Duration(delta) * time.Second)
					want := c.want[:2]
					if delta < 0 {
						want = c.want[:1]
					}
					dstCheck(t, dstRule(t, options).All(), want...)
				})
			}
		})
	}
}

func TestDSTGapAnchorRankPeriods(t *testing.T) {
	for _, f := range []Frequency{YEARLY, MONTHLY, WEEKLY} {
		t.Run(f.String(), func(t *testing.T) {
			r, e := StrToRRule("DTSTART;TZID=America/New_York:20070311T023000\nRRULE:FREQ=" + f.String() + ";COUNT=2;BYHOUR=2,3,4;BYMINUTE=30;BYSETPOS=1,2")
			if e != nil {
				t.Fatal(e)
			}
			dstCheck(t, r.All(), "2007-03-11T07:30:00Z", "2007-03-11T08:30:00Z")
		})
	}
	t.Run("HalfHour", func(t *testing.T) {
		r, e := StrToRRule("DTSTART;TZID=Australia/Lord_Howe:20231001T021500\nRRULE:FREQ=DAILY;COUNT=4;BYHOUR=2,3;BYMINUTE=15,45;BYSETPOS=1,2")
		if e != nil {
			t.Fatal(e)
		}
		dstCheck(t, r.All(), "2023-09-30T15:45:00Z", "2023-09-30T16:15:00Z", "2023-10-01T15:15:00Z", "2023-10-01T15:45:00Z")
	})
}

func TestDSTGapAnchorCrossDayRanks(t *testing.T) {
	for _, c := range []struct{ freq, dates, firstNext, thirdNext string }{
		{"YEARLY", "BYMONTH=12;BYMONTHDAY=30,31", "2012-12-29T19:00:00Z", "2012-12-30T19:00:00Z"},
		{"MONTHLY", "BYMONTHDAY=30,31", "2012-01-29T19:00:00Z", "2012-01-30T19:00:00Z"},
		{"WEEKLY", "BYDAY=FR,SA", "2012-01-05T19:00:00Z", "2012-01-06T19:00:00Z"},
	} {
		for _, pos := range []string{"1,2", "1,3"} {
			t.Run(c.freq+"/"+pos, func(t *testing.T) {
				// Apia's skipped December30 resolves this explicit start to December31.
				// A coarse period can contain its nominal date and normal resolved date.
				input := "DTSTART;TZID=Pacific/Apia:20111230T090000\nRRULE:FREQ=" + c.freq + ";COUNT=3;" + c.dates + ";BYHOUR=9,10;BYSETPOS=" + pos
				r, e := StrToRRule(input)
				if e != nil {
					t.Fatal(e)
				}
				want := []string{"2011-12-30T19:00:00Z", "2011-12-30T20:00:00Z", c.firstNext}
				if pos == "1,3" {
					want = []string{"2011-12-30T19:00:00Z", c.firstNext, c.thirdNext}
				}
				dstCheck(t, r.All(), want...)
			})
		}
	}
}

func TestDSTGapAnchorPeriodControls(t *testing.T) {
	t.Run("LaterDailyPeriodKeepsNormalRanks", func(t *testing.T) {
		r, e := StrToRRule("DTSTART;TZID=Pacific/Apia:20111230T090000\nRRULE:FREQ=DAILY;COUNT=3;BYHOUR=9,10,11;BYSETPOS=1,2")
		if e != nil {
			t.Fatal(e)
		}
		dstCheck(t, r.All(), "2011-12-30T19:00:00Z", "2011-12-30T20:00:00Z", "2011-12-31T19:00:00Z")
	})
	t.Run("FilteredResolvedDayRetainsAnchor", func(t *testing.T) {
		r, e := StrToRRule("DTSTART;TZID=Pacific/Apia:20111230T090000\nRRULE:FREQ=MONTHLY;COUNT=3;BYMONTHDAY=30;BYHOUR=9,10;BYSETPOS=1,2")
		if e != nil {
			t.Fatal(e)
		}
		dstCheck(t, r.All(), "2011-12-30T19:00:00Z", "2012-01-29T19:00:00Z", "2012-01-29T20:00:00Z")
	})
}
