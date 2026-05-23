package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

func TestJSON_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     interface{}
		wantErr bool
		wantMap map[string]interface{}
	}{
		{
			name: "normal JSON", src: []byte(`{"key":"value","num":123}`),
			wantMap: map[string]interface{}{"key": "value", "num": float64(123)},
		},
		{name: "empty object", src: []byte("{}"), wantMap: map[string]interface{}{}},
		// `src == nil` is the path the driver hits when the DB column holds NULL.
		// It must decode to a nil map without touching the type assertion below.
		{name: "nil src decodes to nil map", src: nil, wantMap: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{}
			if err := j.Scan(tt.src); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantMap, map[string]interface{}(*j))
		})
	}
}

func testJSONSampleMap() map[string]interface{} {
	return map[string]interface{}{
		"a": 123,
		"b": "test",
		"c": map[string]interface{}{
			"nested": true,
		},
	}
}

func TestJSON_Value(t *testing.T) {
	j := golang.Ptr(JSON(testJSONSampleMap()))
	value, err := j.Value()
	require.NoError(t, err)
	valueBytes, ok := value.([]byte)
	require.True(t, ok)
	assert.JSONEq(t, `{"a":123,"b":"test","c":{"nested":true}}`, string(valueBytes))
}

func TestJSON_Value_Nil(t *testing.T) {
	j := (*JSON)(nil)
	value, err := j.Value()
	require.NoError(t, err)
	assert.Nil(t, value)
}
