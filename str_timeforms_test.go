package rrule

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"
)

type formParser struct {
	name  string
	parse func(string) (string, []time.Time, error)
}

func formParsers() []formParser {
	option := func(in string, loc *time.Location) (string, []time.Time, error) {
		o, e := StrToROptionInLocation(in, loc)
		if e != nil {
			return "", nil, e
		}
		r, e := NewRRule(*o)
		if e != nil {
			return "", nil, e
		}
		return o.String(), r.All(), nil
	}
	set := func(in string, mode int) (string, []time.Time, error) {
		var s *Set
		var e error
		switch mode {
		case 0:
			s, e = StrToRRuleSet(in)
		case 1:
			s, e = StrSliceToRRuleSet(strings.Split(in, "\n"))
		default:
			s, e = StrSliceToRRuleSetInLoc(strings.Split(in, "\n"), time.UTC)
		}
		if e != nil {
			return "", nil, e
		}
		return s.String(), s.All(), nil
	}
	return []formParser{
		{"RRule", func(in string) (string, []time.Time, error) {
			r, e := StrToRRule(in)
			if e != nil {
				return "", nil, e
			}
			return r.String(), r.All(), nil
		}},
		{"ROption", func(in string) (string, []time.Time, error) {
			o, e := StrToROption(in)
			if e != nil {
				return "", nil, e
			}
			r, e := NewRRule(*o)
			if e != nil {
				return "", nil, e
			}
			return o.String(), r.All(), nil
		}},
		{"ROptionInLocation", func(in string) (string, []time.Time, error) { return option(in, time.UTC) }},
		{"RRuleSet", func(in string) (string, []time.Time, error) { return set(in, 0) }},
		{"SliceSet", func(in string) (string, []time.Time, error) { return set(in, 1) }},
		{"SliceSetInLoc", func(in string) (string, []time.Time, error) { return set(in, 2) }},
	}
}
func formTimes(t *testing.T, got []time.Time, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v; want %v", got, want)
	}
	for i, v := range got {
		if v.UTC().Format(time.RFC3339) != want[i] {
			t.Fatalf("[%d] got %s want %s", i, v, want[i])
		}
	}
}
func TestTimeFormsPairMatrix(t *testing.T) {
	starts := []struct{ name, header, until string }{
		{"DATE", "DTSTART:20201228", "20201230"},
		{"floating", "DTSTART:20201228T090000", "20201230T090000"},
		{"UTC", "DTSTART:20201228T090000Z", "20201230T090000Z"},
		{"TZID", "DTSTART;TZID=Australia/Sydney:20201228T090000", "20201229T220000Z"},
	}
	untils := []struct{ name, value string }{{"DATE", "20201230"}, {"floating", "20201230T090000"}, {"UTC", "20201230T090000Z"}}
	for _, p := range formParsers() {
		for _, s := range starts {
			for _, u := range untils {
				t.Run(p.name+"/"+s.name+"/"+u.name, func(t *testing.T) {
					legal := s.name == u.name || s.name == "TZID" && u.name == "UTC"
					input := s.header + "\nRRULE:FREQ=DAILY;UNTIL=" + u.value
					_, got, e := p.parse(input)
					if !legal {
						if e == nil || !strings.Contains(e.Error(), "UNTIL") {
							t.Fatalf("expected named UNTIL rejection, got %v (%v)", e, got)
						}
						return
					}
					if e != nil {
						t.Fatal(e)
					}
					if len(got) != 3 {
						t.Fatalf("inclusive daily count: %v", got)
					}
				})
			}
		}
	}
}
func TestTimeFormsRoundTrip(t *testing.T) {
	cases := []struct {
		name, input, want string
		times             []string
	}{
		{"floating", "DTSTART:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201230T090000", "DTSTART:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201230T090000", []string{"2020-12-28T09:00:00Z", "2020-12-29T09:00:00Z", "2020-12-30T09:00:00Z"}},
		{"DATE", "DTSTART;VALUE=DATE:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230", "DTSTART;VALUE=DATE:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230", []string{"2020-12-28T00:00:00Z", "2020-12-29T00:00:00Z", "2020-12-30T00:00:00Z"}},
		{"legacyDATE", "DTSTART:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230", "DTSTART;VALUE=DATE:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230", []string{"2020-12-28T00:00:00Z", "2020-12-29T00:00:00Z", "2020-12-30T00:00:00Z"}},
		{"UTC", "DTSTART:20201228T090000Z\nRRULE:FREQ=DAILY;UNTIL=20201230T090000Z", "DTSTART:20201228T090000Z\nRRULE:FREQ=DAILY;UNTIL=20201230T090000Z", []string{"2020-12-28T09:00:00Z", "2020-12-29T09:00:00Z", "2020-12-30T09:00:00Z"}},
		{"TZID", "DTSTART;TZID=Australia/Sydney:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201229T220000Z", "DTSTART;TZID=Australia/Sydney:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201229T220000Z", []string{"2020-12-27T22:00:00Z", "2020-12-28T22:00:00Z", "2020-12-29T22:00:00Z"}},
	}
	for _, p := range formParsers() {
		for _, c := range cases {
			t.Run(p.name+"/"+c.name, func(t *testing.T) {
				s, got, e := p.parse(c.input)
				if e != nil {
					t.Fatal(e)
				}
				if s != c.want {
					t.Errorf("String got %q want %q", s, c.want)
				}
				formTimes(t, got, c.times...)
				s2, again, e := p.parse(s)
				if e != nil {
					t.Fatal(e)
				}
				if s2 != c.want {
					t.Errorf("second String got %q", s2)
				}
				formTimes(t, again, c.times...)
			})
		}
	}
}
func TestTimeFormsParameters(t *testing.T) {
	cases := []struct {
		name, start, until string
		bad                bool
		part               string
	}{
		{"explicit-floating", "VALUE=DATE-TIME:20201228T090000", "20201230T090000", false, ""},
		{"explicit-UTC", "VALUE=DATE-TIME:20201228T090000Z", "20201230T090000Z", false, ""},
		{"explicit-TZID", "VALUE=DATE-TIME;TZID=Australia/Sydney:20201228T090000", "20201229T220000Z", false, ""},
		{"explicit-TZID-reverse", "TZID=Australia/Sydney;VALUE=DATE-TIME:20201228T090000", "20201229T220000Z", false, ""},
		{"TZID-on-Z", "TZID=Australia/Sydney:20201228T090000Z", "20201230T090000Z", true, "TZID"},
		{"TZID-on-date", "TZID=Australia/Sydney:20201228", "20201230", true, "TZID"},
		{"TZID-on-canonical-date", "VALUE=DATE;TZID=Australia/Sydney:20201228", "20201230", true, "TZID"},
		{"TZID-on-canonical-date-reverse", "TZID=Australia/Sydney;VALUE=DATE:20201228", "20201230", true, "TZID"},
		{"DATE-declared-datetime", "VALUE=DATE:20201228T090000", "20201230", true, "DTSTART"},
		{"datetime-declared-DATE", "VALUE=DATE-TIME:20201228", "20201230", true, "DTSTART"},
	}
	for _, p := range formParsers() {
		for _, c := range cases {
			t.Run(p.name+"/"+c.name, func(t *testing.T) {
				_, _, e := p.parse("DTSTART;" + c.start + "\nRRULE:FREQ=DAILY;UNTIL=" + c.until)
				if c.bad {
					if e == nil || !strings.Contains(e.Error(), c.part) {
						t.Fatalf("want %s error, got %v", c.part, e)
					}
				} else if e != nil {
					t.Fatal(e)
				}
			})
		}
	}
	for _, c := range cases {
		t.Run("standalone/"+c.name, func(t *testing.T) {
			_, e := StrToDtStart(c.start, time.UTC)
			if c.bad {
				if e == nil || !strings.Contains(e.Error(), c.part) {
					t.Fatalf("want %s error, got %v", c.part, e)
				}
			} else if e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestTimeFormsOrder(t *testing.T) {
	for _, p := range formParsers() {
		for _, reverse := range []bool{false, true} {
			for _, utcUntil := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/reverse=%t/invalid=%t", p.name, reverse, utcUntil), func(t *testing.T) {
					start := "DTSTART:20201228T090000"
					rule := "RRULE:FREQ=DAILY;UNTIL=20201230T090000"
					if utcUntil {
						rule += "Z"
					}
					input := start + "\n" + rule
					if reverse {
						input = rule + "\n" + start
					}
					_, got, e := p.parse(input)
					if utcUntil {
						if e == nil || !strings.Contains(e.Error(), "UNTIL") {
							t.Fatalf("want UNTIL rejection, got %v", e)
						}
					} else {
						if e != nil {
							t.Fatal(e)
						}
						if len(got) != 3 {
							t.Fatal(got)
						}
					}
				})
			}
		}
	}
	for _, reverse := range []bool{false, true} {
		for _, bad := range []bool{false, true} {
			t.Run(fmt.Sprintf("inline/reverse=%t/invalid=%t", reverse, bad), func(t *testing.T) {
				start := "DTSTART=20201228T090000"
				until := "UNTIL=20201230T090000"
				if bad {
					until += "Z"
				}
				input := "FREQ=DAILY;" + start + ";" + until
				if reverse {
					input = "FREQ=DAILY;" + until + ";" + start
				}
				for _, p := range formParsers()[:3] {
					t.Run(p.name, func(t *testing.T) {
						_, got, e := p.parse(input)
						if bad {
							if e == nil || !strings.Contains(e.Error(), "UNTIL") {
								t.Fatalf("want UNTIL rejection: %v", e)
							}
						} else {
							if e != nil {
								t.Fatal(e)
							}
							if len(got) != 3 {
								t.Fatal(got)
							}
						}
					})
				}
			})
		}
	}
}
func TestTimeFormsSharedSetValidation(t *testing.T) {
	cases := []struct{ name, lines, part string }{
		{"overwritten-invalid", "RRULE:FREQ=DAILY;UNTIL=20201230T090000Z\nRRULE:FREQ=DAILY;UNTIL=20201230T090000", "UNTIL"},
		{"ignored-invalid", "RRULE:FREQ=DAILY;UNTIL=20201230T090000\nEXRULE:FREQ=DAILY;UNTIL=20201230T090000Z", "UNTIL"},
		{"inline-bypass", "RRULE:FREQ=DAILY;DTSTART=20201228T090000Z;UNTIL=20201230T090000Z", "DTSTART"},
		{"inline-anchor", "RRULE:FREQ=DAILY;DTSTART=20201229T090000;UNTIL=20201230T090000", "DTSTART"},
		{"valid-last-wins-ignore", "RRULE:FREQ=WEEKLY;UNTIL=20201230T090000\nEXRULE:FREQ=DAILY;UNTIL=20201230T090000\nRRULE:FREQ=DAILY;UNTIL=20201230T090000", ""},
		{"matching-inline", "RRULE:FREQ=DAILY;DTSTART=20201228T090000;UNTIL=20201230T090000", ""},
	}
	for _, p := range formParsers()[3:] {
		for _, c := range cases {
			for pos := 0; pos < 3; pos++ {
				t.Run(fmt.Sprintf("%s/%s/start-pos%d", p.name, c.name, pos), func(t *testing.T) {
					lines := strings.Split(c.lines, "\n")
					at := pos
					if at > len(lines) {
						at = len(lines)
					}
					lines = append(lines, "")
					copy(lines[at+1:], lines[at:])
					lines[at] = "DTSTART:20201228T090000"
					_, got, e := p.parse(strings.Join(lines, "\n"))
					if c.part != "" {
						if e == nil || !strings.Contains(e.Error(), c.part) {
							t.Fatalf("want %s error, got %v", c.part, e)
						}
					} else {
						if e != nil {
							t.Fatal(e)
						}
						if len(got) != 3 {
							t.Fatal(got)
						}
					}
				})
			}
		}
	}
	// Keep original invalid legacy fixtures as explicit rejections.
	for _, in := range []string{
		"DTSTART;TZID=America/New_York:20180101T090000\nRRULE:FREQ=WEEKLY;COUNT=2\nEXRULE:FREQ=MONTHLY;UNTIL=20180520;BYMONTHDAY=1,2,3",
		"DTSTART;TZID=America/Los_Angeles:20181115T000000\nRRULE:FREQ=DAILY;INTERVAL=1;WKST=SU;UNTIL=20181117T235959",
	} {
		for _, p := range formParsers()[3:] {
			t.Run(p.name+"/old-fixture/"+in, func(t *testing.T) {
				_, _, e := p.parse(in)
				if e == nil || !strings.Contains(e.Error(), "UNTIL") {
					t.Fatalf("want UNTIL error, got %v", e)
				}
			})
		}
	}
}
func TestTimeFormsCopiesAndSetters(t *testing.T) {
	original := "DTSTART:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201230T090000"
	o, e := StrToROption(original)
	if e != nil {
		t.Fatal(e)
	}
	check := func(name string, c ROption, start, until string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			r, e := NewRRule(c)
			if e != nil {
				t.Fatal(e)
			}
			s := r.String()
			if !strings.HasPrefix(s, start+"\n") || !strings.Contains(s, "UNTIL="+until) {
				t.Fatalf("got %q", s)
			}
			if o.String() != original {
				t.Fatalf("source changed: %s", o.String())
			}
		})
	}
	c := *o
	check("copied", c, "DTSTART:20201228T090000", "20201230T090000")
	c.Dtstart = c.Dtstart.Add(time.Hour)
	check("replace-start", c, "DTSTART:20201228T100000Z", "20201230T090000")
	c = *o
	c.Until = c.Until.Add(time.Hour)
	check("replace-until", c, "DTSTART:20201228T090000", "20201230T100000Z")
	other := time.FixedZone("Other", 3600)
	c = *o
	c.Dtstart = c.Dtstart.In(other)
	check("same-instant-other-start-location", c, "DTSTART;TZID=Other:20201228T100000", "20201230T090000")
	c = *o
	c.Until = c.Until.In(other)
	check("same-instant-other-until-location", c, "DTSTART:20201228T090000", "20201230T090000Z")
	r, _ := NewRRule(*o)
	r.DTStart(o.Dtstart)
	if want := "DTSTART:20201228T090000Z\nRRULE:FREQ=DAILY;UNTIL=20201230T090000"; r.String() != want {
		t.Errorf("public start: %s", r.String())
	}
	r, _ = NewRRule(*o)
	r.Until(o.Until)
	if want := "DTSTART:20201228T090000\nRRULE:FREQ=DAILY;UNTIL=20201230T090000Z"; r.String() != want {
		t.Errorf("public until: %s", r.String())
	}
	s, e := StrToRRuleSet(original)
	if e != nil {
		t.Fatal(e)
	}
	replacement, _ := StrToRRule("FREQ=DAILY;COUNT=3")
	s.RRule(replacement)
	if want := "DTSTART:20201228T090000\nRRULE:FREQ=DAILY;COUNT=3"; s.String() != want {
		t.Errorf("private shared transfer: %s", s.String())
	}
	s.DTStart(o.Dtstart)
	if !strings.HasPrefix(s.String(), "DTSTART:20201228T090000Z\n") {
		t.Errorf("public set replacement: %s", s.String())
	}
	if o.String() != original {
		t.Fatal("shared mutation")
	}
}
func TestTimeFormsExactLocations(t *testing.T) {
	ny, e := time.LoadLocation("America/New_York")
	if e != nil {
		t.Fatal(e)
	}
	data := make([]byte, 54)
	copy(data, "TZif")
	binary.BigEndian.PutUint32(data[36:40], 1)
	binary.BigEndian.PutUint32(data[40:44], 4)
	binary.BigEndian.PutUint32(data[44:48], 2*3600)
	copy(data[50:], "XTZ\x00")
	custom, e := time.LoadLocationFromTZData("America/New_York", data)
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		name  string
		loc   *time.Location
		times []string
	}{
		{"IANA", ny, []string{"2026-03-07T14:00:00Z", "2026-03-08T13:00:00Z", "2026-03-09T13:00:00Z"}},
		{"fixed", time.FixedZone("America/New_York", -5*3600), []string{"2026-03-07T14:00:00Z", "2026-03-08T14:00:00Z", "2026-03-09T14:00:00Z"}},
		{"TZif", custom, []string{"2026-03-07T07:00:00Z", "2026-03-08T07:00:00Z", "2026-03-09T07:00:00Z"}},
	} {
		for _, bounded := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/until=%t", c.name, bounded), func(t *testing.T) {
				input := "DTSTART:20260307T090000\nRRULE:FREQ=DAILY;COUNT=3"
				if bounded {
					input = "DTSTART:20260307T090000\nRRULE:FREQ=DAILY;UNTIL=20260309T090000"
				}
				o, e := StrToROptionInLocation(input, c.loc)
				if e != nil {
					t.Fatal(e)
				}
				if o.Dtstart.Location() != c.loc || bounded && o.Until.Location() != c.loc {
					t.Fatal("parser replaced actual location")
				}
				copy := *o
				r, e := NewRRule(copy)
				if e != nil {
					t.Fatal(e)
				}
				if r.GetDTStart().Location() != c.loc {
					t.Fatal("constructor replaced location")
				}
				if r.String() != input {
					t.Errorf("floating form: %s", r.String())
				}
				formTimes(t, r.All(), c.times...)
				s := Set{}
				s.RRule(r)
				if s.GetDTStart().Location() != c.loc || s.String() != input {
					t.Errorf("set lost descriptor/location: %s", s.String())
				}
				formTimes(t, s.All(), c.times...)
				parsed, e := StrSliceToRRuleSetInLoc(strings.Split(input, "\n"), c.loc)
				if e != nil {
					t.Fatal(e)
				}
				if parsed.GetDTStart().Location() != c.loc || parsed.String() != input {
					t.Errorf("set parser: %s", parsed.String())
				}
				formTimes(t, parsed.All(), c.times...)
				r.DTStart(copy.Dtstart)
				if r.GetDTStart().Location() != c.loc {
					t.Fatal("setter replaced pointer")
				}
				formTimes(t, r.All(), c.times...)
				if bounded {
					r.Until(copy.Until)
					if r.GetUntil().Location() != c.loc {
						t.Fatal("until setter replaced pointer")
					}
					formTimes(t, r.All(), c.times...)
				}
			})
		}
	}
}
func TestTimeFormsFragmentsAndProgrammatic(t *testing.T) {
	loc := time.FixedZone("America/New_York", -5*3600)
	for _, value := range []string{"20201230", "20201230T090000"} {
		o, e := StrToROptionInLocation("FREQ=DAILY;UNTIL="+value, loc)
		if e != nil {
			t.Fatal(e)
		}
		if o.Until.Location() != loc {
			t.Fatal("fragment lost pointer")
		}
		if !strings.HasSuffix(o.RRuleString(), "Z") {
			t.Fatal(o.String())
		}
	}
	r, e := NewRRule(ROption{Freq: DAILY, Dtstart: time.Date(2020, 12, 28, 9, 0, 0, 0, loc), Until: time.Date(2020, 12, 30, 9, 0, 0, 0, loc)})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasSuffix(r.String(), "UNTIL=20201230T140000Z") {
		t.Fatal(r.String())
	}
}
func TestTimeFormsIntegratedGapBysetpos(t *testing.T) {
	ny, e := time.LoadLocation("America/New_York")
	if e != nil {
		t.Fatal(e)
	}
	input := "DTSTART:20260307T020000\nRRULE:FREQ=DAILY;UNTIL=20260309T020000;BYSETPOS=1;BYHOUR=2,3"
	o, e := StrToROptionInLocation(input, ny)
	if e != nil {
		t.Fatal(e)
	}
	r, e := NewRRule(*o)
	if e != nil {
		t.Fatal(e)
	}
	formTimes(t, r.All(), "2026-03-07T07:00:00Z", "2026-03-08T07:00:00Z", "2026-03-09T06:00:00Z")
	if r.String() != input {
		t.Fatal(r.String())
	}
	s, e := StrSliceToRRuleSetInLoc(strings.Split(input, "\n"), ny)
	if e != nil {
		t.Fatal(e)
	}
	formTimes(t, s.All(), "2026-03-07T07:00:00Z", "2026-03-08T07:00:00Z", "2026-03-09T06:00:00Z")
	if s.String() != input {
		t.Fatal(s.String())
	}
}

func TestTimeFormsExplicitGapTextAndUntil(t *testing.T) {
	t.Run("start-text", func(t *testing.T) {
		input := "DTSTART;TZID=America/New_York:20070311T023000\nRRULE:FREQ=DAILY;COUNT=2"
		r, e := StrToRRule(input)
		if e != nil {
			t.Fatal(e)
		}
		if r.String() != input {
			t.Errorf("nominal anchor text lost: %s", r.String())
		}
		formTimes(t, r.All(), "2007-03-11T07:30:00Z", "2007-03-12T06:30:00Z")
	})
	t.Run("until-gap", func(t *testing.T) {
		ny, e := time.LoadLocation("America/New_York")
		if e != nil {
			t.Fatal(e)
		}
		o, e := StrToROptionInLocation("DTSTART:20260308T003000\nRRULE:FREQ=HOURLY;UNTIL=20260308T023000", ny)
		if e != nil {
			t.Fatal(e)
		}
		r, e := NewRRule(*o)
		if e != nil {
			t.Fatal(e)
		}
		formTimes(t, r.All(), "2026-03-08T05:30:00Z", "2026-03-08T06:30:00Z", "2026-03-08T07:30:00Z")
	})
}
func TestTimeFormsLegacyExruleUnknownControl(t *testing.T) {
	_, e := StrToRRuleSet("DTSTART:20201228T090000\nRRULE:FREQ=DAILY;COUNT=3\nEXRULE:IGNORED=legacy;UNTIL=20201230T090000")
	if e != nil {
		t.Fatal(e)
	}
}
func TestTimeFormsDateCopyAndSet(t *testing.T) {
	input := "DTSTART:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230"
	o, e := StrToROption(input)
	if e != nil {
		t.Fatal(e)
	}
	copy := *o
	r, e := NewRRule(copy)
	if e != nil {
		t.Fatal(e)
	}
	set := Set{}
	set.RRule(r)
	want := "DTSTART;VALUE=DATE:20201228\nRRULE:FREQ=DAILY;UNTIL=20201230"
	if r.String() != want || set.String() != want {
		t.Errorf("DATE descriptor lost: %s / %s", r.String(), set.String())
	}
	r.Until(o.Until)
	if !strings.HasPrefix(r.String(), "DTSTART;VALUE=DATE:20201228\n") || !strings.HasSuffix(r.String(), "UNTIL=20201230T000000Z") {
		t.Fatal(r.String())
	}
	if o.String() != want {
		t.Fatal("source descriptor changed")
	}
}

func TestTimeFormsSeparateStartCannotBeOverridden(t *testing.T) {
	for _, p := range formParsers() {
		for _, bad := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/bad=%t", p.name, bad), func(t *testing.T) {
				inline := "20201228T090000"
				if bad {
					inline = "20201229T090000"
				}
				_, got, e := p.parse("DTSTART:20201228T090000\nRRULE:FREQ=DAILY;DTSTART=" + inline + ";UNTIL=20201230T090000")
				if bad {
					if e == nil || !strings.Contains(e.Error(), "DTSTART") {
						t.Fatalf("want shared DTSTART rejection, got %v (%v)", e, got)
					}
				} else {
					if e != nil {
						t.Fatal(e)
					}
					if len(got) != 3 {
						t.Fatal(got)
					}
				}
			})
		}
	}
}

func TestTimeFormsDenseParsedAndHorizon(t *testing.T) {
	values := func(n int) string {
		v := make([]string, n)
		for i := range v {
			v[i] = fmt.Sprint(i)
		}
		return strings.Join(v, ",")
	}
	input := "DTSTART:20230101T000000\nRRULE:FREQ=YEARLY;COUNT=4;BYSETPOS=1,366,-366,-1;BYDAY=MO,TU,WE,TH,FR,SA,SU;BYHOUR=" + values(24) + ";BYMINUTE=" + values(60) + ";BYSECOND=" + values(60)
	for _, p := range formParsers() {
		t.Run(p.name+"/dense", func(t *testing.T) {
			_, got, e := p.parse(input)
			if e != nil {
				t.Fatal(e)
			}
			formTimes(t, got, "2023-01-01T00:00:00Z", "2023-01-01T00:06:05Z", "2023-12-31T23:53:54Z", "2023-12-31T23:59:59Z")
		})
		t.Run(p.name+"/short-UNTIL", func(t *testing.T) {
			_, got, e := p.parse("DTSTART:20260105T100000\nRRULE:FREQ=DAILY;UNTIL=20260201T000000;BYMONTH=2;BYMONTHDAY=30")
			if e != nil {
				t.Fatal(e)
			}
			if len(got) != 0 {
				t.Fatal(got)
			}
		})
	}
}
