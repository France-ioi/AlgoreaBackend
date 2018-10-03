package app_test

import (
	"os"
	"testing"

	"github.com/Flaque/filet"
	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	defer filet.CleanUp(t)

	filename := "/tmp/test_conf.yaml"
	filet.File(t, filename, "server:\n  port: 1234")
	os.Setenv("ALGOREA_SERVER.WRITE_TIMEOUT", "999")
	app.ConfigFile = filename
	app.LoadConfig()

	// test config override
	if viper.GetInt("server.port") != 1234 {
		t.Error("invalid port from config file.")
	}

	// test default
	if viper.GetInt("server.read_timeout") != 60 {
		t.Error("invalid default value")
	}

	// test env
	if viper.GetInt("server.write_timeout") != 999 {
		t.Error("invalid env value")
	}

}
