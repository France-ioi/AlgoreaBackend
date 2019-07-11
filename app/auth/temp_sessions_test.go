package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGenerateTempAccessToken(t *testing.T) {
	got1, err := GenerateTempAccessToken()

	assert.NoError(t, err)
	assert.Len(t, got1, 32)
	assert.Regexp(t, `^[0-9a-z]{32}$`, got1)

	got2, err := GenerateTempAccessToken()

	assert.NoError(t, err)
	assert.Len(t, got2, 32)
	assert.Regexp(t, `^[0-9a-z]{32}$`, got2)

	assert.NotEqual(t, got2, got1)
}

func TestGenerateTempAccessToken_HandlesError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(rand.Int, func(rand io.Reader, max *big.Int) (n *big.Int, err error) {
		return nil, expectedError
	})
	defer monkey.UnpatchAll()

	_, err := GenerateTempAccessToken()
	assert.Equal(t, expectedError, err)
}

func TestCreateNewTempSession(t *testing.T) {
	expectedAccessToken := "tmp-01abcdefghijklmnopqrstuvwxyz"
	monkey.Patch(GenerateTempAccessToken, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `sessions` (idUser, sAccessToken, sExpirationDate, sIssuer) VALUES (?, ?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs(expectedUserID, expectedAccessToken, 2*60*60, "backend").
		WillReturnResult(sqlmock.NewResult(1, 1))

	accessToken, expireIn, err := CreateNewTempSession(database.NewDataStore(db).Sessions(), expectedUserID)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccessToken, accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_Retries(t *testing.T) {
	expectedAccessTokens := []string{"tmp-02abcdefghijklmnopqrstuvwxyz", "tmp-03abcdefghijklmnopqrstuvwxyz"}
	accessTokensIndex := -1
	monkey.Patch(GenerateTempAccessToken, func() (string, error) { accessTokensIndex++; return expectedAccessTokens[accessTokensIndex], nil })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `sessions` (idUser, sAccessToken, sExpirationDate, sIssuer) VALUES (?, ?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs(expectedUserID, expectedAccessTokens[0], 2*60*60, "backend").
		WillReturnError(
			&mysql.MySQLError{
				Number:  1062,
				Message: fmt.Sprintf("ERROR 1062 (23000): Duplicate entry '%s' for key 'PRIMARY'", expectedAccessTokens[0]),
			})
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `sessions` (idUser, sAccessToken, sExpirationDate, sIssuer) VALUES (?, ?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs(expectedUserID, expectedAccessTokens[1], 2*60*60, "backend").
		WillReturnResult(sqlmock.NewResult(1, 1))

	accessToken, expireIn, err := CreateNewTempSession(database.NewDataStore(db).Sessions(), expectedUserID)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccessTokens[1], accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_HandlesGeneratorError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(GenerateTempAccessToken, func() (string, error) { return "", expectedError })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)

	accessToken, expireIn, err := CreateNewTempSession(database.NewDataStore(db).Sessions(), expectedUserID)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewTempSession_HandlesDBError(t *testing.T) {
	expectedAccessToken := "tmp-04abcdefghijklmnopqrstuvwxyz"
	monkey.Patch(GenerateTempAccessToken, func() (string, error) { return expectedAccessToken, nil })
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedUserID := int64(12345)
	expectedError := errors.New("some error")
	mock.ExpectExec("^"+regexp.QuoteMeta(
		"INSERT INTO `sessions` (idUser, sAccessToken, sExpirationDate, sIssuer) VALUES (?, ?, NOW() + INTERVAL ? SECOND, ?)",
	)+"$").WithArgs(expectedUserID, expectedAccessToken, 2*60*60, "backend").
		WillReturnError(expectedError)

	accessToken, expireIn, err := CreateNewTempSession(database.NewDataStore(db).Sessions(), expectedUserID)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", accessToken)
	assert.Equal(t, int32(2*60*60), expireIn) // 2 hours

	assert.NoError(t, mock.ExpectationsWereMet())
}
