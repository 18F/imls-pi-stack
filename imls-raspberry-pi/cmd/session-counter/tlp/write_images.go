package tlp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gsa.gov/18f/internal/analysis"
	"gsa.gov/18f/internal/interfaces"
	"gsa.gov/18f/internal/logwrapper"
	"gsa.gov/18f/internal/state"
	"gsa.gov/18f/internal/structs"
)

//This must happen after the data is updated for the day.
func writeImages(durations []structs.Duration, sessionid string) error {
	cfg := state.GetConfig()
	lw := logwrapper.NewLogger(nil)
	var reterr error

	if _, err := os.Stat(cfg.GetString("paths.root")); os.IsNotExist(err) {
		err := os.Mkdir(cfg.GetString("paths.root"), 0777)
		if err != nil {
			lw.Error("could not create web directory: ", cfg.GetString("paths.root"))
			reterr = err
		}
	}
	if _, err := os.Stat(cfg.GetString("paths.images")); os.IsNotExist(err) {
		err := os.Mkdir(cfg.GetString("paths.images"), 0777)
		if err != nil {
			lw.Error("could not create image directory")
			reterr = err
		}
	}

	// FIXME: This filename kinda makes no sense if we're not running
	// a reset on a daily basis at midnight.
	// yesterday := model.GetYesterday(cfg)
	yesterday := state.GetClock().Now().In(time.Local)
	imageFilename := fmt.Sprintf("%04d%02d%02d-%v-%v_%v.png",
		yesterday.Year(),
		int(yesterday.Month()),
		int(yesterday.Day()),
		sessionid,
		cfg.GetString("device.FCFSSeqId"),
		cfg.GetString("device.tag"))

	path := filepath.Join(cfg.GetString("paths.images"), imageFilename)
	// func DrawPatronSessions(cfg *config.Config, durations []Duration, outputPath string) {
	analysis.DrawPatronSessions(durations, path)
	return reterr
}

func WriteImages(db interfaces.Database, iq *state.List) {
	// cfg := state.GetConfig()
	// iq := state.NewList("images")
	imagesToWrite := iq.AsList()
	// cfg.Log().Info("sessions ", imagesToWrite, " are waiting to be written on the image queue")
	for _, nextImage := range imagesToWrite {
		durations := []structs.Duration{}
		var count int
		db.GetPtr().QueryRow("SELECT COUNT(*) FROM durations WHERE session_id=?", nextImage).Scan(&count)
		// cfg.Log().Info("Found ", count, " durations to filter down...")
		err := db.GetPtr().Select(&durations, "SELECT * FROM durations WHERE session_id=?", nextImage)
		// cfg.Log().Debug("found ", len(durations), " durations in WriteImages")
		if err != nil {
			// cfg.Log().Info("error in extracting durations for session", nextImage)
			// cfg.Log().Error(err.Error())
		} else {
			err = writeImages(durations, nextImage)
			if err != nil {
				// cfg.Log().Error("error in writing images")
				// cfg.Log().Error(err.Error())
			} else {
				iq.Remove(nextImage)
			}
		}
	}
}
