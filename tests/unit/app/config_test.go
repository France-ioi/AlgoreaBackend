package app_test

import (
	"os"
	"testing"

	"github.com/Flaque/filet"
	"github.com/France-ioi/AlgoreaBackend/app"
)

func TestLoadConfig(t *testing.T) {
	defer filet.CleanUp(t)

	filename := "/tmp/test_conf.yaml"
	filet.File(t, filename, "server:\n  port: 1234")
	os.Setenv("ALGOREA_SERVER.WRITE_TIMEOUT", "999")
	app.ConfigFile = filename
	app.LoadConfig()

	// test config override
	if app.Config.GetInt("server.port") != 1234 {
		t.Error("invalid port from config file.")
	}

	// test default
	if app.Config.GetInt("server.read_timeout") != 60 {
		t.Error("invalid default value")
	}

	// test env
	if app.Config.GetInt("server.write_timeout") != 999 {
		t.Error("invalid env value")
	}

}
