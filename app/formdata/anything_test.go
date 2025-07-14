package formdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnything_MarshalJSON(t *testing.T) {
	anything := AnythingFromString(`"value"`)
	result, err := anything.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`"value"`), result)
}

func TestAnything_MarshalJSON_EmptyValue(t *testing.T) {
	anything := Anything{raw: nil}
	result, err := anything.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`null`), result)
}

func TestAnything_UnmarshalJSON(t *testing.T) {
	raw := []byte(`"value"`)
	anything := AnythingFromString("")
	err := anything.UnmarshalJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, AnythingFromString(`"value"`), anything)
}

func TestAnything_Bytes(t *testing.T) {
	anything := AnythingFromString(`"value"`)
	assert.Equal(t, []byte(`"value"`), anything.Bytes())
}
