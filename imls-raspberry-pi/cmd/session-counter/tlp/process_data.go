package tlp

import (
	"fmt"

	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/structs"
)

func ProcessData(sq *state.Queue[int64]) bool {
	// Queue up what needs to be sent still.
	thisSession := state.GetCurrentSessionId()
	// cfg.Log().Debug("queueing current session [ ", thissession, " ] to images and send queue... ")
	if thisSession >= 0 {
		sq.Enqueue(thisSession)
	}

	pidCounter := 0
	durations := make([]structs.Duration, 0)

	for _, se := range state.GetMACs() {

		d := structs.Duration{
			PiSerial:  viper.GetString("device.serial"),
			SessionID: fmt.Sprint(state.GetCurrentSessionId()),
			FCFSSeqID: viper.GetString("device.fcfsSeqId"),
			DeviceTag: viper.GetString("device.tag"),
			PatronID:  pidCounter,
			// FIXME: All times should become UNIX epoch seconds...
			Start: se.Start,
			End:   se.End}

		//dDB.GetTableFromStruct(structs.Duration{}).InsertStruct(d)
		durations = append(durations, d)
		pidCounter += 1
	}

	//dDB.GetTableFromStruct(structs.Duration{}).InsertMany(durations)
	state.StoreManyDurations(thisSession, durations)
	return true
}
