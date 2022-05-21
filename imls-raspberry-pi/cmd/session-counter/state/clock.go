package state

import (
	"time"

	"github.com/benbjohnson/clock"
	"gsa.gov/18f/internal/interfaces"
	"gsa.gov/18f/internal/state"
)

var clockSingleton *clock.Clock

func GetClock() clock.Clock {
	if clockSingleton == nil {
		c := clock.New()
		clockSingleton = &c
	}
	return *clockSingleton
}

func SetClock(clock clock.Clock) {
	clockSingleton = &clock
}

func GetYesterday(cfg interfaces.Config) time.Time {
	offset := -24
	yesterday := state.GetClock().Now().In(time.Local).Add(time.Duration(offset) * time.Hour)
	return yesterday
}
