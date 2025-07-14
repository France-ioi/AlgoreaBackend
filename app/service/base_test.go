package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func TestBase_GetUser(t *testing.T) {
	middleware := auth.MockUserMiddleware(&database.User{GroupID: 42})
	called := false
	testServer := httptest.NewServer(middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		srv := &Base{}
		user := srv.GetUser(r)
		assert.Equal(t, int64(42), user.GroupID)
	})))
	defer testServer.Close()

	request, _ := http.NewRequest(http.MethodGet, testServer.URL, http.NoBody)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	if err == nil {
		_ = response.Body.Close()
	}

	assert.True(t, called)
}

func TestBase_GetStore(t *testing.T) {
	db, _ := database.NewDBMock()
	defer func() { _ = db.Close() }()
	expectedDB := db
	expectedContext := context.Background()
	expectedStore := database.NewDataStoreWithContext(expectedContext, expectedDB)
	req := (&http.Request{}).WithContext(expectedContext)
	store := (&Base{store: database.NewDataStore(expectedDB)}).GetStore(req)
	assert.Equal(t, *expectedStore.DB, *store.DB)
}

func TestBase_GetStore_WithNilStore(t *testing.T) {
	req := &http.Request{}
	assert.Nil(t, (&Base{}).GetStore(req))
}

func TestBase_GetPropagationEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		ServerConfig func() *viper.Viper
		want         string
	}{
		{
			name:         "should be empty if no config",
			ServerConfig: viper.New,
			want:         "",
		},
		{
			name: "should be empty if the propagation endpoint is not set",
			ServerConfig: func() *viper.Viper {
				config := viper.New()
				config.Set("propagation_endpoint", "")
				return config
			},
			want: "",
		},
		{
			name: "should return the endpoint if it is set",
			ServerConfig: func() *viper.Viper {
				config := viper.New()
				config.Set("propagation_endpoint", "https://example.com")
				return config
			},
			want: "https://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Base{
				ServerConfig: tt.ServerConfig(),
			}
			assert.Equalf(t, tt.want, srv.GetPropagationEndpoint(), "GetPropagationEndpoint()")
		})
	}
}
