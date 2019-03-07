package groups

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func Test_validateUpdateGroupInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{"sType=Class", `{"type":"Class"}`, false},
		{"sType=Team", `{"type":"Team"}`, false},
		{"sType=Club", `{"type":"Club"}`, false},
		{"sType=Friends", `{"type":"Friends"}`, false},
		{"sType=Other", `{"type":"Other"}`, false},

		{"sType=Root", `{"type":"Root"}`, true},
		{"sType=UserSelf", `{"type":"UserSelf"}`, true},
		{"sType=UserAdmin", `{"type":"UserAdmin"}`, true},
		{"sType=RootSelf", `{"type":"RootSelf"}`, true},
		{"sType=RootAdmin", `{"type":"RootAdmin"}`, true},
		{"sType=unknown", `{"type":"unknown"}`, true},
		{"sType=", `{"type":""}`, true},

		{"sPasswordTimer=99:59:59", `{"password_timer":"99:59:59"}`, false},
		{"sPasswordTimer=00:00:00", `{"password_timer":"00:00:00"}`, false},

		{"sPasswordTimer=99:60:59", `{"password_timer":"99:60:59"}`, true},
		{"sPasswordTimer=99:59:60", `{"password_timer":"99:59:60"}`, true},
		{"sPasswordTimer=59:59", `{"password_timer":"59:59"}`, true},
		{"sPasswordTimer=59", `{"password_timer":"59"}`, true},
		{"sPasswordTimer=59", `{"password_timer":"invalid"}`, true},
		{"sPasswordTimer=", `{"password_timer":""}`, true},

		{"sRedirectPath=9", `{"redirect_path":"9"}`, false},
		{"sRedirectPath=1234567890", `{"redirect_path":"1234567890"}`, false},
		{"sRedirectPath=1234567890/0", `{"redirect_path":"1234567890/0"}`, false},
		{"sRedirectPath=0/1234567890", `{"redirect_path":"0/1234567890"}`, false},
		{"sRedirectPath=1234567890/1234567890", `{"redirect_path":"1234567890/1234567890"}`, false},
		// empty strings are allowed (there are some in the DB)
		{"sRedirectPath=", `{"redirect_path":""}`, false},

		{"sRedirectPath=invalid", `{"redirect_path":"invalid"}`, true},
		{"sRedirectPath=1A", `{"redirect_path":"1A"}`, true},
		{"sRedirectPath=1A/2B", `{"redirect_path":"1A/2B"}`, true},
		{"sRedirectPath=1234567890/1234567890/1", `{"redirect_path":"1234567890/1234567890/1"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("PUT", "/", strings.NewReader(tt.json))
			_, err := validateUpdateGroupInput(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateGroupInput() error = %#v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_updateGroup_ErrorOnReadInTransaction(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT .+ WHERE \(users\.ID = \?\)`).WithArgs(2).WillReturnError(errors.New("error"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.bFreeAccess FROM `groups` JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID WHERE (groups_ancestors.idGroupAncestor=?) AND (groups.ID = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(0, 1).WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnRefusingSentGroupRequests(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT .+ WHERE \(users\.ID = \?\)`).WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.bFreeAccess FROM `groups` JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID WHERE (groups_ancestors.idGroupAncestor=?) AND (groups.ID = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(0, 1).WillReturnRows(sqlmock.NewRows([]string{"bFreeAccess"}).AddRow(true))
		mock.ExpectExec("UPDATE `groups_groups` .+").WithArgs("requestRefused", 1).
			WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnUpdatingGroup(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT .+ WHERE \(users\.ID = \?\)`).WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.bFreeAccess FROM `groups` JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID WHERE (groups_ancestors.idGroupAncestor=?) AND (groups.ID = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(0, 1).WillReturnRows(sqlmock.NewRows([]string{"bFreeAccess"}).AddRow(false))
		mock.ExpectExec("UPDATE `groups` .+").
			WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func assertUpdateGroupFailsOnDBErrorInTransaction(t *testing.T, setMockExpectationsFunc func(sqlmock.Sqlmock)) {
	response, mock, _, err := service.GetResponseForTheRouteWithMockedDBAndUser(
		"PUT", "/groups/1", `{"free_access":false}`, 2,
		setMockExpectationsFunc,
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: *baseService}
			router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
		})

	assert.NoError(t, err)
	assert.Equal(t, 500, response.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}
