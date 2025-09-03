package formdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

func TestAnything_MarshalJSON(t *testing.T) {
	anything := AnythingFromString(`"value"`)
	result, err := anything.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`"value"`), result)
}

func TestAnything_MarshalJSON_NilReceiver(t *testing.T) {
	anything := (*Anything)(nil)
	result, err := anything.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`null`), result)
}

func TestAnything_MarshalJSON_NilPointer(t *testing.T) {
	anything := Anything{raw: nil}
	result, err := anything.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`null`), result)
}

func TestAnything_MarshalJSON_EmptyValue(t *testing.T) {
	anything := Anything{raw: golang.Ptr([]byte{})}
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

func TestAnything_Bytes_NilReceiver(t *testing.T) {
	anything := (*Anything)(nil)
	assert.Equal(t, []byte(nil), anything.Bytes())
}

func TestAnything_Bytes_NilPointer(t *testing.T) {
	anything := &Anything{raw: nil}
	assert.Equal(t, []byte(nil), anything.Bytes())
}
