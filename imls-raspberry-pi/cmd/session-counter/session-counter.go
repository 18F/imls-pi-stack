package main

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/tlp"

	"gsa.gov/18f/internal/wifi-hardware-search/search"
)

func initConfigFromFlags() {
	// versionPtr := flag.Bool("version", false, "Get the software version and exit.")
	// // configPathPtr := flag.String("config", "", "Path to config.sqlite. REQUIRED.")
	// flag.Parse()

	// // If they just want the version, print and exit.
	// if *versionPtr {
	// 	fmt.Println(version.GetVersion())
	// 	os.Exit(0)
	// }

	// // Make sure a config is passed.
	// if *configPathPtr == "" {
	// 	log.Fatal("The flag --config MUST be provided.")
	// 	os.Exit(1)
	// }

	// if _, err := os.Stat(*configPathPtr); os.IsNotExist(err) {
	// 	log.Println("Looked for config at ", *configPathPtr)
	// 	log.Fatal("Cannot find config file. Exiting.")
	// }

	// Using Viper; just path --config
	// state.SetConfigAtPath(*configPathPtr)

}

func runEvery(crontab string, c *cron.Cron, fun func()) {
	// cfg := state.GetConfig()
	// id was first param
	_, err := c.AddFunc(crontab, fun)
	// cfg.Log().Debug("launched crontab ", crontab, " with id ", id)
	if err != nil {
		// cfg.Log().Error("cron: could not set up crontab entry")
		// cfg.Log().Fatal(err.Error())
	}
}

func run2() {
	cfg := state.GetConfig()
	sq := state.NewQueue("to_send")
	iq := state.NewQueue("images")
	durationsdb := state.GetDurationsDatabase()
	c := cron.New()

	go runEvery("*/1 * * * *", c,
		func() {
			// cfg.Log().Debug("RUNNING SIMPLESHARK")
			tlp.SimpleShark(
				search.SetMonitorMode,
				search.SearchForMatchingDevice,
				tlp.TSharkRunner)
		})

	go runEvery(cfg.GetString("cron.reset"), c,
		func() {
			// cfg.Log().Info("RUNNING PROCESSDATA at ", state.GetClock().Now().In(time.Local))
			// Copy ephemeral durations over to the durations table
			tlp.ProcessData(durationsdb, sq, iq)
			// Draw images of the data
			//tlp.WriteImages(durationsdb)
			// Try sending the data
			tlp.SimpleSend(durationsdb, sq)
			// Increment the session counter
			state.IncrementSessionId()
			// Clear out the ephemeral data for the next day of monitoring
			state.ClearEphemeralDB()
		})

	// go runEvery()

	// Start the cron jobs...
	c.Start()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().
		Str("msg", "startup session id").
		Int64("session_id", state.GetCurrentSessionId())

	var wg sync.WaitGroup
	wg.Add(1)
	go run2()

	// Stay a while. STAY FOREVER!
	// https://en.wikipedia.org/wiki/Impossible_Mission
	wg.Wait()
}
