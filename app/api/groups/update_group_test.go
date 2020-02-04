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
		{"code_lifetime=99:59:59", `{"code_lifetime":"99:59:59"}`, false},
		{"code_lifetime=00:00:00", `{"code_lifetime":"00:00:00"}`, false},

		{"code_lifetime=99:60:59", `{"code_lifetime":"99:60:59"}`, true},
		{"code_lifetime=99:59:60", `{"code_lifetime":"99:59:60"}`, true},
		{"code_lifetime=59:59", `{"code_lifetime":"59:59"}`, true},
		{"code_lifetime=59", `{"code_lifetime":"59"}`, true},
		{"code_lifetime=59", `{"code_lifetime":"invalid"}`, true},
		{"code_lifetime=", `{"code_lifetime":""}`, true},

		{"redirect_path=9", `{"redirect_path":"9"}`, false},
		{"redirect_path=1234567890", `{"redirect_path":"1234567890"}`, false},
		{"redirect_path=1234567890/0", `{"redirect_path":"1234567890/0"}`, false},
		{"redirect_path=0/1234567890", `{"redirect_path":"0/1234567890"}`, false},
		{"redirect_path=1234567890/1234567890", `{"redirect_path":"1234567890/1234567890"}`, false},
		// empty strings are allowed (there are some in the DB)
		{"redirect_path=", `{"redirect_path":""}`, false},

		{"redirect_path=invalid", `{"redirect_path":"invalid"}`, true},
		{"redirect_path=1A", `{"redirect_path":"1A"}`, true},
		{"redirect_path=1A/2B", `{"redirect_path":"1A/2B"}`, true},
		{"redirect_path=1234567890/1234567890/1", `{"redirect_path":"1234567890/1234567890/1"}`, false},
	}
	for _, tt := range tests {
		tt := tt
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
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.is_public FROM `groups` "+
			"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id "+
			"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id "+
			"JOIN groups_ancestors_active AS user_ancestors "+
			"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND "+
			"user_ancestors.child_group_id = ? "+
			"WHERE (groups.id = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(2, 1).WillReturnError(errors.New("error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnRefusingSentGroupRequests_Insert(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.is_public FROM `groups` "+
			"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id "+
			"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id "+
			"JOIN groups_ancestors_active AS user_ancestors "+
			"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND "+
			"user_ancestors.child_group_id = ? "+
			"WHERE (groups.id = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))
		mock.ExpectExec("INSERT INTO group_membership_changes .+").
			WithArgs(2, 1).WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnRefusingSentGroupRequests_Delete(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.is_public FROM `groups` "+
			"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id "+
			"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id "+
			"JOIN groups_ancestors_active AS user_ancestors "+
			"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND "+
			"user_ancestors.child_group_id = ? "+
			"WHERE (groups.id = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))
		mock.ExpectExec("INSERT INTO group_membership_changes .+").WithArgs(2, 1).
			WillReturnResult(sqlmock.NewResult(-1, 1))
		mock.ExpectExec("DELETE FROM `group_pending_requests` .+").WithArgs(1).
			WillReturnError(errors.New("some error"))
		mock.ExpectRollback()
	})
}

func TestService_updateGroup_ErrorOnUpdatingGroup(t *testing.T) {
	assertUpdateGroupFailsOnDBErrorInTransaction(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT groups.is_public FROM `groups` "+
			"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id "+
			"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id "+
			"JOIN groups_ancestors_active AS user_ancestors "+
			"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND "+
			"user_ancestors.child_group_id = ? "+
			"WHERE (groups.id = ?) LIMIT 1 FOR UPDATE")).
			WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(false))
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
			srv := &Service{Base: *baseService}
			router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
		})

	assert.NoError(t, err)
	assert.Equal(t, 500, response.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}
