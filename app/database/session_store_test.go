package database

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestSessionStore_InsertNewOAuth(t *testing.T) {
	tests := []struct {
		name    string
		dbError error
	}{{name: "success"}, {name: "error", dbError: errors.New("some error")}}
	for _, test := range tests {
		test := test

		db, mock := NewDBMock()
		defer func() { _ = db.Close() }()

		userID := int64(123456)
		token := oauth2.Token{
			AccessToken: "accesstoken",
			Expiry:      time.Now(),
		}
		expectedExec := mock.ExpectExec("^" + regexp.QuoteMeta(
			"INSERT INTO `sessions` (access_token, expiration_date, issued_at_date, issuer, user_id) VALUES "+
				"(?, ?, NOW(), ?, ?)") + "$")

		if test.dbError != nil {
			expectedExec.WillReturnError(test.dbError)
		} else {
			expectedExec.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		err := NewDataStore(db).Sessions().InsertNewOAuth(userID, &token)
		assert.Equal(t, test.dbError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
