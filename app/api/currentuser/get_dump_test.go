package currentuser

import (
	"errors"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/servicetest"
)

func TestService_getDump_ReturnsErrorRightInsideTheResponseBody(t *testing.T) {
	response, mock, _, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
		"GET", "/current-user/dump", ``,
		&database.User{ID: 1, OwnedGroupID: ptrInt64(10), SelfGroupID: ptrInt64(11)},
		func(sqlmock sqlmock.Sqlmock) {
			sqlmock.ExpectQuery("^" + regexp.QuoteMeta(
				"SELECT CONCAT('`', TABLE_NAME, '`.`', COLUMN_NAME, '`') FROM `INFORMATION_SCHEMA`.`COLUMNS`  "+
					"WHERE (TABLE_SCHEMA = ?) AND (TABLE_NAME = ?) AND (COLUMN_NAME NOT IN (?))",
			) + "$").WillReturnRows(sqlmock.NewRows([]string{"names"}).AddRow("users.ID").AddRow("users.sName"))
			sqlmock.ExpectQuery("^" + regexp.QuoteMeta(
				"SELECT users.ID, users.sName FROM `users`  WHERE (users.ID = ?)") + "$").
				WillReturnError(errors.New("some error"))
		},
		func(router *chi.Mux, baseService *service.Base) {
			srv := &Service{Base: *baseService}
			srv.Config = &config.Root{Database: config.Database{Connection: mysql.Config{DBName: "test_db"}}}
			router.Get("/current-user/dump", service.AppHandler(srv.getDump).ServeHTTP)
		})
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "attachment; filename=user_data.json", response.Header.Get("Content-Disposition"))
	assert.Equal(t, "application/json; charset=utf-8", response.Header.Get("Content-Type"))
	body, _ := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	assert.Equal(t, `{"current_user":{"success":false,"message":"Internal Server Error","error_text":"Some error"}`+"\n",
		string(body))
	assert.NoError(t, mock.ExpectationsWereMet())

}

func ptrInt64(i int64) *int64 { return &i }
