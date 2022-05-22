package structs

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

type Entry struct {
	MAC   string
	Mfg   string
	Count int
}

type EphemeralDuration struct {
	Start int64  `db:"start"`
	End   int64  `db:"end"`
	MAC   string `db:"mac"`
}

type ByStart []Duration

func (a ByStart) Len() int { return len(a) }
func (a ByStart) Less(i, j int) bool {
	// it, _ := time.Parse(time.RFC3339, a[i].Start)
	// jt, _ := time.Parse(time.RFC3339, a[j].Start)
	return a[i].Start < a[j].Start
	//return it.Before(jt)
}
func (a ByStart) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type Durations struct {
	Data []Duration `json:"data"`
}

type Duration struct {
	ID        int    `json:"id"`
	PiSerial  string `json:"pi_serial"`
	SessionID string `json:"session_id"`
	FCFSSeqID string `json:"fcfs_seq_id"`
	DeviceTag string `json:"device_tag"`
	PatronID  int    `json:"patron_index"`
	Start     int64  `json:"start,string"`
	End       int64  `json:"end,string"`
}

func (d *Duration) AsMap() map[string]interface{} {
	e, err := json.Marshal(d)
	var js map[string]interface{}
	json.Unmarshal([]byte(e), &js)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("could not marshal duration struct")
	}
	return js
}
