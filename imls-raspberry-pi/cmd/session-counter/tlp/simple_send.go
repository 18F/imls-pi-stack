package tlp

import (
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/internal/http"
)

func SimpleSend(sq *state.Queue[int64]) {
	// cfg.Log().Debug("Starting BatchSend")
	// This only comes in on reset...
	// sq := state.NewList("to_send")
	sessionsToSend := sq.AsList()

	for _, nextSessionIDToSend := range sessionsToSend {
		durations := state.GetDurations(nextSessionIDToSend)

		log.Debug().
			Int("duration count", len(durations)).
			Int64("session id", nextSessionIDToSend).
			Msg("SimpleSend")

		if len(durations) == 0 {
			// cfg.Log().Info("found zero durations to send/draw. dequeing session [", nextSessionIDToSend, "]")
			log.Debug().Msg("found zero durations to send. dequeueing.")
			sq.Remove(nextSessionIDToSend)
		} else if state.IsStoringToAPI() {
			// cfg.Log().Info("attempting to send batch [", nextSessionIDToSend, "][", len(durations), "] to the API server")
			// convert []Duration to an array of map[string]interface{}
			data := make([]map[string]interface{}, 0)
			for _, duration := range durations {
				data = append(data, duration.AsMap())
			}
			// After writing images, we come back and try and send the data remotely.
			// cfg.Log().Debug("PostJSONing ", len(data), " duration datas")
			err := http.PostJSON(viper.GetString("storage.durationsURI"), data)
			if err != nil {
				log.Info().
					Int64("session id", nextSessionIDToSend).
					Err(err).
					Msg("could not log into API; session id left on queue.")
			} else {
				// If we successfully sent the data remotely, we can now mark it is as sent.
				sq.Remove(nextSessionIDToSend)
			}
		} else {
			// Always dequeue. We're storing locally "for free" into the
			// durations table before trying to do the send.
			// cfg.Log().Info("not in API mode, not sending data...")
			// state.DumpEphemeralDurationsDB(nextSessionIDToSend)
			sq.Remove(nextSessionIDToSend)
		}
	}

}
