package token

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_recoverPanics_and_mustNotBeError(t *testing.T) {
	expectedError := errors.New("some error")
	err := func() (err error) {
		defer recoverPanics(&err)
		mustNotBeError(expectedError)
		return nil
	}()
	assert.Equal(t, &UnexpectedError{expectedError}, err)
	assert.Equal(t, expectedError.Error(), err.Error())
}

func Test_UnexpectedError(t *testing.T) {
	assert.True(t, IsUnexpectedError(&UnexpectedError{err: errors.New("some error")}))
	assert.False(t, IsUnexpectedError(errors.New("some error")))
	assert.False(t, IsUnexpectedError(nil))
}
