package main

import (
	appVersion "github.com/France-ioi/AlgoreaBackend/app/version"
	"github.com/France-ioi/AlgoreaBackend/cmd"
)

var (
	version = "unknown"
	_       = version
)

func main() {
	appVersion.Version = version

	cmd.Execute()
}
