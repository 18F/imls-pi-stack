package state

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func (suite *ConfigSuite) SetupTest() {
}

func (suite *ConfigSuite) AfterTest(suiteName, testName string) {
}

func (suite *ConfigSuite) TestConfigDefaults() {
	GetConfig()
	var expected = []string{"local:stderr", "local:tmp", "api:directus"}
	result := viper.GetStringSlice("logging.loggers")
	for i := 0; i < 3; i += 1 {
		if result[i] != expected[i] {
			suite.Fail("loggers were not equal")
		}
	}
}

func (suite *ConfigSuite) TestConfigWrite() {
	dc := GetConfig()
	dc.Set("device.deviceTag", "a random string")
	result := dc.Get("device.deviceTag")
	if result != "a random string" {
		suite.Fail("write was not reflected")
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
