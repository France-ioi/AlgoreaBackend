package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/cors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

// dispatchCORSRequest drives the CORS middleware end-to-end and returns the
// response headers together with whether the wrapped handler was reached.
// Going through Handler() is the only way to observe what go-chi/cors actually
// does for a given (origin, allow-list, credentials) combination, which is
// what the security guarantees in corsConfig hinge on -- mere "non-nil
// returned" assertions cannot catch a regression where, say, an empty
// AllowedOrigins silently turns into "allow any origin".
func dispatchCORSRequest(t *testing.T, c *cors.Cors, method, origin string) (http.Header, bool) {
	t.Helper()

	nextCalled := false
	handler := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(method, "https://api.example.com/", http.NoBody)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if method == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr.Header(), nextCalled
}

func TestResolveAllowedOrigins_ExplicitListIsForwarded(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "https://algorea.org"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"https://www.france-ioi.org", "https://algorea.org"}, got)
}

func TestResolveAllowedOrigins_MissingKeyDefaultsToDisallowedSentinel(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Unset allowedOrigins must NOT pass through as an empty slice -- that
	// would flip go-chi/cors into "allow any origin" mode (see
	// disallowedAllOriginsSentinel for the rationale). Instead we expect
	// the sentinel substitution that keeps the runtime behavior
	// fail-closed for every cross-origin request.
	corsConf := viper.New()

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{disallowedAllOriginsSentinel}, got)
}

func TestResolveAllowedOrigins_ExplicitEmptyListResolvesToDisallowedSentinel(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Explicitly setting allowedOrigins to [] must behave the same as
	// leaving the key unset: both resolve to the sentinel, otherwise an
	// "I want CORS off" intent expressed as `allowedOrigins: []` would
	// silently turn into the dangerous "allow any origin" branch.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{disallowedAllOriginsSentinel}, got)
}

func TestResolveAllowedOrigins_WildcardEntryIsPassedThrough(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"*"}, got)
}

func TestResolveAllowedOrigins_SplitsCommaSeparatedEntry(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Operator-facing escape hatch: a comma-separated env var (which viper
	// hands us as a single comma-glued slice element on the v1.3.1 / cast
	// v1.3.0 codepath) must be split into the two origins the operator
	// intended. Without this, the resolved list would contain one
	// nonsensical "https://a.example,https://b.example" entry that no
	// browser ever matches AND -- because len > 0 -- would NOT trigger
	// the disallowedAllOriginsSentinel substitution, silently bypassing
	// the credentials-without-origins startup check too.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://a.example,https://b.example"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, got)
}

func TestResolveAllowedOrigins_TrimsWhitespaceAndDropsEmpty(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Trailing commas, surrounding whitespace, and lone empty entries
	// (e.g. from `allowedOrigins: ["", "https://a.example "]` in YAML or
	// `https://a.example,` in an env var) must not survive into
	// AllowedOrigins. A surviving "" entry would otherwise be a literal
	// allowed origin in go-chi/cors while still keeping len > 0, which
	// would defeat the sentinel substitution below.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"  ", " https://a.example , ", "https://b.example,"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, got)
}

func TestResolveAllowedOrigins_OnlyEmptyStringsResolveToDisallowedSentinel(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// `allowedOrigins: [""]` (or any list of exclusively empty / whitespace
	// entries) must collapse to the same fail-closed default as an unset
	// list. Pre-fix this slipped through as a `cors.Cors` whose only
	// "explicit" origin was the empty string -- runtime never matched it,
	// but the credentials-without-origins startup check did NOT fire
	// because len(AllowedOrigins) was 1, so an operator could boot a
	// credentialed setup with effectively no trusted origins.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{""})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{disallowedAllOriginsSentinel}, got)
}

func TestCORSConfig_ReturnsCORSHandler(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)
	assert.NotNil(t, corsHandler.Handler)
}

func TestCORSConfig_ExplicitListEchoesAllowedOrigin(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "https://algorea.org"})

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	headers, nextCalled := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://algorea.org")
	assert.True(t, nextCalled, "wrapped handler must run for an allowed actual request")
	assert.Equal(t, "https://algorea.org", headers.Get("Access-Control-Allow-Origin"))
}

func TestCORSConfig_ExplicitListRejectsUnknownOrigin(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "https://algorea.org"})

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	// Unknown origin must not get Access-Control-Allow-Origin echoed back.
	// This is the assertion that pins the explicit allow-list contract:
	// a future go-chi/cors upgrade or config refactor that turns the
	// resolved list into "allow any" would fail here instead of silently
	// loosening the policy.
	headers, nextCalled := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://attacker.example")
	assert.True(t, nextCalled, "wrapped handler still runs; CORS just withholds headers")
	assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, headers.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_AllowCredentialsDefaultsToFalse(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})

	// Default allowCredentials is false, so combining the bare "*" with the
	// default credentials setting must not trip the safety rule.
	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	// And -- importantly -- no Access-Control-Allow-Credentials header is
	// emitted at runtime: the absence of the header is the contract that
	// keeps wildcard origins safe.
	headers, _ := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://anything.example")
	assert.Equal(t, "https://anything.example", headers.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, headers.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_AllowCredentialsTrueIsForwarded(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org"})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	// AllowCredentials=true must reach the underlying middleware -- assert
	// on the response headers rather than on the constructed *cors.Cors,
	// whose credential flag is unexported.
	headers, _ := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://www.france-ioi.org")
	assert.Equal(t, "https://www.france-ioi.org", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_RejectsWildcardWithCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSCredentialsRequireExplicitOrigins)
	assert.Nil(t, corsHandler)
}

func TestCORSConfig_AllowsWildcardWithoutCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})
	corsConf.Set(allowCredentialsKey, false)

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	// Bare "*" without credentials must echo back any origin (that is the
	// whole point of allowing it) but must never carry credentials.
	headers, _ := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://random.example")
	assert.Equal(t, "https://random.example", headers.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, headers.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_RejectsWildcardMixedWithExplicit(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Even when the bare "*" is mixed with explicit origins, the unsafe
	// echo-the-Origin behavior in go-chi/cors still kicks in, so we must
	// reject the combination just as we would for a lone "*".
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "*"})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSCredentialsRequireExplicitOrigins)
	assert.Nil(t, corsHandler)
}

func TestCORSConfig_RejectsEmptyOriginsWithCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// An explicit `allowedOrigins: []` resolves to the disallowed-origins
	// sentinel, so the runtime behavior is fail-closed -- but pairing it
	// with allowCredentials=true is still an operator misconfiguration
	// (credentials enabled with no usable trusted origin), which we want
	// to surface at startup rather than silently boot a permanently broken
	// credentialed setup.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSCredentialsRequireExplicitOrigins)
	assert.Nil(t, corsHandler)
}

func TestCORSConfig_RejectsAllEmptyStringOriginsWithCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// `allowedOrigins: [""]` + allowCredentials: true used to silently
	// produce a *cors.Cors whose only "explicit" origin was "". Runtime
	// would never match, but the startup safety check in corsConfig also
	// would not fire (len(AllowedOrigins) was 1, not 0), so an operator
	// could boot a credentialed setup with effectively no trusted
	// origins. After the empty-string filter in resolveAllowedOrigins,
	// the list collapses to the disallowedAllOriginsSentinel and the
	// usual credentials-without-origins rejection kicks in.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{""})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSCredentialsRequireExplicitOrigins)
	assert.Nil(t, corsHandler)
}

func TestCORSConfig_RejectsMissingOriginsWithCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// "Credentials-only" config: only allowCredentials is set (e.g. via
	// ALGOREA_CORS__ALLOWCREDENTIALS=true), allowedOrigins is left unset.
	// Resolves to the disallowed-origins sentinel, so nothing would ever
	// match at runtime -- but credentials enabled without any trusted
	// origin is almost always an operator mistake, so we still reject it
	// at startup with a clear error.
	corsConf := viper.New()
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSCredentialsRequireExplicitOrigins)
	assert.Nil(t, corsHandler)
}

func TestCORSConfig_DefaultBlocksAllOrigins(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Without any cors configuration the middleware must deny every
	// cross-origin request -- even non-credentialed ones. This is the
	// regression test for the "default safe" framing in ARCHITECTURE.md:
	// before the disallowedAllOriginsSentinel substitution, an unset
	// allowedOrigins fell through to go-chi/cors's allowedOriginsAll
	// branch and echoed any caller's Origin back, leaving every public
	// endpoint cross-origin readable from any site by default.
	corsConf := viper.New()

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	headers, nextCalled := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://attacker.example")
	assert.True(t, nextCalled, "wrapped handler still runs; CORS just withholds headers")
	assert.Empty(t, headers.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, headers.Get("Access-Control-Allow-Credentials"))

	// Preflight must also be denied: no allowed methods/credentials leak
	// for an origin that is not on the list.
	preflight, _ := dispatchCORSRequest(t, corsHandler, http.MethodOptions, "https://attacker.example")
	assert.Empty(t, preflight.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, preflight.Get("Access-Control-Allow-Methods"))
	assert.Empty(t, preflight.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_PerHostWildcardWithCredentialsIsAllowed(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Per-host wildcards are not the CSRF-via-CORS footgun -- only the bare
	// "*" is rejected. go-chi/cors expands "https://*.example.com" into a
	// pattern match without echoing arbitrary origins back.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://*.example.com"})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	// Matching subdomain: ACAO is echoed and credentials are forwarded.
	matched, _ := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://api.example.com")
	assert.Equal(t, "https://api.example.com", matched.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", matched.Get("Access-Control-Allow-Credentials"))

	// Non-matching origin: no ACAO, no ACAC -- the wildcard does not leak
	// across hosts despite credentials being enabled.
	rejected, _ := dispatchCORSRequest(t, corsHandler, http.MethodGet, "https://attacker.example")
	assert.Empty(t, rejected.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rejected.Get("Access-Control-Allow-Credentials"))
}

func TestCORSConfig_PreflightAllowedOrigin(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Preflight has a different code path in go-chi/cors than the actual
	// request, so cover it explicitly: the configured methods/credentials
	// must reach the OPTIONS response, and an unknown origin must still be
	// rejected at the preflight stage.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org"})
	corsConf.Set(allowCredentialsKey, true)

	corsHandler, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, corsHandler)

	allowed, _ := dispatchCORSRequest(t, corsHandler, http.MethodOptions, "https://www.france-ioi.org")
	assert.Equal(t, "https://www.france-ioi.org", allowed.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", allowed.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, http.MethodGet, allowed.Get("Access-Control-Allow-Methods"))

	rejected, _ := dispatchCORSRequest(t, corsHandler, http.MethodOptions, "https://attacker.example")
	assert.Empty(t, rejected.Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rejected.Get("Access-Control-Allow-Credentials"))
}
