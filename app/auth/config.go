package auth

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// GetOAuthConfig generates the OAuth2 config from a configuration
func GetOAuthConfig(config *viper.Viper) *oauth2.Config {

	oauthConfig := oauth2.Config{
		ClientID:     config.GetString("clientID"),
		ClientSecret: config.GetString("clientSecret"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.GetString("loginModuleURL") + "/oauth/authorize",
			TokenURL: config.GetString("loginModuleURL") + "/oauth/token",

			// AuthStyle optionally specifies how the endpoint wants the
			// client id & client secret sent. The zero value means to
			// auto-detect.
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{"account"},
	}
	return &oauthConfig
}
