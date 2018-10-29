package app_test

import (
	"os"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestLoadConfig(t *testing.T) {

	os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
	config.Path = "test_config.yaml"
	config, err := config.Load()
	if err != nil {
		t.Error("unable to load config")
	}

	// test config override
	if config.Server.Port != 1234 {
		t.Error("invalid port from config file.")
	}

	// test default
	if config.Server.ReadTimeout != 60 {
		t.Error("invalid default value")
	}

	// test env
	if config.Server.WriteTimeout != 999 {
		t.Error("invalid env value")
	}

}
