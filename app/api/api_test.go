package api

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func TestNewCtx_ReturnsErrorWhenReverseProxyServerIsInvalid(t *testing.T) {
	result, err := NewCtx(&config.Root{ReverseProxy: config.ReverseProxy{Server: "::::"}}, &database.DB{}, &token.Config{})
	assert.Nil(t, result)
	assert.Equal(t, &url.Error{Op: "parse", URL: "::::", Err: errors.New("missing protocol scheme")}, err)
}
