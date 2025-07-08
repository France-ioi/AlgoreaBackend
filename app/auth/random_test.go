package auth

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	got1, err := GenerateKey()

	assert.NoError(t, err)
	assert.Len(t, got1, 32)
	assert.Regexp(t, `^[0-9a-z]{32}$`, got1)

	got2, err := GenerateKey()

	assert.NoError(t, err)
	assert.Len(t, got2, 32)
	assert.Regexp(t, `^[0-9a-z]{32}$`, got2)

	assert.NotEqual(t, got2, got1)
}

func TestGenerateKey_HandlesError(t *testing.T) {
	expectedError := errors.New("some error")
	monkey.Patch(rand.Int, func(_ io.Reader, _ *big.Int) (n *big.Int, err error) {
		return nil, expectedError
	})
	defer monkey.UnpatchAll()

	_, err := GenerateKey()
	assert.Equal(t, expectedError, err)
}
