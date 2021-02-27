package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFromContext(t *testing.T) {
	expectedConfig := &CtxConfig{AllUsersGroupID: 101, TempUsersGroupID: 103}
	ctx := context.WithValue(context.Background(), ctxDomainConfig, expectedConfig)
	conf := ConfigFromContext(ctx)

	assert.False(t, expectedConfig == conf)
	assert.EqualValues(t, expectedConfig, conf)
}

func TestDomainFromContext(t *testing.T) {
	expectedDomain := "somedomain.com"
	ctx := context.WithValue(context.Background(), ctxDomain, expectedDomain)
	domain := CurrentDomainFromContext(ctx)
	assert.Equal(t, expectedDomain, domain)
}
