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

// disallowedAllOriginsSentinel is the placeholder value injected into
// AllowedOrigins when the operator did not configure any origin (key unset
// or set to an empty list). It exists to short-circuit a footgun in
// go-chi/cors@v1.0.0: passing an empty AllowedOrigins slice flips the
// internal allowedOriginsAll flag, after which handleActualRequest echoes
// the caller's Origin header back on every response. That makes every
// public endpoint cross-origin readable from any site by default, even
// without credentials. By substituting a string that no real browser will
// ever send as Origin, we keep the slice non-empty (allowedOriginsAll
// stays false) and make the default fail-closed: isOriginAllowed returns
// false for every actual origin, so Access-Control-Allow-Origin is never
// set. Operators who want to expose endpoints cross-origin must add an
// explicit allowedOrigins list.
const disallowedAllOriginsSentinel = "none"

// errCORSCredentialsRequireExplicitOrigins is returned whenever
// "allowCredentials" is true but "allowedOrigins" cannot safely carry
// credentials. Two configurations trip it:
//   - the bare "*" appears in "allowedOrigins": go-chi/cors then echoes any
//     caller's Origin back together with Access-Control-Allow-Credentials:
//     true, the canonical CSRF-via-CORS misconfiguration; or
//   - "allowedOrigins" is unset/empty (so resolveAllowedOrigins substitutes
//     the disallowedAllOriginsSentinel): the runtime behavior is fail-closed
//     -- no real origin can match -- but combining "credentials enabled" with
//     "no usable trusted origins" is almost always an operator mistake (e.g.
//     setting only ALGOREA_CORS__ALLOWCREDENTIALS=true and forgetting the
//     companion list), so we surface it at startup rather than letting the
//     server boot into a permanently-broken credentialed setup.
//
// Per-host wildcards like "https://*.example.com" are not affected and remain
// allowed alongside credentials.
var errCORSCredentialsRequireExplicitOrigins = errors.New(
	`cors: "allowCredentials"=true requires "allowedOrigins" to be an explicit non-empty list of trusted origins that does not contain "*" ` +
		`(an unset/empty list is rewritten to the "none" sentinel and would deny every cross-origin request, ` +
		`while a bare "*" mixed with credentials is the canonical CSRF-via-CORS misconfiguration ` +
		`where go-chi/cors echoes the caller's Origin back)`)

// resolveAllowedOrigins returns the configured CORS allow-list, with one
// transformation: an unset/empty list is rewritten to
// []string{disallowedAllOriginsSentinel}. That keeps go-chi/cors out of its
// "allow any origin" branch (triggered by len(AllowedOrigins) == 0), so the
// out-of-the-box default is fail-closed for every request -- credentialed or
// not. Per-host wildcards like "https://*.example.com" are forwarded
// unchanged and remain matched normally by the library.
//
// The bare "*" is still accepted here (and simply forwarded), because the
// safety check in corsConfig is what enforces the rule that "*" cannot be
// combined with allowCredentials=true. Do NOT remove that check or the
// sentinel substitution without revisiting
// TestCORSConfig_DefaultBlocksAllOrigins,
// TestCORSConfig_RejectsWildcardWithCredentials,
// TestCORSConfig_RejectsWildcardMixedWithExplicit,
// TestCORSConfig_RejectsEmptyOriginsWithCredentials, and
// TestCORSConfig_RejectsMissingOriginsWithCredentials, which together lock
// in both halves of the policy.
func resolveAllowedOrigins(corsConf *viper.Viper) []string {
	origins := corsConf.GetStringSlice(allowedOriginsKey)
	if len(origins) == 0 {
		return []string{disallowedAllOriginsSentinel}
	}
	return origins
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

// isOnlyDisallowedSentinel reports whether the resolved list is exactly the
// fail-closed default produced by resolveAllowedOrigins for an unset/empty
// configuration. Used by corsConfig to surface the credentials-without-origins
// misconfiguration at startup instead of silently booting into a state where
// no cross-origin request can ever succeed.
func isOnlyDisallowedSentinel(origins []string) bool {
	return len(origins) == 1 && origins[0] == disallowedAllOriginsSentinel
}

// corsConfig builds the CORS middleware from the "cors" subconfig:
//   - allowedOrigins ([]string, default ["none"] sentinel): see
//     resolveAllowedOrigins. The sentinel default makes every cross-origin
//     request -- credentialed or not -- fail-closed until the operator
//     opts in by listing trusted origins.
//   - allowCredentials (bool, default false): whether to send
//     Access-Control-Allow-Credentials: true so cookies / Authorization
//     headers are usable cross-origin.
//
// When allowCredentials is true, an allowedOrigins list that is either the
// fail-closed sentinel or contains the bare "*" is rejected with
// errCORSCredentialsRequireExplicitOrigins so the application refuses to
// start instead of either silently exposing every authenticated endpoint to
// any origin or booting into a permanently-broken credentialed setup.
func corsConfig(corsConf *viper.Viper) (*cors.Cors, error) {
	allowedOrigins := resolveAllowedOrigins(corsConf)
	allowCredentials := corsConf.GetBool(allowCredentialsKey)

	if allowCredentials && (isOnlyDisallowedSentinel(allowedOrigins) || containsString(allowedOrigins, "*")) {
		return nil, errCORSCredentialsRequireExplicitOrigins
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
