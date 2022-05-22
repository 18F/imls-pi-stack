package state

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func (suite *ConfigSuite) SetupTest() {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config-pi")

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msg(viper.ConfigFileUsed())
	} else {
		panic("could not find config. exiting.")
	}
}

func (suite *ConfigSuite) AfterTest(suiteName, testName string) {
}

func (suite *ConfigSuite) TestConfigDefaults() {
	var expected = []string{"local:stderr", "local:tmp", "api:directus"}
	result := viper.GetStringSlice("logging.loggers")
	for i := 0; i < 3; i += 1 {
		if result[i] != expected[i] {
			suite.Fail("loggers were not equal")
		}
	}
}

func (suite *ConfigSuite) TestConfigWrite() {
	viper.Set("device.deviceTag", "a random string")
	result := viper.Get("device.deviceTag")
	if result != "a random string" {
		suite.Fail("write was not reflected")
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
