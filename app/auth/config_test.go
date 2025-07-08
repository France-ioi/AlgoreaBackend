package auth

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestGetOAuthConfig(t *testing.T) {
	config := viper.New()
	config.Set("clientid", "c1")
	config.Set("clientSECRET", "c2")
	config.Set("LOGINMODULEURL", "http://lm.org")
	oauthConfig := GetOAuthConfig(config)
	assert.Equal(t, "c1", oauthConfig.ClientID)
	assert.Equal(t, "c2", oauthConfig.ClientSecret)
	assert.Equal(t, "http://lm.org/oauth/authorize", oauthConfig.Endpoint.AuthURL)
	assert.Equal(t, "http://lm.org/oauth/token", oauthConfig.Endpoint.TokenURL)
	assert.Equal(t, oauth2.AuthStyleInParams, oauthConfig.Endpoint.AuthStyle)
	assert.Equal(t, []string{"account"}, oauthConfig.Scopes)
}
