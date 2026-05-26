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

func TestJSON_OrEmpty(t *testing.T) {
	populated := JSON{"children_layout": "Grid", "foo": "bar"}
	nilMap := JSON(nil)
	emptyMap := JSON{}
	tests := []struct {
		name string
		in   *JSON
		want JSON
	}{
		{name: "nil pointer receiver returns empty allocated map", in: nil, want: JSON{}},
		{name: "non-nil pointer to nil map returns empty allocated map", in: &nilMap, want: JSON{}},
		{name: "already-empty map is preserved", in: &emptyMap, want: JSON{}},
		{
			name: "populated map is preserved (same keys, same values)",
			in:   &populated,
			want: JSON{"children_layout": "Grid", "foo": "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.OrEmpty()
			assert.Equal(t, tt.want, got)
			// Result must always be a non-nil map so encoding/json emits `{}`
			// rather than `null` for the "never null" response contract.
			assert.NotNil(t, got)
		})
	}
}
