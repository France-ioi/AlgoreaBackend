package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Middleware is a middleware setting domain-specific configuration into the request context.
func Middleware(domains []ConfigItem, domainOverride string) func(next http.Handler) http.Handler {
	var defaultConfig *CtxConfig // the config that will be used (if defined) if no domain match

	domainsMap := map[string]*CtxConfig{}
	for _, domain := range domains {
		for _, host := range domain.Domains {
			domainsMap[host] = &CtxConfig{
				AllUsersGroupID:     domain.AllUsersGroup,
				NonTempUsersGroupID: domain.NonTempUsersGroup,
				TempUsersGroupID:    domain.TempUsersGroup,
			}
			if host == "default" {
				defaultConfig = domainsMap[host]
			}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			domain := domainOverride
			if domain == "" {
				domain = strings.SplitN(httpRequest.Host, ":", 2)[0] //nolint:mnd // get the domain from the request host, ignoring the port if any
			}
			configuration := domainsMap[domain]
			if configuration == nil {
				configuration = defaultConfig // if no match, set the default one (that can be nil as well)
			}
			if configuration == nil {
				responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
				responseWriter.WriteHeader(http.StatusNotImplemented)
				data, _ := json.Marshal(struct {
					Success   bool   `json:"success"`
					Message   string `json:"message"`
					ErrorText string `json:"error_text"`
				}{Success: false, Message: "Not implemented", ErrorText: fmt.Sprintf("Wrong domain %q", domain)})
				_, _ = responseWriter.Write(data)
				return
			}
			ctx := context.WithValue(
				context.WithValue(httpRequest.Context(), ctxDomainConfig, configuration), ctxDomain, domain)
			next.ServeHTTP(responseWriter, httpRequest.WithContext(ctx))
		})
	}
}
