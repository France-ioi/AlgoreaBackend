package app

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestResolveAllowedOrigins_ExplicitListIsForwarded(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "https://algorea.org"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"https://www.france-ioi.org", "https://algorea.org"}, got)
}

func TestResolveAllowedOrigins_MissingKeyDefaultsToEmpty(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()

	got := resolveAllowedOrigins(corsConf)
	assert.Empty(t, got)
}

func TestResolveAllowedOrigins_ExplicitEmptyListStaysEmpty(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{})

	got := resolveAllowedOrigins(corsConf)
	assert.Empty(t, got)
}

func TestResolveAllowedOrigins_WildcardEntryIsPassedThrough(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})

	got := resolveAllowedOrigins(corsConf)
	assert.Equal(t, []string{"*"}, got)
}

func TestCORSConfig_ReturnsCORSHandler(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	c, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.NotNil(t, c.Handler)
}

func TestCORSConfig_AllowCredentialsDefaultsToFalse(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})

	// Default allowCredentials is false, so combining the bare "*" with the
	// default credentials setting must not trip the safety rule.
	c, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestCORSConfig_AllowCredentialsTrueIsForwarded(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org"})
	corsConf.Set(allowCredentialsKey, true)

	c, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestCORSConfig_RejectsWildcardWithCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})
	corsConf.Set(allowCredentialsKey, true)

	c, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSWildcardWithCredentials)
	assert.Nil(t, c)
}

func TestCORSConfig_AllowsWildcardWithoutCredentials(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"*"})
	corsConf.Set(allowCredentialsKey, false)

	c, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestCORSConfig_RejectsWildcardMixedWithExplicit(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Even when the bare "*" is mixed with explicit origins, the unsafe
	// echo-the-Origin behavior in go-chi/cors still kicks in, so we must
	// reject the combination just as we would for a lone "*".
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://www.france-ioi.org", "*"})
	corsConf.Set(allowCredentialsKey, true)

	c, err := corsConfig(corsConf)
	require.ErrorIs(t, err, errCORSWildcardWithCredentials)
	assert.Nil(t, c)
}

func TestCORSConfig_PerHostWildcardWithCredentialsIsAllowed(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// Per-host wildcards are not the CSRF-via-CORS footgun -- only the bare
	// "*" is rejected. go-chi/cors expands "https://*.example.com" into a
	// pattern match without echoing arbitrary origins back.
	corsConf := viper.New()
	corsConf.Set(allowedOriginsKey, []string{"https://*.example.com"})
	corsConf.Set(allowCredentialsKey, true)

	c, err := corsConfig(corsConf)
	require.NoError(t, err)
	require.NotNil(t, c)
}
