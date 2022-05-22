package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/tlp"
	"gsa.gov/18f/internal/wifi-hardware-search/models"
)

var NUMMACS int
var NUMFOUNDPERMINUTE int
var consistentMACs = make([]string, 0)

type CheckFun func(int) bool

func generateFakeMac() string {
	var letterRunes = []rune("ABCDEF0123456789")
	b := make([]rune, 17)
	colons := [...]int{2, 5, 8, 11, 14}
	for i := 0; i < 17; i++ {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]

		for v := range colons {
			if i == colons[v] {
				b[i] = rune(':')
			}
		}
	}
	return string(b)
}

var mac_counter int = 0

func generateDeterministicMAC() string {
	mac := fmt.Sprintf("00:00:00:00:00:%02x", mac_counter%NUMMACS)
	mac_counter += 1
	return mac

}

func runFakeWireshark(device string) []string {

	thisTime := rand.Intn(NUMFOUNDPERMINUTE)
	send := make([]string, thisTime)
	// log.Debug().
	// 	Int("count", thisTime).
	// 	Msg("devices from fake wireshark run")

	for i := 0; i < thisTime; i++ {
		send[i] = consistentMACs[rand.Intn(len(consistentMACs))]
	}
	return send
}

func runDeterministicFakeWireshark(device string) []string {

	thisTime := NUMFOUNDPERMINUTE
	send := make([]string, thisTime)
	// log.Debug().
	// 	Int("count", thisTime).
	// 	Msg("devices from fake wireshark run")

	for i := 0; i < thisTime; i++ {
		send[i] = consistentMACs[i]
	}
	return send
}

func isItMidnight(now time.Time) bool {
	return (now.Hour() == 0 &&
		now.Minute() == 0 &&
		now.Second() == 0)

}

func checkDeterministic(sq *state.Queue[int64], rundays int, nummacs int, numfoundperminute int, check CheckFun) bool {
	// In a deterministic run, I will have nummacs in the pool.
	// Every minute, I will find the first numfoundperminute of them.
	// That will happen every minute of the day. This means I should... filter them all out. :/
	session, _ := sq.Peek()
	ds := state.GetDurations(session)
	log.Debug().
		Msg(fmt.Sprintf("%v", ds))

	ok := true

	if len(ds) == nummacs {
		ok = ok && true
	}

	if numfoundperminute < nummacs {
		for _, d := range ds {
			// We are one minute short of a day for testing.
			// Why? Because... we don't count 00:00:00 of the next day in our current day.
			log.Debug().
				Int64("id", int64(d.ID)).
				Int64("duration", d.End-d.Start).
				Int("target", rundays*60*60*24-60).
				Msg("device")
			if d.End-d.Start == int64(rundays*60*60*24-60) {
				ok = ok && true
			} else {
				ok = ok && false
			}
		}
	}

	if ok {
		log.Debug().
			Msg("everything is OK")
	} else {
		log.Debug().
			Msg("everything is bad")
	}

	return ok
}

func MockRun(rundays int, nummacs int, numfoundperminute int, deterministic bool, check func(int) bool) bool {
	sq := state.NewQueue[int64]("to_send")

	// Create a pool of NUMMACS devices to draw from.
	// We will send NUMFOUNDPERMINUTE each minute
	NUMMACS = nummacs
	NUMFOUNDPERMINUTE = numfoundperminute
	consistentMACs = make([]string, NUMMACS)
	for i := 0; i < NUMMACS; i++ {
		if deterministic {
			consistentMACs[i] = generateDeterministicMAC()
		} else {
			consistentMACs[i] = generateFakeMac()
		}
	}

	log.Debug().
		Str("macs", fmt.Sprintf("%v", consistentMACs)).
		Msg("MACs in the pool")

	var chooseAShark func(string) []string
	if deterministic {
		chooseAShark = runDeterministicFakeWireshark
	} else {
		chooseAShark = runFakeWireshark
	}
	ok := true

	for days := 0; days < rundays; days++ {
		// Pretend we run once per minute for 24 hours
		for minutes := 0; minutes < 60*24; minutes++ {
			tlp.SimpleShark(
				// search.SetMonitorMode,
				func(d *models.Device) {},
				// search.SearchForMatchingDevice,
				func() *models.Device { return &models.Device{Exists: true, Logicalname: "fakewan0"} },
				// tlp.TSharkRunner
				chooseAShark)
			// Add one minute to the fake clock
			state.GetClock().(*clock.Mock).Add(1 * time.Minute)

			if isItMidnight(state.GetClock().(*clock.Mock).Now().In(time.Local)) {
				// Then we run the processing at midnight (once per 24 hours)
				log.Debug().
					Str("mock", fmt.Sprint(state.GetClock().Now().In(time.Local))).
					Msg("running ProcessData")

				// Copy ephemeral durations over to the durations table
				tlp.ProcessData(sq)
				// Try sending the data
				//tlp.SimpleSend(sq)
				if deterministic {
					ok = checkDeterministic(sq, 1, nummacs, numfoundperminute, check) && ok
				}
				// Increment the session counter
				state.IncrementSessionId()
				// Clear out the ephemeral data for the next day of monitoring
				state.ClearEphemeralMACDB()
			} else {
				// log.Debug().
				// 	Str("mock", fmt.Sprint(state.GetClock().(*clock.Mock).Now().In(time.Local))).
				// 	Msg("timecheck")
			}
		}

	}
	return ok
}

func TestAllUp(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println(filename)
	path := filepath.Dir(filename)
	viper.Set("storage.mode", "local")
	viper.Set("paths.root", filepath.Join(path, "test", "www"))
	viper.Set("paths.durations", filepath.Join(path, "test", "durations.sqlite"))

	// Fake the clock
	mock := clock.NewMock()
	// FIXME MCJ 20220522 I cannot get this to parse time into UTC correctly.
	// I JUST WANT MIDNIGHT.
	// I suspect this fails when run in a timezone other than EST.
	mt, _ := time.ParseInLocation("2006-01-02T15:04", "1975-10-11T04:00", time.UTC)
	mock.Set(mt.UTC())
	state.SetClock(mock)

	log.Info().Msg("=== Running 1 day ===")
	cf := func(n int) bool { return true }
	ok := MockRun(1, 20, 10, true, cf)
	if !ok {
		t.Fail()
	}
	time.Sleep(2 * time.Second)
	ok = MockRun(2, 20, 10, true, cf)
	if !ok {
		t.Fail()
	}
	time.Sleep(2 * time.Second)

	/*
		state.IncrementSessionId()
		// cfg.Log().Info("next session id: ", cfg.GetCurrentSessionID())
		log.Info().
			Int64("session id", state.GetCurrentSessionId()).
			Msg("next session id")

		// Fake the clock
		mt, _ = time.Parse("2006-01-02T15:04", "1976-11-12T00:01")
		mock.Set(mt)
		state.SetClock(mock)

		log.Info().Msg("=== Running 5 days ===")
		MockRun(5, 2000, 10, true)

		mt, _ = time.Parse("2006-01-02T15:04", "1978-01-01T00:01")
		mock.Set(mt)
		state.SetClock(mock)
	*/
	// MockRun(90, 2000000, 10)
}
