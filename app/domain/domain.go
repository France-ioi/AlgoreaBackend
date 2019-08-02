package domain

import "context"

// ConfigFromContext retrieves the current domain configuration from a context set by the middleware
func ConfigFromContext(ctx context.Context) *Configuration {
	conf := ctx.Value(ctxDomainConfig).(*Configuration)
	confCopy := *conf
	return &confCopy
}
