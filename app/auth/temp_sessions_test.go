package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)
	expectedError := errors.New("some error")
	assert.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}
