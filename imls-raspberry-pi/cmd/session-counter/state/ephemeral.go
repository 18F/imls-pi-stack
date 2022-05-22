package state

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gsa.gov/18f/cmd/session-counter/structs"
)

// For how long do we recognize a device?
// 2 hours. This is 2 * 60 minutes * 60 seconds.
// If we see a MAC within this window, we "remember" it.
// If we see a MAC, 2h go by, and we see it again, we're going
// to "forget" the original sighting, and pretend the device is new.
const MAC_MEMORY_DURATION_SEC = 2 * 60 * 60

type StartEnd struct {
	Start int64
	End   int64
}

type EphemeralMACDB map[string]StartEnd

var emd EphemeralMACDB = make(EphemeralMACDB)

func GetEphemeralDBLength() int {
	return len(emd)
}

func GetMACs() EphemeralMACDB {
	return emd
}

func ClearEphemeralMACDB() {
	emd = make(EphemeralMACDB)
}

/*
// NOTE: Do not log MAC addresses.
func RecordMAC(mac string) {
	now := GetClock().Now().In(time.Local).Unix()
	// Check if we already have the MAC address in the ephemeral table.
	if p, ok := emd[mac]; ok {
		// Do not log MAC addresses.
		log.Debug().Str("mac", mac).Msg("recordmac: exists, updating")
		// Has this device been away for more than 2 hours?
		// Start by grabbing the start/end times.
		se := emd[mac]
		if (now > se.End) && ((now - se.End) > MAC_MEMORY_DURATION_SEC) {
			// If it has been, we need to "forget" the old device.
			// Do this by hashing the mac with the current time, store the original data
			// unchanged, and create a new entry for the current mac address, in case we
			// see it again (in less than 2h).
			// Do not log MAC addresses.
			log.Debug().Str("mac", mac).Msg("recordmac: refreshing/changing")
			sha1 := sha1.Sum([]byte(mac + fmt.Sprint(now)))
			emd[fmt.Sprintf("%x", sha1)] = se
			emd[mac] = StartEnd{Start: now, End: now}
		} else {
			// Just update the mac address. It has been less than 2h.
			emd[mac] = StartEnd{Start: p.Start, End: now}
		}
	} else {
		// We have never seen the MAC address.
		// Do not log MAC addresses.
		log.Debug().Str("mac", mac).Msg("recordmac: new, inserting")
		emd[mac] = StartEnd{Start: now, End: now}
	}
}
*/

// NOTE: Do not log MAC addresses.
func RecordMAC(mac string) {
	// FIXME MCJ 20220522 Should this be time.Local or time.UTC?
	now := GetClock().Now().In(time.UTC).Unix()
	// Check if we already have the MAC address in the ephemeral table.
	if p, ok := emd[mac]; ok {
		// Do not log MAC addresses.
		//log.Debug().Str("mac", mac).Msg("recordmac: exists, updating")
		// We track devices for the full day. We don't worry about "forgetting" them if they
		// go away for two hours. Keeps things easier.
		emd[mac] = StartEnd{Start: p.Start, End: now}

	} else {
		// We have never seen the MAC address.
		// Do not log MAC addresses.
		//log.Debug().Str("mac", mac).Msg("recordmac: new, inserting")
		emd[mac] = StartEnd{Start: now, End: now}
	}
}

type EphemeralDurationsDB map[int64][]structs.Duration

var edd EphemeralDurationsDB = make(EphemeralDurationsDB)

func StoreManyDurations(session_id int64, durations []structs.Duration) {
	log.Debug().
		Int("len", len(durations)).
		Int64("session_id", session_id).
		Msg("storing many durations")

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

func DumpEphemeralDurationsDB(session_id int64) {
	for _, d := range edd[session_id] {
		fmt.Println(d)
	}
}
