package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFromContext(t *testing.T) {
	expectedConfig := &CtxConfig{AllUsersGroupID: 101, NonTempUsersGroupID: 102, TempUsersGroupID: 103}
	ctx := context.WithValue(context.Background(), ctxKeyDomainConfig, expectedConfig)
	conf := ConfigFromContext(ctx)

	assert.NotSame(t, expectedConfig, conf)
	assert.Equal(t, expectedConfig, conf)
}

func TestDomainFromContext(t *testing.T) {
	expectedDomain := "somedomain.com"
	ctx := context.WithValue(context.Background(), ctxKeyDomain, expectedDomain)
	domain := CurrentDomainFromContext(ctx)
	assert.Equal(t, expectedDomain, domain)
}
