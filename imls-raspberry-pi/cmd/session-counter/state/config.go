package state

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gsa.gov/18f/cmd/session-counter/interfaces"
	"gsa.gov/18f/cmd/session-counter/structs"
)

func GetSerial() string {
	// allow the serial to be stored so we can test out different serial and api
	// key settings. if not, default to reading from /proc (as a cached read)
	serial := viper.GetString("device.serial")
	if serial == "DEFAULTSERIAL" {
		return getCachedSerial()
	}
	return serial
}

// func (dc *databaseConfig) GetFCFSSeqID() string {
// 	return dc.config.GetTextField("fcfs_seq_id")
// }

// func (dc *databaseConfig) GetDeviceTag() string {
// 	return dc.config.GetTextField("device_tag")
// }

// // GetAPIKey decodes the api key stored in the database.
// func (dc *databaseConfig) GetAPIKey() string {
// 	apiKey := dc.config.GetTextField("api_key")
// 	serial := dc.GetSerial()
// 	var key [32]byte
// 	copy(key[:], serial)
// 	b64, err := base64.StdEncoding.DecodeString(apiKey)
// 	if err != nil {
// 		log.Print("config: cannot b64 decode auth token: ", err)
// 	}
// 	dec, err := cryptopasta.Decrypt(b64, &key)
// 	if err != nil {
// 		log.Print("config: failed to decrypt auth token after decoding: ", err)
// 	}
// 	return string(dec)
// }

// func (dc *databaseConfig) SetFCFSSeqID(id string) {
// 	dc.config.SetTextField("fcfs_seq_id", id)
// }

// func (dc *databaseConfig) SetDeviceTag(tag string) {
// 	dc.config.SetTextField("device_tag", tag)
// }

// func (dc *databaseConfig) SetAPIKey(key string) {
// 	dc.config.SetTextField("api_key", key)
// }

// func (dc *databaseConfig) SetStorageMode(mode string) {
// 	dc.config.SetTextField("storage_mode", mode)
// }

// func (dc *databaseConfig) SetRunMode(mode string) {
// 	dc.config.SetTextField("run_mode", mode)
// }

// func (dc *databaseConfig) SetQueuesPath(mode string) {
// 	dc.config.SetTextField("queues_path", mode)
// }

// func (dc *databaseConfig) SetDurationsPath(mode string) {
// 	dc.config.SetTextField("durations_path", mode)
// }

// func (dc *databaseConfig) SetRootPath(mode string) {
// 	dc.config.SetTextField("www_root", mode)
// }

// func (dc *databaseConfig) SetImagesPath(mode string) {
// 	dc.config.SetTextField("www_images", mode)
// }

// func (dc *databaseConfig) SetUniquenessWindow(window int) {
// 	dc.config.SetIntegerField("uniqueness_window", window)
// }

// func (dc *databaseConfig) GetLogLevel() string {
// 	return dc.config.GetTextField("log_level")
// }

// func (dc *databaseConfig) GetLoggers() []string {
// 	loggers := dc.config.GetTextField("loggers")
// 	return strings.Split(loggers, ",")
// }

// func (dc *databaseConfig) Log() interfaces.Logger {
// 	return dc.logger
// }

// func (dc *databaseConfig) GetEventsURI() string {
// 	scheme := dc.config.GetTextField("umbrella_scheme")
// 	host := dc.config.GetTextField("umbrella_host")
// 	path := dc.config.GetTextField("events_uri")
// 	return (scheme + "://" +
// 		removeLeadingAndTrailingSlashes(host) +
// 		startsWithSlash(removeLeadingSlashes(path)))
// }

// func (dc *databaseConfig) GetDurationsURI() string {
// 	scheme := dc.config.GetTextField("umbrella_scheme")
// 	host := dc.config.GetTextField("umbrella_host")
// 	path := dc.config.GetTextField("durations_uri")
// 	return (scheme + "://" +
// 		removeLeadingAndTrailingSlashes(host) +
// 		startsWithSlash(removeLeadingSlashes(path)))
// }

func NewSessionId() int64 {
	viper.Set("sessionId", GetClock().Now().In(time.Local).Unix())
	return GetCurrentSessionId()
}

func GetCurrentSessionId() int64 {
	return int64(viper.GetInt("sessionId"))
}

func IncrementSessionId() int64 {
	NewSessionId()
	return GetCurrentSessionId()
}

func IsStoringToAPI() bool {
	mode := viper.GetString("storage.mode")
	return strings.Contains(strings.ToLower(mode), "api")
}

func IsStoringLocally() bool {
	mode := viper.GetString("storage.mode")
	either := false
	for _, s := range []string{"local", "sqlite"} {
		either = either || strings.Contains(strings.ToLower(mode), s)
	}
	return either
}

func IsProductionMode() bool {
	mode := viper.GetString("storage.runMode")
	return strings.Contains(strings.ToLower(mode), "prod")
}

func IsDeveloperMode() bool {
	mode := viper.GetString("storage.runMode")
	either := false
	for _, s := range []string{"dev", "test"} {
		either = either || strings.Contains(strings.ToLower(mode), s)
	}
	if either {
		log.Println("running in developer mode")
	}
	return either
}

// func (dc *databaseConfig) IsTestMode() bool {
// 	mode := dc.config.GetTextField("run_mode")
// 	return strings.Contains(strings.ToLower(mode), "test")
// }

func GetDurationsDatabase() interfaces.Database {
	// path := dc.config.GetTextField("durations_path")
	path := viper.GetString("paths.durations")
	// always make sure we have a durations db created
	db := NewSqliteDB(path)
	if !db.CheckTableExists("durations") {
		db.CreateTableFromStruct(structs.Duration{})
	}
	return db
}

// func (dc *databaseConfig) GetDatabasePath() string {
// 	return dc.configDB.GetPath()
// }

// func (dc *databaseConfig) GetQueuesDatabase() interfaces.Database {
// 	path := dc.config.GetTextField("queues_path")
// 	return NewSqliteDB(path)
// }

// func (dc *databaseConfig) GetWiresharkPath() string {
// 	return dc.config.GetTextField("wireshark_path")
// }

// func (dc *databaseConfig) GetWiresharkDuration() int {
// 	return dc.config.GetIntegerField("wireshark_duration")
// }

// func (dc *databaseConfig) GetMinimumMinutes() int {
// 	return dc.config.GetIntegerField("minimum_minutes")
// }

// func (dc *databaseConfig) GetMaximumMinutes() int {
// 	return dc.config.GetIntegerField("maximum_minutes")
// }

// func (dc *databaseConfig) GetUniquenessWindow() int {
// 	return dc.config.GetIntegerField("uniqueness_window")
// }

// func (dc *databaseConfig) GetResetCron() string {
// 	return dc.config.GetTextField("reset_cron")
// }

// func (dc *databaseConfig) GetHeartbeatCron() string {
// 	return dc.config.GetTextField("heartbeat_cron")
// }

// func (dc *databaseConfig) GetWWWRoot() string {
// 	return dc.config.GetTextField("www_root")
// }

// func (dc *databaseConfig) GetWWWImages() string {
// 	return dc.config.GetTextField("www_images")
// }

// func (dc *databaseConfig) Close() {
// 	dc.configDB.Close()
// }

// type ConfigDB struct {
// 	logLevel          string `db:"log_level" sqlite:"TEXT"`
// 	loggers           string `db:"loggers" sqlite:"TEXT"` // comma separated
// 	apiKey            string `db:"api_key" sqlite:"TEXT"`
// 	deviceTag         string `db:"device_tag" sqlite:"TEXT"`
// 	fcfsSeqID         string `db:"fcfs_seq_id" sqlite:"TEXT"`
// 	heartbeatCron     string `db:"heartbeat_cron sqlite:"TEXT"`
// 	serial            string `db:"serial" sqlite:"TEXT"`
// 	storageMode       string `db:"storage_mode" sqlite:"TEXT"`
// 	runMode           string `db:"run_mode" sqlite:"TEXT"`
// 	durationsPath     string `db:"durations_path" sqlite:"TEXT"`
// 	queuesPath        string `db:"queues_path" sqlite:"TEXT"`
// 	umbrellaScheme    string `db:"umbrella_scheme" sqlite:"TEXT"`
// 	umbrellaHost      string `db:"umbrella_host" sqlite:"TEXT"`
// 	eventsURI         string `db:"events_uri" sqlite:"TEXT"`
// 	durationsURI      string `db:"durations_uri" sqlite:"TEXT"`
// 	minimumMinutes    int    `db:"minimum_minutes" sqlite:"INTEGER"`
// 	maximumMinutes    int    `db:"maximum_minutes" sqlite:"INTEGER"`
// 	uniquenessWindow  int    `db:"uniqueness_window" sqlite:"INTEGER"`
// 	wiresharkPath     string `db:"wireshark_path" sqlite:"TEXT"`
// 	wiresharkDuration int    `db:"wireshark_duration" sqlite:"INTEGER"`
// 	resetCron         string `db:"reset_cron" sqlite:"TEXT"`
// 	wwwRoot           string `db:"www_root" sqlite:"TEXT"`
// 	wwwImages         string `db:"www_images" sqlite:"TEXT"`
// }

// func ConfigDefaults() ConfigDB {
// 	var defaults ConfigDB
// 	defaults.logLevel = "DEBUG"
// 	defaults.loggers = "local:stderr,local:tmp,api:directus"
// 	// APIKey filled in by user
// 	// DeviceTag filled in by user
// 	// FCFSSeqID filled in by user
// 	// Serial filled in by device or user
// 	defaults.storageMode = "api"
// 	defaults.runMode = "prod"
// 	defaults.umbrellaScheme = "https"
// 	defaults.umbrellaHost = "api.data.gov"
// 	defaults.eventsURI = "/TEST/10x-imls/v2/events/"
// 	defaults.durationsURI = "/TEST/10x-imls/v2/durations/"
// 	defaults.minimumMinutes = 5
// 	defaults.maximumMinutes = 600
// 	defaults.wiresharkDuration = 45
// 	defaults.resetCron = "0 0 * * *"
// 	defaults.heartbeatCron = "*/5 * * * *"
// 	if runtime.GOOS == "windows" {
// 		defaults.wiresharkPath = "c:/Program Files/Wireshark/bin/tshark.exe"
// 		defaults.wwwRoot = "c:/imls"
// 		defaults.wwwImages = "c:/imls/images"
// 		defaults.durationsPath = "c:/imls/durations.sqlite"
// 		defaults.queuesPath = "c:/imls/queues.sqlite"
// 	} else {
// 		defaults.wiresharkPath = "/usr/bin/tshark"
// 		defaults.wwwRoot = "/www/imls"
// 		defaults.wwwImages = "/www/imls/images"
// 		defaults.durationsPath = "/www/imls/durations.sqlite"
// 		defaults.queuesPath = "/www/imls/queues.sqlite"
// 	}
// 	return defaults
// }