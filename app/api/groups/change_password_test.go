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

	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/servicetest"
)

func TestGenerateGroupPassword(t *testing.T) {
	got, err := GenerateGroupPassword()

	assert.NoError(t, err)
	assert.Len(t, got, 10)
	assert.Regexp(t, `^[3-9a-kmnp-y]+$`, got)
}

func TestGenerateGroupPassword_HandlesError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(rand.Int, func(rand io.Reader, max *big.Int) (n *big.Int, err error) {
		return nil, expectedError
	})
	defer monkey.UnpatchAll()

	_, err := GenerateGroupPassword()
	assert.Equal(t, expectedError, err)
}

func TestService_changePassword_RetriesOnDuplicateEntryError(t *testing.T) {
	response, _, logs, _ := assertMockedChangePasswordRequest(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`SELECT .+ WHERE \(users\.ID = \?\)`).WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `groups_ancestors` WHERE (groups_ancestors.idGroupAncestor=?) AND (idGroupChild = ?)")).
			WithArgs(0, 1).WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))
		mock.ExpectExec("UPDATE `groups` .+").
			WillReturnError(errors.New("ERROR 1062 (23000): Duplicate entry 'aaaaaaaaaa' for key 'sPassword'"))
		mock.ExpectExec("UPDATE `groups` .+").WillReturnResult(sqlmock.NewResult(-1, 1))
	})
	assert.Equal(t, 200, response.StatusCode, logs)
}

func assertMockedChangePasswordRequest(t *testing.T, setMockExpectationsFunc func(sqlmock.Sqlmock)) (*http.Response, sqlmock.Sqlmock, string, error) {
	response, mock, logs, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"POST", "/groups/1/change_password", ``, 2,
		setMockExpectationsFunc,
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: *baseService}
			router.Post("/groups/{group_id}/change_password", service.AppHandler(srv.changePassword).ServeHTTP)
		})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	return response, mock, logs, err
}
