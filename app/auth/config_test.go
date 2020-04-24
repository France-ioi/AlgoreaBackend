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
	config.Set("CallbackURL", "https://AlBackend.com/login")
	c := GetOAuthConfig(config)
	assert.Equal(t, "c1", c.ClientID)
	assert.Equal(t, "c2", c.ClientSecret)
	assert.Equal(t, "http://lm.org/oauth/authorize", c.Endpoint.AuthURL)
	assert.Equal(t, "http://lm.org/oauth/token", c.Endpoint.TokenURL)
	assert.Equal(t, oauth2.AuthStyleInParams, c.Endpoint.AuthStyle)
	assert.Equal(t, "https://AlBackend.com/login", c.RedirectURL)
	assert.Equal(t, []string{"account"}, c.Scopes)
}
