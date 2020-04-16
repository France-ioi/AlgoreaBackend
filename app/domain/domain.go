package domain

import (
	"context"
)

type ctxKey int

const (
	ctxDomainConfig ctxKey = iota
)

// Configuration contains domain-specific settings
type Configuration struct {
	RootGroupID     int64
	RootSelfGroupID int64
	RootTempGroupID int64
}

// AppConfigItem is one item of the configuration list
type AppConfigItem struct {
	Domains       []string
	RootGroup     int64
	RootSelfGroup int64
	RootTempGroup int64
}

// ConfigFromContext retrieves the current domain configuration from a context set by the middleware
func ConfigFromContext(ctx context.Context) *Configuration {
	conf := ctx.Value(ctxDomainConfig).(*Configuration)
	confCopy := *conf
	return &confCopy
}
