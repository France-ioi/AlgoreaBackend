package app

import (
	"errors"

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

// resolveAllowedOrigins returns the configured CORS allow-list verbatim.
// Per-host wildcards like "https://*.example.com" are supported by go-chi/cors
// and are always safe. There is no env-based fallback -- an unset key
// resolves to an empty slice in every environment.
//
// In contrast, three configurations make go-chi/cors match every origin: the
// bare "*" alone, "*" mixed with explicit entries, and an empty/unset list
// (the library sets allowedOriginsAll=true when len(AllowedOrigins)==0). All
// three are accepted here -- the safety check in corsConfig rejects them at
// startup when allowCredentials is also true, so they remain the only shapes
// that can produce a fail-closed credentialed configuration. Do NOT remove
// the containsString / len() guards in corsConfig without thinking through
// that interaction (TestCORSConfig_RejectsWildcardWithCredentials,
// TestCORSConfig_RejectsWildcardMixedWithExplicit, and
// TestCORSConfig_RejectsEmptyOriginsWithCredentials lock the rule in).
func resolveAllowedOrigins(corsConf *viper.Viper) []string {
	return corsConf.GetStringSlice(allowedOriginsKey)
}

// containsString reports whether s is present in slice. Inlined instead of
// using the stdlib "slices" package so this file builds against pre-1.21 Go
// toolchains that some local environments still pin via GOROOT/GOTOOLCHAIN,
// even when go.mod declares 1.21.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
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

	if allowCredentials && (len(allowedOrigins) == 0 || containsString(allowedOrigins, "*")) {
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
