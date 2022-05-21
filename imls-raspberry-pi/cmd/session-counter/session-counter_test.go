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

func isItMidnight(now time.Time) bool {
	return (now.Hour() == 0 &&
		now.Minute() == 0 &&
		now.Second() == 0)

}

func MockRun(rundays int, nummacs int, numfoundperminute int) {
	sq := state.NewQueue[int64]("to_send")

	// Create a pool of NUMMACS devices to draw from.
	// We will send NUMFOUNDPERMINUTE each minute
	NUMMACS = nummacs
	NUMFOUNDPERMINUTE = numfoundperminute
	consistentMACs = make([]string, NUMMACS)
	for i := 0; i < NUMMACS; i++ {
		consistentMACs[i] = generateFakeMac()
	}

	for days := 0; days < rundays; days++ {
		// Pretend we run once per minute for 24 hours
		for minutes := 0; minutes < 60*24; minutes++ {
			tlp.SimpleShark(
				// search.SetMonitorMode,
				func(d *models.Device) {},
				// search.SearchForMatchingDevice,
				func() *models.Device { return &models.Device{Exists: true, Logicalname: "fakewan0"} },
				// tlp.TSharkRunner
				runFakeWireshark)
			// Add one minute to the fake clock
			state.GetClock().(*clock.Mock).Add(1 * time.Minute)

			if isItMidnight(state.GetClock().Now().In(time.Local)) {
				// Then we run the processing at midnight (once per 24 hours)
				log.Debug().
					Str("mock", fmt.Sprint(state.GetClock().Now().In(time.Local))).
					Msg("running ProcessData")

				// Copy ephemeral durations over to the durations table
				tlp.ProcessData(sq)
				// Try sending the data
				tlp.SimpleSend(sq)
				// Increment the session counter
				state.IncrementSessionId()
				// Clear out the ephemeral data for the next day of monitoring
				state.ClearEphemeralMACDB()
			}
		}

	}
}

func TestAllUp(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println(filename)
	path := filepath.Dir(filename)
	viper.Set("storage.mode", "local")                          // SetStorageMode("local")
	viper.Set("paths.root", filepath.Join(path, "test", "www")) //SetRootPath(filepath.Join(path, "test", "www"))
	viper.Set("paths.durations", filepath.Join(path, "test", "durations.sqlite"))

	// Fake the clock
	mock := clock.NewMock()
	mt, _ := time.Parse("2006-01-02T15:04", "1975-10-11T00:01")
	mock.Set(mt)
	state.SetClock(mock)

	MockRun(1, 200000, 10)
	log.Info().Msg("WAITING")
	time.Sleep(2 * time.Second)

	state.IncrementSessionId()
	// cfg.Log().Info("next session id: ", cfg.GetCurrentSessionID())
	log.Info().
		Int64("session id", state.GetCurrentSessionId()).
		Msg("next session id")

	// Fake the clock
	mt, _ = time.Parse("2006-01-02T15:04", "1976-11-12T00:01")
	mock.Set(mt)
	state.SetClock(mock)

	MockRun(5, 200000, 10)

	mt, _ = time.Parse("2006-01-02T15:04", "1978-01-01T00:01")
	mock.Set(mt)
	state.SetClock(mock)

	// MockRun(90, 2000000, 10)
}
