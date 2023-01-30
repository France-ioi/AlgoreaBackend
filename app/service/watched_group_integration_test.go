// +build !unit

package service_test

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestBase_ResolveWatchedGroupID(t *testing.T) {
	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1, type: Class}, {id: 2, type: Team}, {id: 3, type: Other}, {id: 4, type: User}, {id: 5, type: User}, {id: 6, type: Team}]
		groups_groups:
			- {parent_group_id: 1, child_group_id: 4}
			- {parent_group_id: 3, child_group_id: 6}
		group_managers:
			- {manager_id: 1, group_id: 2}
			- {manager_id: 1, group_id: 3, can_watch_members: true}
			- {manager_id: 4, group_id: 5, can_watch_members: true}
	`)
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)
	assert.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
		return trStore.GroupGroups().After()
	}))

	srv := &service.Base{}
	srv.SetGlobalStore(store)
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(srv), "GetUser",
		func(*service.Base, *http.Request) *database.User { return &database.User{GroupID: 4} })
	defer patch.Unpatch()

	forbiddenError := service.ErrForbidden(errors.New("no rights to watch for watched_group_id"))

	tests := []struct {
		name               string
		url                string
		wantWatchedGroupID int64
		wantOk             bool
		wantAPIError       service.APIError
	}{
		{name: "watched_group_id is not managed by the user", url: "?watched_group_id=4", wantAPIError: forbiddenError},
		{name: "no can_watch_members permission", url: "?watched_group_id=2", wantAPIError: forbiddenError},
		{name: "managed by an ancestor", url: "?watched_group_id=3", wantWatchedGroupID: 3, wantOk: true},
		{name: "managed by the user", url: "?watched_group_id=5", wantWatchedGroupID: 5, wantOk: true},
		{name: "an ancestor is managed", url: "?watched_group_id=6", wantWatchedGroupID: 6, wantOk: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			watchedGroupID, ok, apiError := srv.ResolveWatchedGroupID(req)
			assert.Equal(t, tt.wantWatchedGroupID, watchedGroupID)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}
