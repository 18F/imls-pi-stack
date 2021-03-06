package tlp

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/suite"
	"gsa.gov/18f/internal/interfaces"
	"gsa.gov/18f/internal/logwrapper"
	"gsa.gov/18f/internal/state"
	"gsa.gov/18f/internal/structs"
)

const PASS = true
const FAIL = false

type TLPSuite struct {
	suite.Suite
	mock *clock.Mock
	lw   *logwrapper.StandardLogger
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *TLPSuite) SetupTest() {
	tempDB, err := os.CreateTemp("", "tlp-test.sqlite")
	if err != nil {
		suite.Fail(err.Error())
	}
	state.SetConfigAtPath(tempDB.Name())

	cfg := state.GetConfig()
	cfg.SetRunMode("test")
	cfg.SetStorageMode("sqlite")

	suite.lw = logwrapper.NewLogger(cfg)
	suite.lw.SetLogLevel("DEBUG")

	mock := clock.NewMock()
	suite.mock = mock
	if mock == nil {
		suite.Fail("mock is nil")
	}
	mt, _ := time.Parse("2006-01-02T15:04", "1975-10-11T18:00")
	mock.Set(mt)
	state.SetClock(mock)
	suite.lw.Debug("mock is now ", state.GetClock().Now().In(time.Local))

	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Dir(filename)
	queuePath := filepath.Join(path, "..", "test", "queue.sqlite")
	cfg.SetQueuesPath(queuePath)

	rootPath := filepath.Join(path, "..", "test", "www")
	cfg.SetRootPath(rootPath)
	imagesPath := filepath.Join(path, "..", "test", "www", "images")
	cfg.SetImagesPath(imagesPath)

	os.Mkdir(cfg.GetWWWRoot(), 0755)
}

func (suite *TLPSuite) AfterTest(suiteName, testName string) {
	dc := state.GetConfig()
	// ensure a clean run.
	os.Remove(dc.GetDatabasePath())
	dc.Close()
}

func generateFakeMac() string {
	var letterRunes = []rune("ABCDEF0123456789")
	b := make([]rune, 17)
	colons := [...]int{2, 5, 8, 11, 14}
	for i := 0; i < 17; i++ {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]

		for v := range colons {
			if i == colons[v] {
				b[i] = rune(':')
			}
		}
	}
	return string(b)
}

func RunFakeWireshark(ka *Keepalive, kb *KillBroker, in <-chan bool, out chan []string) {
	NUMMACS := 200
	NUMRANDOM := 10
	lw := logwrapper.NewLogger(nil)
	lw.Debug("RunFakeWireshark in the house.")

	chKill := kb.Subscribe()
	// Lets have 30 consistent devices
	macs := make([]string, NUMMACS)
	for i := 0; i < NUMMACS; i++ {
		macs[i] = generateFakeMac()
	}

	for {
		select {
		case <-in:
			// Pick NUMRANDOM devices every minute
			send := make([]string, NUMRANDOM)
			for i := 0; i < NUMRANDOM; i++ {
				send[i] = macs[rand.Intn(len(macs))]
			}
			out <- send

		case <-chKill:
			log.Println("Exiting RunFakeWireshark")
			return
		}
	}
}

func PingAtBogoMidnight(ka *Keepalive,
	rb *ResetBroker,
	kb *KillBroker,
	m *clock.Mock) {
	// counter := 0
	// chKill := kb.Subscribe()
	lw := logwrapper.NewLogger(nil)
	pinged := false
	for {
		if m.Now().Hour() == 0 && !pinged {
			pinged = true
			lw.Debug("IT IS BOGOMIDNIGHT.")
			rb.Publish(Ping{})
		}
		if m.Now().Hour() != 0 {
			pinged = false
		}
	}
}

func (suite *TLPSuite) TestManyTLPCycles() {

	// Create channels for process network
	chSec := make(chan bool)

	chNsec := make(chan bool)
	chMacs := make(chan []string)
	chMacsCounted := make(chan map[string]int)
	chDataForReport := make(chan []structs.WifiEvent)

	chWifiDB := make(chan interfaces.Database)
	chDurationsDB := make(chan interfaces.Database)
	chDdbPar := make([]chan interfaces.Database, 2)

	chAck := make(chan Ping)
	for i := 0; i < 2; i++ {
		chDdbPar[i] = make(chan interfaces.Database)
	}

	resetbroker := NewResetBroker()
	go resetbroker.Start()
	killbroker := NewKillBroker()
	go killbroker.Start()

	// Tock every two seconds.
	go TockEveryN(nil, killbroker, 2, chSec, chNsec)

	// Need a fake RunWireshark
	// go tlp.RunWireshark(nil, cfg, ch_nsec1, ch_macs, KC[1])
	go RunFakeWireshark(nil, killbroker, chNsec, chMacs)

	go AlgorithmTwo(nil, resetbroker, killbroker, chMacs, chMacsCounted)
	go PrepEphemeralWifi(nil, killbroker, chMacsCounted, chDataForReport)
	// At midnight, flush internal structures and restart.
	//go tlp.PingAtMidnight(nil, cfg, chs_reset[0], KC[4])
	go PingAtBogoMidnight(nil, resetbroker, killbroker, suite.mock)
	go CacheWifi(nil, resetbroker, killbroker, chDataForReport, chWifiDB, chAck)
	// Make sure we don't hang...
	go GenerateDurations(nil, killbroker, chWifiDB, chDurationsDB, chAck)

	go ParDeltaTempDB(killbroker, chDurationsDB, chDdbPar...)
	go BatchSend(nil, killbroker, chDdbPar[0])
	go WriteImages(nil, killbroker, chDdbPar[1])

	// See if we can wait and shut down the test...
	var wg sync.WaitGroup
	wg.Add(1)

	NUMCYCLESTORUN := 400

	go func() {
		minutes := 0
		skip := 20
		m, _ := time.ParseDuration(fmt.Sprintf("%vm", skip))
		for secs := 0; secs < NUMCYCLESTORUN; secs++ {
			chSec <- true
			if secs%2 == 0 {
				suite.mock.Add(m)
				suite.lw.Debug("MOCK NOW ", suite.mock.Now())
				// var m runtime.MemStats
				// runtime.ReadMemStats(&m)
				minutes += skip
				// memstats := fmt.Sprintf("Alloc[%vMB] Sys[%vMB], NumGC[%v]", bToMb(m.Alloc), bToMb(m.Sys), m.NumGC)
				// log.Println(days, "d", hours, "h", minutes%60, "m", memstats)
			}
		}
		log.Println("Killing the test network.")
		killbroker.Publish(Ping{})
		wg.Done()
	}()

	wg.Wait()
}

func macs(arr ...string) []string {
	// h := make(map[string]int)
	// for _, s := range arr {
	// 	h[s] = rand.Intn(1024)
	// }
	// return h
	return arr
}

func hashes(arr ...string) [][]string {
	// Return a list of hashes, one hash for each string
	harr := make([][]string, 0)
	for _, s := range arr {
		harr = append(harr, []string{s})
	}
	return harr
}

// IDs will be assigned in the raw_to_uids proc
// on a sorted list of MAC addrs. Therefore, we *know*
// in this test set that "next" will always be UID 0.
// (If "next" and "apple" are together)
var m = map[string]string{
	"next":     "00:00:0f:aa:bb:cc", // ID 0
	"ericsson": "00:01:ec:aa:bb:cc",
	"apple":    "00:03:93:aa:bb:cc",
	"next2":    "00:00:0f:ee:ff:00",
}

var tests = []struct {
	description      string
	passfail         bool
	uniquenessWindow int
	initMap          []string
	loopMaps         [][]string
	resultMap        map[string]int
}{
	// One input hash.
	{"one input mac, one loop mac",
		PASS, 5,
		macs(m["next"]),
		hashes(m["next"]),
		map[string]int{
			"0:0": 0,
		},
	},
	// // Two input hashes
	{"two input macs, one loop mac",
		PASS, 5,
		macs(m["next"], m["apple"]),
		hashes(m["next"]),
		// Why zero and one?
		// Zero for deadbeef, because we send it in the loop.
		// One for beefcafe, because it was only sent once, and
		// one tick goes by.
		map[string]int{
			"0:0": 0,
			"1:1": 1,
		},
	},
	// // Two input hashes
	{"two input macs, one loop mac, both next",
		PASS, 5,
		macs(m["next"], m["next2"]),
		hashes(m["next"]),
		// Why zero and one?
		// Zero for deadbeef, because we send it in the loop.
		// One for beefcafe, because it was only sent once, and
		// one tick goes by.
		map[string]int{
			"0:0": 0,
			"0:1": 1,
		},
	},
	// Three hashes, three minutes
	{"three input macs, three comms in the middle",
		PASS, 5,
		// Next, Apple, Ericsson
		macs(m["next"], m["apple"], m["ericsson"]),
		hashes("de:ad:be:ef", "de:ad:be:ef", "de:ad:be:ef"),
		// IDs will be assigned by MAC address sort!
		map[string]int{
			"0:0": 3,
			"1:1": 3,
			"2:2": 3,
			"3:3": 0,
		},
	},

	// Next times out, because it is considered to
	// have "disconnected" after 5 minutes.
	{"Next should disappear",
		PASS, 5,
		macs(m["next"], m["apple"], m["ericsson"]),
		hashes(
			"de:ad:be:ef",
			"de:ad:be:ef",
			"de:ad:be:ef",
			m["apple"],
			m["ericsson"]),
		// Why zero and one?
		// Zero for deadbeef, because we send it in the loop.
		// One for beefcafe, because it was only sent once, and
		// one tick goes by.
		map[string]int{
			"1:1": 1,
			"2:2": 0,
			"3:3": 2,
		},
	},

	// Next times out, comes back. Still ID 0.
	// Apple is considered to have disconnected.
	{"Drop two",
		PASS, 5,
		macs(m["next"], m["apple"], m["ericsson"]),
		hashes(
			"de:ad:be:ef",
			"de:ad:be:ef",
			"de:ad:be:ef",
			"de:ad:be:ef",
			m["next"]),
		// Why zero and one?
		// Zero for deadbeef, because we send it in the loop.
		// One for beefcafe, because it was only sent once, and
		// one tick goes by.
		map[string]int{
			"0:0": 0,
			"3:3": 1,
		},
	},
}

func assertEqual(a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	log.Fatal(message, "\n\texpected: ", a, "\n\treceived: ", b)
}

func (suite *TLPSuite) RawToUid() {
	log.Println("TestRawToUid")
	cfg := state.GetConfig()
	ka := NewKeepalive()

	for testNdx, e := range tests {
		log.Printf("Test #%v: %v\n", testNdx, e.description)
		cfg.SetUniquenessWindow(e.uniquenessWindow)

		var wg sync.WaitGroup
		resetbroker := NewResetBroker()
		go resetbroker.Start()
		// The kill broker lets us poison the network.
		// var killbroker *tlp.Broker = nil
		killbroker := NewKillBroker()
		go killbroker.Start()

		chMacs := make(chan []string)
		chUniq := make(chan map[string]int)
		var u map[string]int = nil

		wg.Add(1)
		go func() {
			chMacs <- e.initMap
			for _, sarr := range e.loopMaps {

				chMacs <- sarr
			}
			defer wg.Done()
		}()

		// Not using the reset here.
		go AlgorithmTwo(ka, resetbroker, killbroker, chMacs, chUniq)

		wg.Add(1)
		go func() {
			// The init map
			<-chUniq
			count := len(e.loopMaps) - 1
			for i := 0; i < count; i++ {
				// This reads in the intervening maps.
				<-chUniq
			}
			u = <-chUniq
			killbroker.Publish(Ping{})
			defer wg.Done()
		}()

		wg.Wait()

		// The last value we receive needs to have its time updated.
		expected := fmt.Sprint(e.resultMap)
		received := fmt.Sprint(u)
		//log.Println("expected", expected, "received", received)

		if e.passfail {
			assertEqual(expected, received, "not equal")
		}
	} // end for over tests
}

func TestTLPSuite(t *testing.T) {
	suite.Run(t, new(TLPSuite))
}
