package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZero(t *testing.T) {
	assert.False(t, Zero[bool]())
	assert.Equal(t, 0, Zero[int]())
	assert.Equal(t, int64(0), Zero[int64]())
	assert.Equal(t, "", Zero[string]())
	assert.Equal(t, []byte(nil), Zero[[]byte]())
	assert.Equal(t, struct{}{}, Zero[struct{}]())
	assert.Equal(t, (*bool)(nil), Zero[*bool]())
	assert.Equal(t, (*int64)(nil), Zero[*int64]())
	assert.Equal(t, (*string)(nil), Zero[*string]())
}
