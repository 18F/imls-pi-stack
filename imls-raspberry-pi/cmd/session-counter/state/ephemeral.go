package state

import (
	"crypto/sha1"
	"fmt"
	"time"

	"gsa.gov/18f/cmd/session-counter/structs"
)

type StartEnd struct {
	Start int64
	End   int64
}

type EphemeralMACDB map[string]StartEnd

var emd EphemeralMACDB = make(EphemeralMACDB)

func GetMACs() EphemeralMACDB {
	return emd
}

func ClearEphemeralMACDB() {
	emd = make(EphemeralMACDB)
}

// NOTE: Do not log MAC addresses.
func RecordMAC(mac string) {
	now := GetClock().Now().In(time.Local).Unix()
	// cfg := GetConfig()
	// cfg.Log().Debug("THE TIME IS NOW ", GetClock().Now().In(time.Local), " or ", now)

	// Check if we already have the MAC address in the ephemeral table.
	if p, ok := emd[mac]; ok {
		//cfg.Log().Debug(mac, " exists, updating")
		// Has this device been away for more than 2 hours?
		// Start by grabbing the start/end times.
		se := emd[mac]
		if (now > se.End) && ((now - se.End) > MAC_MEMORY_DURATION_SEC) {
			// If it has been, we need to "forget" the old device.
			// Do this by hashing the mac with the current time, store the original data
			// unchanged, and create a new entry for the current mac address, in case we
			// see it again (in less than 2h).
			// cfg.Log().Debug(mac, " is an old mac, refreshing/changing")
			sha1 := sha1.Sum([]byte(mac + fmt.Sprint(now)))
			emd[fmt.Sprintf("%x", sha1)] = se
			emd[mac] = StartEnd{Start: now, End: now}
		} else {
			// Just update the mac address. It has been less than 2h.
			emd[mac] = StartEnd{Start: p.Start, End: now}
		}
	} else {
		// We have never seen the MAC address.
		//cfg.Log().Debug(mac, " is new, inserting")
		emd[mac] = StartEnd{Start: now, End: now}
	}
}

type EphemeralDurationsDB map[int64][]structs.Duration

var edd EphemeralDurationsDB = make(EphemeralDurationsDB)

func StoreManyDurations(session_id int64, durations []structs.Duration) {
	if _, ok := edd[session_id]; ok {
		edd[session_id] = append(edd[session_id], durations...)
	} else {
		edd[session_id] = make([]structs.Duration, 0)
		edd[session_id] = append(edd[session_id], durations...)
	}

}

func GetDurations(session_id int64) []structs.Duration {
	if db, ok := edd[session_id]; ok {
		return db
	} else {
		return make([]structs.Duration, 0)
	}
}

func ClearAllEphemeralDurations() {
	edd = make(EphemeralDurationsDB)
}

func ClearEphemeralDurationsSession(session_id int64) {
	edd[session_id] = make([]structs.Duration, 0)
}
