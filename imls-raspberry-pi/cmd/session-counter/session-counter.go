package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/tlp"

	"gsa.gov/18f/internal/wifi-hardware-search/search"
)

var (
	cfgFile  string
	logLevel string
)

func runEvery(crontab string, c *cron.Cron, fun func()) {

	id, err := c.AddFunc(crontab, fun)
	log.Debug().
		Str("crontab", crontab).
		Str("id", fmt.Sprintf("%v", id)).
		Msg("runEvery")

	if err != nil {
		// cfg.Log().Error("cron: could not set up crontab entry")
		// cfg.Log().Fatal(err.Error())
	}
}

func run2() {
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

	go runEvery(viper.GetString("cron.reset"), c,
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

func launchTLP() {
	log.Info().
		Int64("session_id", state.GetCurrentSessionId()).
		Msg("session id at launch")

	var wg sync.WaitGroup
	wg.Add(1)
	go run2()
	// Stay a while. STAY FOREVER!
	// https://en.wikipedia.org/wiki/Impossible_Mission
	wg.Wait()
}

var rootCmd = &cobra.Command{
	Use:   "session-counter",
	Short: "A tool for monitoring wifi devices while preserving privacy.",
	Long: `session-counter watches to see what wifi devices are 
nearby while carefully leaving out information that would impose 
on the privacy of people around you.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchTLP()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of session-counter",
	Long:  `All software has versions. This is session-counter's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.1.0")
	},
}

func initLogs() {
	switch lvl := logLevel; lvl {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func initConfig() {
	initLogs()
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(filepath.Join(home, ".session-counter"))
		viper.AddConfigPath(".")

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msg(viper.ConfigFileUsed())
	} else {
		panic("could not find config. exiting.")
	}

}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.session-counter/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "logging", "info", "logging level (debug, info, warn, error)")

	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(versionCmd)
	rootCmd.Execute()
}
