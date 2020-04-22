package api

import (
	"testing"

	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func TestSetAuthConfig(t *testing.T) {
	assert := assertlib.New(t)
	db, _ := database.NewDBMock()
	serverConfig := viper.New()
	authConfig := viper.New()
	authConfig.Set("foo", "bar")
	domainConfig := []domain.ConfigItem{}
	tokenConfig := &token.Config{}
	ctx := NewCtx(db, serverConfig, authConfig, domainConfig, tokenConfig)
	ctx.service = &service.Base{AuthConfig: authConfig}
	assert.Equal("bar", ctx.AuthConfig.Get("foo"))
	assert.Equal("bar", ctx.service.AuthConfig.Get("foo"))

	newAuthConfig := viper.New()
	newAuthConfig.Set("foo", "bar2")
	ctx.SetAuthConfig(newAuthConfig)
	assert.Equal("bar2", ctx.AuthConfig.Get("foo"))
	assert.Equal("bar2", ctx.service.AuthConfig.Get("foo"))
}
