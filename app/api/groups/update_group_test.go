package groups

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/servicetest"
)

func Test_validateUpdateGroupInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{"code_lifetime=2147483647", `{"code_lifetime":2147483647}`, false},
		{"code_lifetime=0", `{"code_lifetime":0}`, false},
		{"code_lifetime=null", `{"code_lifetime":null}`, false},

		{"code_lifetime=2147483648", `{"code_lifetime":2147483648}`, true},
		{"code_lifetime=", `{"code_lifetime":""}`, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := database.NewDBMock()
			defer func() { _ = db.Close() }()
			database.ClearAllDBEnums()
			database.MockDBEnumQueries(mock)
			defer func() { database.ClearAllDBEnums() }()
			store := database.NewDataStore(db)
			r, _ := http.NewRequest("PUT", "/", strings.NewReader(tt.json))
			_, err := validateUpdateGroupInput(r, &groupUpdateInput{
				CanManageValue: store.GroupManagers().CanManageIndexByName("memberships_and_group"),
			}, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateGroupInput() error = %#v, wantErr %v", err, tt.wantErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_updateGroup_ErrorOnReadInTransaction(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery("^SELECT .* FOR UPDATE").
			WithArgs(2, 1).WillReturnError(errors.New("error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnRefusingSentGroupRequests_Insert(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FOR UPDATE").
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"is_public", "can_manage_value"}).AddRow(true, int64(3)))
		database.ClearAllDBEnums()
		database.MockDBEnumQueries(mock)
		mock.ExpectExec("INSERT INTO group_membership_changes .+").
			WithArgs(2, 1, "join_request").WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
	database.ClearAllDBEnums()
}

func TestService_updateGroup_ErrorOnRefusingSentGroupRequests_Delete(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FOR UPDATE").
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"is_public", "can_manage_value"}).AddRow(true, int64(3)))
		database.ClearAllDBEnums()
		database.MockDBEnumQueries(mock)
		mock.ExpectExec("INSERT INTO group_membership_changes .+").WithArgs(2, 1, "join_request").
			WillReturnResult(sqlmock.NewResult(-1, 1))
		mock.ExpectExec("DELETE FROM `group_pending_requests` .+").WithArgs("join_request", 1).
			WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
	database.ClearAllDBEnums()
}

func TestService_updateGroup_ErrorOnUpdatingGroup(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FOR UPDATE").
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"is_public", "can_manage_value"}).AddRow(false, int64(3)))
		mock.ExpectExec("UPDATE `groups` .+").
			WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func assertUpdateGroupFailsOnDBErrorInTransaction(t *testing.T, setMockExpectationsFunc func(sqlmock.Sqlmock)) {
	response, mock, _, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"PUT", "/groups/1", `{"is_public":false}`, &database.User{GroupID: 2},
		setMockExpectationsFunc,
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: baseService}
			router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
		})

	if err == nil {
		_ = response.Body.Close()
	}
	assert.NoError(t, err)
	assert.Equal(t, 500, response.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_isTryingToChangeOfficialSessionActivity(t *testing.T) {
	type args struct {
		dbMap                map[string]interface{}
		oldIsOfficialSession bool
		activityIDChanged    bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is not an official session, no changes, the field is not present",
			args: args{dbMap: map[string]interface{}{}, oldIsOfficialSession: false, activityIDChanged: false},
			want: false,
		},
		{
			name: "is an official session, no changes, the field is not present",
			args: args{dbMap: map[string]interface{}{}, oldIsOfficialSession: true, activityIDChanged: false},
			want: false,
		},
		{
			name: "is not an official session, no changes, the field is present",
			args: args{dbMap: map[string]interface{}{"is_official_session": false}, oldIsOfficialSession: false, activityIDChanged: false},
			want: false,
		},
		{
			name: "is an official session, no changes, the field is present",
			args: args{dbMap: map[string]interface{}{"is_official_session": true}, oldIsOfficialSession: true, activityIDChanged: false},
			want: false,
		},
		{
			name: "becomes an official session",
			args: args{dbMap: map[string]interface{}{"is_official_session": true}, oldIsOfficialSession: false, activityIDChanged: false},
			want: true,
		},
		{
			name: "becomes an unofficial session",
			args: args{dbMap: map[string]interface{}{"is_official_session": false}, oldIsOfficialSession: true, activityIDChanged: false},
			want: false,
		},
		{
			name: "becomes an unofficial session and the root_activity_id is changed",
			args: args{dbMap: map[string]interface{}{"is_official_session": false}, oldIsOfficialSession: true, activityIDChanged: true},
			want: false,
		},
		{
			name: "is an unofficial session and the root_activity_id is changed",
			args: args{dbMap: map[string]interface{}{"is_official_session": false}, oldIsOfficialSession: false, activityIDChanged: true},
			want: false,
		},
		{
			name: "is an official session and the root_activity_id is changed",
			args: args{dbMap: map[string]interface{}{}, oldIsOfficialSession: true, activityIDChanged: true},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTryingToChangeOfficialSessionActivity(tt.args.dbMap, tt.args.oldIsOfficialSession, tt.args.activityIDChanged))
		})
	}
}

func Test_int64PtrEqualValues(t *testing.T) {
	type args struct {
		a *int64
		b *int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "both are nils", args: args{a: nil, b: nil}, want: true},
		{name: "a is nil, b is not nil", args: args{a: nil, b: ptrInt64(1)}, want: false},
		{name: "a is not nil, b is nil", args: args{a: ptrInt64(0), b: nil}, want: false},
		{name: "both are not nils, but not equal", args: args{a: ptrInt64(0), b: ptrInt64(1)}, want: false},
		{name: "both are not nils, equal", args: args{a: ptrInt64(1), b: ptrInt64(1)}, want: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, int64PtrEqualValues(tt.args.a, tt.args.b))
		})
	}
}

func ptrInt64(i int64) *int64 { return &i }
