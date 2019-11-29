package groups

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/servicetest"
)

func TestGenerateGroupCode(t *testing.T) {
	got, err := GenerateGroupCode()

	assert.NoError(t, err)
	assert.Len(t, got, 10)
	assert.Regexp(t, `^[3-9a-kmnp-y]+$`, got)
}

func TestGenerateGroupCode_HandlesError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(rand.Int, func(rand io.Reader, max *big.Int) (n *big.Int, err error) {
		return nil, expectedError
	})
	defer monkey.UnpatchAll()

	_, err := GenerateGroupCode()
	assert.Equal(t, expectedError, err)
}

func TestService_changeCode_RetriesOnDuplicateEntryError(t *testing.T) {
	response, _, logs, _ := assertMockedChangeCodeRequest(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `groups_ancestors` "+
			"JOIN group_managers ON group_managers.group_id = `groups_ancestors`.ancestor_group_id "+
			"JOIN groups_ancestors_active AS user_ancestors "+
			"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND "+
			"user_ancestors.child_group_id = ? "+
			"WHERE (NOW() < `groups_ancestors`.expires_at) AND (groups_ancestors.child_group_id = ?)")).
			WithArgs(2, 1).WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `groups` .+").
			WillReturnError(errors.New("ERROR 1062 (23000): Duplicate entry 'aaaaaaaaaa' for key 'code'"))
		mock.ExpectExec("UPDATE `groups` .+").WillReturnResult(sqlmock.NewResult(-1, 1))
		mock.ExpectCommit()
	})
	assert.Equal(t, 200, response.StatusCode, logs)
}

func assertMockedChangeCodeRequest(t *testing.T,
	setMockExpectationsFunc func(sqlmock.Sqlmock)) (*http.Response, sqlmock.Sqlmock, string, error) {
	response, mock, logs, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"POST", "/groups/1/code", ``, &database.User{GroupID: 2, OwnedGroupID: ptrInt64(10)},
		setMockExpectationsFunc,
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: *baseService}
			router.Post("/groups/{group_id}/code", service.AppHandler(srv.changeCode).ServeHTTP)
		})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	return response, mock, logs, err
}

func ptrInt64(i int64) *int64 { return &i }
