package state

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestMock(t *testing.T) {
	mock := clock.NewMock()
	SetClock(mock)
	year := GetClock().Now().UTC().Year()
	if year != 1970 {
		t.Fatal("wrong year: ", year)
	}
}

func TestSetMock(t *testing.T) {
	mock := clock.NewMock()
	SetClock(mock)
	if GetClock().Now().UTC().Year() != 1970 {
		t.Fatal("wrong year")
	}

	d, e := time.ParseDuration("24h")
	if e != nil {
		t.Fatal("could not parse duration")
	}
	mock.Set(GetClock().Now().UTC().Add(3 * 365 * d))
	year := GetClock().Now().UTC().Year()
	if year != 1972 {
		t.Fatal("wrong year: ", year)
	}
}
