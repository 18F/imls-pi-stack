package structs

import (
	"testing"
	"time"
)

func TestAsMapDuration(t *testing.T) {
	e := Duration{
		// ID:        1,
		PiSerial:  "asdf",
		DeviceTag: "abd-dc",
		Start:     time.Now().Unix(),
		End:       time.Now().Unix(),
		SessionID: "hello",
		PatronID:  0,
	}
	e.PiSerial = "bcdf"
	// FIXME: MCJ 20220522 Need a new way to jsonify data.
	// m := e.AsMap()
	// if v, ok := m["ID"]; ok {
	// 	t.Log("map should not have `id` in it", v)
	// 	t.Fail()
	// }
}
