package app

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

func TestNew_Success(t *testing.T) {
	assert := assertlib.New(t)
	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)
	assert.NotNil(app.Config)
	assert.NotNil(app.Database)
	assert.NotNil(app.HTTPHandler)
	assert.Len(app.HTTPHandler.Middlewares(), 8)
	assert.True(len(app.HTTPHandler.Routes()) > 0)
}

func TestNew_ConfigErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(config.Load, func() (*config.Root, error) {
		return nil, errors.New("config loading error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "config loading error")
}

func TestNew_DBErr(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	patch := monkey.Patch(database.Open, func(interface{}) (*database.DB, error) {
		return nil, errors.New("db opening error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)
	logMsg := hook.LastEntry()
	assert.Equal(logrus.ErrorLevel, logMsg.Level)
	assert.Equal("db opening error", logMsg.Message)
	assert.Equal("database", logMsg.Data["module"])
}

func TestNew_APIErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(api.NewCtx, func(*config.Root, *database.DB) (*api.Ctx, error) {
		return nil, errors.New("api creation error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "api creation error")
}
