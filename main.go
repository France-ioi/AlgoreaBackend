/*
Package main - the entry point of the application
*/
package main

import (
	appVersion "github.com/France-ioi/AlgoreaBackend/v2/app/version"
	"github.com/France-ioi/AlgoreaBackend/v2/cmd"
)

var (
	version = "unknown"
	_       = version
)

func main() {
	appVersion.Version = version

	cmd.Execute()
}
