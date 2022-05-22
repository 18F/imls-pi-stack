package tlp

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/structs"
)

func ProcessData(sq *state.Queue[int64]) bool {
	// Queue up what needs to be sent still.
	thisSession := state.GetCurrentSessionId()
	sq.Enqueue(thisSession)

	log.Debug().
		Int64("session id", thisSession).
		Msg("queueing current session to the send queue")

	pidCounter := 0
	durations := make([]structs.Duration, 0)

	log.Debug().
		Int("len", state.GetEphemeralDBLength()).
		Msg("MACs in the ephemeral DB")

	for _, se := range state.GetMACs() {

		d := structs.Duration{
			ID:        pidCounter,
			PiSerial:  viper.GetString("device.serial"),
			SessionID: fmt.Sprint(state.GetCurrentSessionId()),
			FCFSSeqID: viper.GetString("device.fcfsSeqId"),
			DeviceTag: viper.GetString("device.tag"),
			PatronID:  pidCounter,
			// FIXME: All times should become UNIX epoch seconds...
			Start: se.Start,
			End:   se.End}

		durations = append(durations, d)
		pidCounter += 1
	}

	log.Debug().
		Int("len", len(durations)).
		Msg("durations in this session")

	state.StoreManyDurations(thisSession, durations)
	return true
}
