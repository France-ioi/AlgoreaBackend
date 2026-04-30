package app

import (
	"errors"
	"slices"

	"github.com/go-chi/cors"
	"github.com/spf13/viper"
)

const corsMaxAge = 86400 // 24 hours in seconds

const (
	allowedOriginsKey   = "allowedOrigins"
	allowCredentialsKey = "allowCredentials"
)

// errCORSWildcardWithCredentials is returned when "allowedOrigins" contains the
// bare "*" while "allowCredentials" is true. This is the canonical
// CSRF-via-CORS misconfiguration: go-chi/cors echoes the caller's Origin back
// in Access-Control-Allow-Origin while still sending
// Access-Control-Allow-Credentials: true, letting any site read responses on
// behalf of a logged-in user. Per-host wildcards like
// "https://*.example.com" are not affected and remain allowed.
var errCORSWildcardWithCredentials = errors.New(
	`cors: "allowedOrigins" must not contain "*" when "allowCredentials" is true ` +
		`(canonical CSRF-via-CORS misconfiguration; use an explicit origin list)`)

// resolveAllowedOrigins returns the configured CORS allow-list. The wildcard
// "*" is allowed (alone or mixed with explicit entries) and per-host wildcards
// like "https://*.example.com" are also supported by go-chi/cors. An unset key
// resolves to an empty slice in every environment: there is no env-based
// fallback, so misconfigured deployments fail closed.
func resolveAllowedOrigins(corsConf *viper.Viper) []string {
	return corsConf.GetStringSlice(allowedOriginsKey)
}

// corsConfig builds the CORS middleware from the "cors" subconfig:
//   - allowedOrigins ([]string, default []): see resolveAllowedOrigins.
//   - allowCredentials (bool, default false): whether to send
//     Access-Control-Allow-Credentials: true so cookies / Authorization
//     headers are usable cross-origin.
//
// The combination "*" + allowCredentials=true is rejected with
// errCORSWildcardWithCredentials so the application refuses to start instead
// of silently exposing every authenticated endpoint to any origin.
func corsConfig(corsConf *viper.Viper) (*cors.Cors, error) {
	allowedOrigins := resolveAllowedOrigins(corsConf)
	allowCredentials := corsConf.GetBool(allowCredentialsKey)

	if allowCredentials && slices.Contains(allowedOrigins, "*") {
		return nil, errCORSWildcardWithCredentials
	}

	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Content-Encoding"},
		ExposedHeaders:   []string{"Date", "Backend-Version"},
		AllowCredentials: allowCredentials,
		MaxAge:           corsMaxAge,
	}), nil
}
