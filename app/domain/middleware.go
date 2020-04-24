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

	domainsMap := map[string]*CtxConfig{}
	for _, domain := range domains {
		for _, host := range domain.Domains {
			domainsMap[host] = &CtxConfig{
				RootGroupID:     domain.RootGroup,
				RootSelfGroupID: domain.RootSelfGroup,
				RootTempGroupID: domain.RootTempGroup,
			}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := strings.SplitN(r.Host, ":", 2)[0]
			configuration := domainsMap[domain]
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
			ctx := context.WithValue(r.Context(), ctxDomainConfig, configuration)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
