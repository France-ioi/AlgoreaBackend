package api

import (
	"testing"

	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func TestSetAuthConfig(t *testing.T) {
	assert := assertlib.New(t)
	authConfig := viper.New()
	authConfig.Set("foo", "bar")
	ctx := &Ctx{&service.Base{AuthConfig: authConfig}}
	assert.Equal("bar", ctx.service.AuthConfig.Get("foo"))

	newAuthConfig := viper.New()
	newAuthConfig.Set("foo", "bar2")
	ctx.SetAuthConfig(newAuthConfig)
	assert.Equal("bar2", ctx.service.AuthConfig.Get("foo"))
}

func TestSetDomainsConfig(t *testing.T) {
	assert := assertlib.New(t)
	domainConfig := []domain.ConfigItem{}
	ctx := &Ctx{&service.Base{DomainConfig: domainConfig}}
	assert.Len(ctx.service.DomainConfig, 0)

	newDomainConfig := []domain.ConfigItem{{}}
	ctx.SetDomainsConfig(newDomainConfig)
	assert.Len(ctx.service.DomainConfig, 1)
}
