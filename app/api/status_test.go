package api

import (
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func TestDbOk(t *testing.T) {
	assert := assertlib.New(t)
	ctx := &Ctx{&service.Base{}}
	assert.HTTPSuccess(ctx.status, "GET", "", nil)
	assert.HTTPBodyContains(ctx.status, "GET", "", nil, "The web service is responding! The database connection fails.")
}

func TestDbNotOk(t *testing.T) {
	assert := assertlib.New(t)
	dbMock, _ := database.NewDBMock()
	ctx := &Ctx{&service.Base{Store: database.NewDataStore(dbMock)}}
	assert.HTTPSuccess(ctx.status, "GET", "", nil)
	assert.HTTPBodyContains(ctx.status, "GET", "", nil, "The web service is responding! The database connection is established.")
}
