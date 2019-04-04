package groups

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func TestService_checkThatUserOwnsTheGroup_HandlesError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(1).
		WillReturnError(expectedError)

	user := database.NewUser(1, database.NewDataStore(db).Users(), nil)
	srv := Service{service.Base{Store: database.NewDataStore(db)}}

	apiErr := srv.checkThatUserOwnsTheGroup(user, 123)

	assert.Equal(t, service.ErrUnexpected(expectedError), apiErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}
