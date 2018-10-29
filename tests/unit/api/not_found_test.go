package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/spf13/viper"
)

func TestNotFoundProxy(t *testing.T) {

	expected_path := "/a_path_on_backend"

	/* setup backend */
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Got-Query", r.URL.Path)
		w.Write([]byte("hi"))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	/* setup tested web server */
	config.Path = "../../../conf/default.yaml"
	viper.Set("ReverseProxy.Server", backendURL.String())
	application, err := app.New()
	if err != nil {
		fmt.Println("Unable to load app")
		panic(err)
	}

	/* calling web server */
	frontend := httptest.NewServer(application.HTTPHandler)
	req, _ := http.NewRequest("GET", frontend.URL+expected_path, nil)
	req.Close = true
	res, err := frontend.Client().Do(req)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g := res.Header.Get("X-Got-Query"); g != expected_path {
		t.Errorf("got query %q; expected %q", g, expected_path)
	}
	res.Body.Close()
	frontend.Close()
}
