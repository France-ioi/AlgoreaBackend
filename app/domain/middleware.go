package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Middleware is a middleware setting domain-specific configuration into the request context
func Middleware(domains []ConfigItem) func(next http.Handler) http.Handler {

	var defaultConfig *CtxConfig // the config that will be used (if defined) if no domain match

	domainsMap := map[string]*CtxConfig{}
	for _, domain := range domains {
		for _, host := range domain.Domains {
			domainsMap[host] = &CtxConfig{
				AllUsersGroupID:  domain.AllUsersGroup,
				TempUsersGroupID: domain.TempUsersGroup,
			}
			if host == "default" {
				defaultConfig = domainsMap[host]
			}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := strings.SplitN(r.Host, ":", 2)[0]
			configuration := domainsMap[domain]
			if configuration == nil {
				configuration = defaultConfig // if no match, set the default one (that can be nil as well)
			}
			if configuration == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusNotImplemented)
				data, _ := json.Marshal(struct {
					Success   bool   `json:"success"`
					Message   string `json:"message"`
					ErrorText string `json:"error_text"`
				}{Success: false, Message: "Not implemented", ErrorText: fmt.Sprintf("Wrong domain %q", domain)})
				_, _ = w.Write(data)
				return
			}
			ctx := context.WithValue(
				context.WithValue(r.Context(), ctxDomainConfig, configuration), ctxDomain, domain)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
