package state

import (
	"time"

	"github.com/benbjohnson/clock"
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

func GetYesterday() time.Time {
	offset := -24
	yesterday := GetClock().Now().In(time.Local).Add(time.Duration(offset) * time.Hour)
	return yesterday
}
