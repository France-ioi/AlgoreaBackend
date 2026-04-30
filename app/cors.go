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

// errCORSPermissiveOriginsWithCredentials is returned whenever the resolved
// allow-list would let go-chi/cors match every origin while
// "allowCredentials" is true. That covers two cases that are semantically
// identical inside the library:
//   - the bare "*" appears in "allowedOrigins"; or
//   - "allowedOrigins" is empty/unset -- go-chi/cors@v1.0.0 sets
//     allowedOriginsAll=true when len(AllowedOrigins)==0, so the
//     "credentials-only" config (only allowCredentials=true, e.g. set via the
//     ALGOREA_CORS__ALLOWCREDENTIALS env var) silently behaves like "*".
//
// In both cases the middleware would echo the caller's Origin back together
// with Access-Control-Allow-Credentials: true, the canonical CSRF-via-CORS
// misconfiguration. Per-host wildcards like "https://*.example.com" are not
// affected and remain allowed alongside credentials.
var errCORSPermissiveOriginsWithCredentials = errors.New(
	`cors: "allowCredentials"=true requires "allowedOrigins" to be an explicit non-empty list that does not contain "*" ` +
		`(go-chi/cors treats both an empty list and a bare "*" as "allow any origin", ` +
		`which combined with credentials is the canonical CSRF-via-CORS misconfiguration)`)

// resolveAllowedOrigins returns the configured CORS allow-list. The wildcard
// "*" is allowed (alone or mixed with explicit entries) and per-host wildcards
// like "https://*.example.com" are also supported by go-chi/cors. An unset key
// resolves to an empty slice in every environment: there is no env-based
// fallback. Note that go-chi/cors interprets an empty slice the same as
// ["*"]; the safety check in corsConfig closes that loophole when credentials
// are enabled.
func resolveAllowedOrigins(corsConf *viper.Viper) []string {
	return corsConf.GetStringSlice(allowedOriginsKey)
}

// corsConfig builds the CORS middleware from the "cors" subconfig:
//   - allowedOrigins ([]string, default []): see resolveAllowedOrigins.
//   - allowCredentials (bool, default false): whether to send
//     Access-Control-Allow-Credentials: true so cookies / Authorization
//     headers are usable cross-origin.
//
// When allowCredentials is true, an allowedOrigins list that go-chi/cors
// would treat as "any origin" (empty/unset, or containing the bare "*") is
// rejected with errCORSPermissiveOriginsWithCredentials so the application
// refuses to start instead of silently exposing every authenticated endpoint
// to any origin.
func corsConfig(corsConf *viper.Viper) (*cors.Cors, error) {
	allowedOrigins := resolveAllowedOrigins(corsConf)
	allowCredentials := corsConf.GetBool(allowCredentialsKey)

	if allowCredentials && (len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*")) {
		return nil, errCORSPermissiveOriginsWithCredentials
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
