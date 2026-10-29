package rrule_test

import (
	rrule "github.com/kentonium3/rrule-go"
	"testing"
	"time"
)

func TestTimeFormsKeyedPublicAPI(t *testing.T) {
	o := rrule.ROption{Freq: rrule.DAILY, Dtstart: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC), Count: 2}
	r, e := rrule.NewRRule(o)
	if e != nil {
		t.Fatal(e)
	}
	var start func(time.Time) = r.DTStart
	var until func(time.Time) = r.Until
	start(o.Dtstart)
	until(o.Dtstart.AddDate(0, 0, 2))
	set := rrule.Set{}
	set.RRule(r)
	if len(set.All()) != 2 {
		t.Fatal(set.String())
	}
}
