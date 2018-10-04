package app_test

import (
	"os"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/app"
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
	app.ConfigFile = "test_config.yaml"
	app.LoadConfig()

	// test config override
	if app.Config.Server.Port != 1234 {
		t.Error("invalid port from config file.")
	}

	// test default
	if app.Config.Server.ReadTimeout != 60 {
		t.Error("invalid default value")
	}

	// test env
	if app.Config.Server.WriteTimeout != 999 {
		t.Error("invalid env value")
	}

}
