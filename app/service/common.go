package service

import (
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// Base is the common service context data
type Base struct {
	Store  *database.DataStore
	Config *config.Root
}
