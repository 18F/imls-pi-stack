// Package logwrapper wraps logging to various Write interfaces.
package logwrapper

import (
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"gsa.gov/18f/internal/http"
	"gsa.gov/18f/internal/state"
)

type APILogger struct {
	l   *StandardLogger
	cfg *viper.Viper
}

func NewAPILogger(cfg *viper.Viper) (api *APILogger) {
	api = &APILogger{cfg: cfg}
	return api
}

func (a *APILogger) Write(p []byte) (n int, err error) {

	if state.IsStoringLocally() {
		// do nothing.
		return len(p), nil
	}

	data := map[string]interface{}{
		"pi_serial":   state.GetSerial(),
		"fcfs_seq_id": viper.GetString("device.FCFSSeqId"),
		"device_tag":  viper.GetString("device.tag"),
		"session_id":  strconv.FormatInt(state.GetCurrentSessionId(), 10),
		"localtime":   time.Now().Format(time.RFC3339),
		"tag":         a.l.GetLogLevelName(),
		"info":        string(p),
	}

	err = http.PostJSON(viper.GetString("storage.eventsURI"), []map[string]interface{}{data})
	if err != nil {
		log.Println("could not log to API")
		log.Println(err.Error())
	}

	return len(p), nil
}
