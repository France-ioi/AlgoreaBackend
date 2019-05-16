package payloads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnything_MarshalJSON(t *testing.T) {
	anything := Anything(`"value"`)
	result, err := anything.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"value"`), result)
}

func TestAnything_UnmarshalJSON(t *testing.T) {
	raw := []byte(`"value"`)
	anything := Anything("")
	err := anything.UnmarshalJSON(raw)
	assert.NoError(t, err)
	assert.Equal(t, Anything(`"value"`), anything)
}
