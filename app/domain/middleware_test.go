package domain

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		domains            []ConfigItem
		domainOverride     string
		expectedConfig     *CtxConfig
		expectedDomain     string
		expectedStatusCode int
		expectedBody       string
		shouldEnterService bool
	}{
		{
			name: "ok",
			domains: []ConfigItem{
				{
					Domains:       []string{"france-ioi.org", "www.france-ioi.org"},
					AllUsersGroup: 6, TempUsersGroup: 7,
				},
				{
					Domains:       []string{"192.168.0.1", "127.0.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedConfig:     &CtxConfig{AllUsersGroupID: 2, TempUsersGroupID: 4},
			expectedDomain:     "127.0.0.1",
			expectedStatusCode: http.StatusOK,
			shouldEnterService: true,
		},
		{
			name: "use default",
			domains: []ConfigItem{
				{
					Domains:       []string{"france-ioi.org", "www.france-ioi.org"},
					AllUsersGroup: 6, TempUsersGroup: 7,
				},
				{
					Domains:       []string{"default"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedDomain:     "127.0.0.1",
			expectedConfig:     &CtxConfig{AllUsersGroupID: 2, TempUsersGroupID: 4},
			expectedStatusCode: http.StatusOK,
			shouldEnterService: true,
		},
		{
			name: "domain override",
			domains: []ConfigItem{
				{
					Domains:       []string{"france-ioi.org", "www.france-ioi.org"},
					AllUsersGroup: 6, TempUsersGroup: 7,
				},
				{
					Domains:       []string{"default"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			domainOverride:     "www.france-ioi.org",
			expectedDomain:     "www.france-ioi.org",
			expectedConfig:     &CtxConfig{AllUsersGroupID: 6, TempUsersGroupID: 7},
			expectedStatusCode: http.StatusOK,
			shouldEnterService: true,
		},
		{
			name: "wrong domain",
			domains: []ConfigItem{
				{
					Domains:       []string{"france-ioi.org", "www.france-ioi.org"},
					AllUsersGroup: 5, TempUsersGroup: 7,
				},
				{
					Domains:       []string{"192.168.0.1"},
					AllUsersGroup: 2, TempUsersGroup: 4,
				},
			},
			expectedStatusCode: http.StatusNotImplemented,
			expectedBody:       "{\"success\":false,\"message\":\"Not implemented\",\"error_text\":\"Wrong domain \\\"127.0.0.1\\\"\"}",
			shouldEnterService: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assertMiddleware(t, tt.domains, tt.domainOverride, tt.shouldEnterService,
				tt.expectedStatusCode, tt.expectedBody, tt.expectedConfig, tt.expectedDomain)
		})
	}
}

func assertMiddleware(t *testing.T, domains []ConfigItem, domainOverride string, shouldEnterService bool,
	expectedStatusCode int, expectedBody string, expectedConfig *CtxConfig, expectedDomain string,
) {
	// dummy server using the middleware
	middleware := Middleware(domains, domainOverride)
	enteredService := false // used to log if the service has been reached
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enteredService = true // has passed into the service
		configuration := r.Context().Value(ctxDomainConfig).(*CtxConfig)
		assert.Equal(t, expectedConfig, configuration)
		domain := r.Context().Value(ctxDomain).(string)
		assert.Equal(t, expectedDomain, domain)
		w.WriteHeader(http.StatusOK)
	})
	mainSrv := httptest.NewServer(middleware(handler))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest("GET", mainSrv.URL, http.NoBody)
	client := &http.Client{}
	response, err := client.Do(mainRequest)
	var body string
	if err == nil {
		bodyData, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		body = string(bodyData)
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)
	assert.Equal(t, expectedStatusCode, response.StatusCode)
	assert.Equal(t, shouldEnterService, enteredService)
}
