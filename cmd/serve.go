package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/common"
)

func init() { // nolint:gochecknoinits

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "start http server",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			log.Println("Starting application: environment =", common.Env())

			var application *app.Application
			application, err = app.New()
			if err != nil {
				log.Fatal(err)
			}

			if common.IsEnvDev() {
				backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					dataJSON := fmt.Sprintf(`{"userID": %d, "error":""}`, 1) // user_id = 1
					_, _ = w.Write([]byte(dataJSON))                         // nolint
				}))
				defer backend.Close()

				// put the backend URL into the config
				backendURL, _ := url.Parse(backend.URL) // nolint
				application.Config.Auth.ProxyURL = backendURL.String()
			}

			var server *app.Server
			server, err = app.NewServer(application)
			if err != nil {
				log.Fatal(err)
			}
			<-server.Start()
		},
	}

	rootCmd.AddCommand(serveCmd)
}
