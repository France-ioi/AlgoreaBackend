package app_test // nolint

import (
  "os"
  "testing"

  "github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestLoadConfig(t *testing.T) {

  os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
  config.Path = "test_config.yaml"
  conf, err := config.Load()
  if err != nil {
    t.Error("unable to load config")
  }

  // test config override
  if conf.Server.Port != 1234 {
    t.Error("invalid port from config file.")
  }

  // test default
  if conf.Server.ReadTimeout != 60 {
    t.Error("invalid default value")
  }

  // test env
  if conf.Server.WriteTimeout != 999 {
    t.Error("invalid env value")
  }

}
