package auth

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// GetOAuthConfig generates the OAuth2 config from a configuration
func GetOAuthConfig(config *viper.Viper) *oauth2.Config {

	oauthConfig := oauth2.Config{
		ClientID:     config.GetString("ClientID"),
		ClientSecret: config.GetString("ClientSecret"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.GetString("LoginModuleURL") + "/oauth/authorize",
			TokenURL: config.GetString("LoginModuleURL") + "/oauth/token",

			// AuthStyle optionally specifies how the endpoint wants the
			// client id & client secret sent. The zero value means to
			// auto-detect.
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: config.GetString("CallbackURL"),
		Scopes:      []string{"account"},
	}
	return &oauthConfig
}