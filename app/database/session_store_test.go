package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore_InsertNewOAuth(t *testing.T) {
	tests := []struct {
		name    string
		issuer  string
		dbError error
	}{
		{name: "success", issuer: "some issuer"},
		{name: "error", dbError: errors.New("some error")},
	}

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	for _, test := range tests {
		test := test

		userID := int64(123456)
		token := "accesstoken"
		secondsUntilExpiry := int32(1234)
		expectedExec := mock.ExpectExec("^"+regexp.QuoteMeta(
			"INSERT INTO `sessions` "+
				"(`access_token`, `expires_at`, `issued_at`, `issuer`, `user_id`) VALUES "+
				"(?, NOW() + INTERVAL ? SECOND, NOW(), ?, ?)")+"$").
			WithArgs(token, secondsUntilExpiry, test.issuer, userID)

		if test.dbError != nil {
			expectedExec.WillReturnError(test.dbError)
		} else {
			expectedExec.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		err := NewDataStore(db).Sessions().InsertNewOAuth(userID, token, secondsUntilExpiry, test.issuer)
		assert.Equal(t, test.dbError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
