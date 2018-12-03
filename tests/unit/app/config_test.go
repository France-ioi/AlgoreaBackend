package app_test // nolint

import (
  "io/ioutil"
  "os"
  "testing"

  "github.com/France-ioi/AlgoreaBackend/app/config"
)

func TestLoadConfig(t *testing.T) {

  // create a temp config file
  tmpFile, err := ioutil.TempFile(os.TempDir(), "config-*.yaml")
  if err != nil {
    t.Error(err)
  }
  defer os.Remove(tmpFile.Name())
  defer tmpFile.Close()

  text := []byte("server:\n  port: 1234\n")
  if _, err = tmpFile.Write(text); err != nil {
    t.Error(err)
  }

  config.Path = tmpFile.Name()

  os.Setenv("ALGOREA_SERVER.WRITETIMEOUT", "999")
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
