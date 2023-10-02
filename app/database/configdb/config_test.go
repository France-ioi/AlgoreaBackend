package configdb

import (
	"errors"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func Test_mustNotBeError(t *testing.T) {
	mustNotBeError(nil)

	expectedError := errors.New("some error")
	assertlib.PanicsWithValue(t, expectedError, func() {
		mustNotBeError(expectedError)
	})
}
