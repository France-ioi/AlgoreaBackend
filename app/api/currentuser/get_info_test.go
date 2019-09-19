package currentuser

import (
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func TestService_getInfo_Returns403WhenUserNotFound(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("^SELECT").WillReturnRows(mock.NewRows([]string{"id"})) // no rows

	srv := &Service{Base: service.Base{Store: database.NewDataStore(db)}}
	monkey.PatchInstanceMethod(reflect.TypeOf(&srv.Base), "GetUser", func(*service.Base, *http.Request) *database.User {
		return &database.User{ID: 123}
	})
	result := srv.getInfo(nil, nil)
	assert.Equal(t, service.InsufficientAccessRightsError, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}
