package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus" //nolint:depguard
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func TestNew_Success(t *testing.T) {
	assert := assertlib.New(t)
	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)
	assert.NotNil(app.Config)
	assert.NotNil(app.Database)
	assert.NotNil(app.HTTPHandler)
	assert.Len(app.HTTPHandler.Middlewares(), 7)
	assert.True(len(app.HTTPHandler.Routes()) > 0)
}

func TestNew_DBErr(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	expectedError := errors.New("db opening error")
	patch := monkey.Patch(database.Open, func(interface{}) (*database.DB, error) {
		return nil, expectedError
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.Equal(expectedError, err)
	logMsg := hook.LastEntry()
	assert.Equal(logrus.ErrorLevel, logMsg.Level)
	assert.Equal("db opening error", logMsg.Message)
	assert.Equal("database", logMsg.Data["module"])
}

func TestNew_APIErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(api.NewCtx,
		func(conf *config.Root, db *database.DB, tokenConfig *token.Config) (*api.Ctx, error) {
			return nil, errors.New("api creation error")
		})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "api creation error")
}

func TestNew_TokenErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(token.Initialize, func(*config.Token) (*token.Config, error) {
		return nil, errors.New("keys loading error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "keys loading error")
}

// The goal of the following `TestMiddlewares*` tests are not to test the middleware themselves
// but their interaction (impacted by the order of definition)

func TestMiddlewares_OnPanic(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	app, _ := New()
	router := app.HTTPHandler
	router.Get("/dummy", func(http.ResponseWriter, *http.Request) {
		panic("error in service")
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	nbLogsBeforeRequest := len(hook.AllEntries())
	request, _ := http.NewRequest("GET", srv.URL+"/dummy", nil)
	request.Header.Set("X-Forwarded-For", "1.1.1.1")
	response, _ := http.DefaultClient.Do(request)
	respBody, _ := ioutil.ReadAll(response.Body)

	// check that the error has been handled by the recover
	assert.Equal(http.StatusInternalServerError, response.StatusCode)
	assert.Equal("Internal Server Error\n", string(respBody))
	assert.Equal("text/plain; charset=utf-8", response.Header.Get("Content-type"))
	allLogs := hook.AllEntries()
	assert.Equal(2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the recovere put the error info in the logs
	assert.Equal("error in service", hook.LastEntry().Data["panic"])
	assert.NotNil(hook.LastEntry().Data["stack"])
	// check that the real IP is used in the logs
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
}

func TestMiddlewares_OnSuccess(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	app, _ := New()
	router := app.HTTPHandler
	router.Get("/dummy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"data\":\"datadatadata\"}"))
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	nbLogsBeforeRequest := len(hook.AllEntries())
	request, _ := http.NewRequest("GET", srv.URL+"/dummy", nil)
	request.Header.Set("X-Real-IP", "1.1.1.1")
	request.Header.Set("Accept-Encoding", "gzip, deflate")
	response, _ := http.DefaultClient.Do(request)
	assert.NotNil(response.Header.Get("Content-type"))
	assert.Equal("application/json", response.Header.Get("Content-Type"))
	allLogs := hook.AllEntries()
	assert.Equal(2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the real IP is used in the logs
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
	// check that the compression has been applied but the length in the logs is not altered by compression i
	assert.Equal(23, hook.LastEntry().Data["resp_bytes_length"])
	assert.Equal("gzip", response.Header.Get("Content-Encoding"))
}

func TestNew_MountsPprofInDev(t *testing.T) {
	assert := assertlib.New(t)

	monkey.Patch(appenv.IsEnvDev, func() bool { return true })
	defer monkey.UnpatchAll()

	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest("GET", srv.URL+"/debug", nil)
	response, _ := http.DefaultClient.Do(request)
	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Contains(string(body), "Types of profiles available:")
}

func TestNew_DoesNotMountPprofInEnvironmentsOtherThanDev(t *testing.T) {
	assert := assertlib.New(t)

	monkey.Patch(appenv.IsEnvDev, func() bool { return false })
	defer monkey.UnpatchAll()

	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest("GET", srv.URL+"/debug", nil)
	response, _ := http.DefaultClient.Do(request)
	assert.Equal(502, response.StatusCode)
	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Equal("", string(body))
}

type relationSpec struct {
	database.ParentChild
	exists bool
	error  bool
}

func TestApplication_SelfCheck(t *testing.T) {
	type groupSpec struct {
		name   string
		id     int64
		exists bool
		error  bool
	}

	tests := []struct {
		name                     string
		config                   *config.Root
		expectedGroupsToCheck    []groupSpec
		expectedRelationsToCheck []relationSpec
		expectedError            error
	}{
		{
			name: "everything is okay",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1", "192.168.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
				{
					Domains:   []string{"www.france-ioi.org"},
					RootGroup: 5, RootSelfGroup: 6, RootAdminGroup: 7, RootTempGroup: 8,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4, exists: true},
				{name: "Root", id: 5, exists: true}, {name: "RootSelf", id: 6, exists: true},
				{name: "RootAdmin", id: 7, exists: true}, {name: "RootTemp", id: 8, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 2}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 3}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 2, ChildID: 4}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 5, ChildID: 6}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 5, ChildID: 7}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 6, ChildID: 8}, exists: true},
			},
		},
		{
			name: "Root is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"192.168.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1},
			},
			expectedError: errors.New("no Root group for domain \"192.168.0.1\""),
		},
		{
			name: "RootSelf is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"192.168.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2},
			},
			expectedError: errors.New("no RootSelf group for domain \"192.168.0.1\""),
		},
		{
			name: "RootAdmin is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3},
			},
			expectedError: errors.New("no RootAdmin group for domain \"127.0.0.1\""),
		},
		{
			name: "RootTemp is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4},
			},
			expectedError: errors.New("no RootTemp group for domain \"127.0.0.1\""),
		},
		{
			name: "Root -> RootSelf relation is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 2}},
			},
			expectedError: errors.New("no Root -> RootSelf link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name: "Root -> RootAdmin relation is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 2}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 3}},
			},
			expectedError: errors.New("no Root -> RootAdmin link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name: "RootSelf -> RootTemp relation is missing",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 2}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 3}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 2, ChildID: 4}},
			},
			expectedError: errors.New("no RootSelf -> RootTemp link in groups_groups for domain \"127.0.0.1\""),
		},
		{
			name: "error on group checking",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, error: true},
			},
			expectedError: errors.New("some error"),
		},
		{
			name: "error on relation checking",
			config: &config.Root{Domains: []config.Domain{
				{
					Domains:   []string{"127.0.0.1"},
					RootGroup: 1, RootSelfGroup: 2, RootAdminGroup: 3, RootTempGroup: 4,
				},
			}},
			expectedGroupsToCheck: []groupSpec{
				{name: "Root", id: 1, exists: true}, {name: "RootSelf", id: 2, exists: true},
				{name: "RootAdmin", id: 3, exists: true}, {name: "RootTemp", id: 4, exists: true},
			},
			expectedRelationsToCheck: []relationSpec{
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 2}, exists: true},
				{ParentChild: database.ParentChild{ParentID: 1, ChildID: 3}, error: true},
			},
			expectedError: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := database.NewDBMock()
			defer func() { _ = db.Close() }()
			mock.MatchExpectationsInOrder(false)

			var expectedError error

			for _, expectedGroupToCheck := range tt.expectedGroupsToCheck {
				queryMock := mock.ExpectQuery("^"+regexp.QuoteMeta(
					"SELECT 1 FROM `groups`  WHERE (groups.ID = ?) AND (sType = 'Base') AND (sName = ?) AND (sTextId = ?) LIMIT 1",
				)+"$").
					WithArgs(expectedGroupToCheck.id, expectedGroupToCheck.name, expectedGroupToCheck.name)
				if !expectedGroupToCheck.error {
					rowsToReturn := mock.NewRows([]string{"1"})
					if expectedGroupToCheck.exists {
						rowsToReturn.AddRow(1)
					}
					queryMock.WillReturnRows(rowsToReturn)
				} else {
					expectedError = errors.New("some error")
					queryMock.WillReturnError(expectedError)
				}
			}
			if expectedError == nil {
				for _, expectedRelationToCheck := range tt.expectedRelationsToCheck {
					rowsToReturn := mock.NewRows([]string{"1"})
					if expectedRelationToCheck.exists {
						rowsToReturn.AddRow(1)
					}
					queryMock := mock.ExpectQuery("^"+regexp.QuoteMeta(
						"SELECT 1 FROM `groups_groups`  WHERE (sType = 'direct') AND (idGroupParent = ?) AND (idGroupChild = ?) LIMIT 1",
					)+"$").
						WithArgs(expectedRelationToCheck.ParentID, expectedRelationToCheck.ChildID)
					if !expectedRelationToCheck.error {
						queryMock.WillReturnRows(rowsToReturn)
					} else {
						expectedError = errors.New("some error")
						queryMock.WillReturnError(expectedError)
					}
				}
			}

			app := &Application{
				Config:   tt.config,
				Database: db,
			}
			err := app.SelfCheck()
			assertlib.Equal(t, tt.expectedError, err)
			assertlib.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
