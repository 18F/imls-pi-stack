package tlp

import (
	"fmt"

	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/interfaces"
	"gsa.gov/18f/cmd/session-counter/state"
	"gsa.gov/18f/cmd/session-counter/structs"
)

func ProcessData(dDB interfaces.Database, sq *state.Queue, iq *state.Queue) bool {
	// Queue up what needs to be sent still.
	thissession := state.GetCurrentSessionId()
	// cfg.Log().Debug("queueing current session [ ", thissession, " ] to images and send queue... ")
	if thissession >= 0 {
		sq.Enqueue(fmt.Sprint(thissession))
		iq.Enqueue(fmt.Sprint(thissession))
	}

	pidCounter := 0
	durations := make([]interface{}, 0)

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

	dDB.GetTableFromStruct(structs.Duration{}).InsertMany(durations)
	return true
}
