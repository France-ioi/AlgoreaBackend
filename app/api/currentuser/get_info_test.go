package currentuser

import (
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

func TestService_getInfo_Returns403WhenUserNotFound(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("^SELECT").WillReturnRows(mock.NewRows([]string{"id"})) // no rows

	srv := &Service{Base: &service.Base{}}
	srv.Base.SetGlobalStore(database.NewDataStore(db))
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(srv.Base), "GetUser", func(*service.Base, *http.Request) *database.User {
		return &database.User{GroupID: 123}
	})
	defer patch.Unpatch()
	request, _ := http.NewRequest("GET", "", http.NoBody)
	result := srv.getInfo(nil, request)
	assert.Equal(t, service.ErrAPIInsufficientAccessRights, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}
