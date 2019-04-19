package types

import (
	"encoding/json"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

type SampleInt32Input struct {
	Required         RequiredInt32
	Nullable         NullableInt32
	Optional         OptionalInt32
	NullableOptional OptNullInt32
}

func (v *SampleInt32Input) validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "NullableOptional"},
		&v.Required, &v.Nullable, &v.Optional, &v.NullableOptional)
}

func TestNewInt32(t *testing.T) {
	assert := assertlib.New(t)
	var value int32 = 2147483647
	n := NewInt32(value)
	val, null, set := n.AllAttributes()
	assert.Equal(value, n.Value)
	assert.Equal(value, val)
	assert.True(n.Set)
	assert.True(set)
	assert.False(n.Null)
	assert.False(null)
}

func TestInt32Valid(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": 2147483647, "Nullable": 22, "Optional": -1, "NullableOptional": 7 }`
	input := &SampleInt32Input{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.EqualValues(2147483647, input.Required.Value)
	assert.EqualValues(22, input.Nullable.Value)
	assert.EqualValues(-1, input.Optional.Value)
	assert.EqualValues(7, input.NullableOptional.Value)
	assert.NoError(input.validate())
}

func TestInt32WithNonInt(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "not an int", "Nullable": 22, "Optional": -1, "NullableOptional": 7 }`
	input := &SampleInt32Input{}
	assert.Error(json.Unmarshal([]byte(jsonInput), &input))
}

func TestInt32WithDefault(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": 0, "Nullable": 0, "Optional": 0, "NullableOptional": 0 }`
	input := &SampleInt32Input{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.NoError(input.validate())
}

func TestInt32WithNull(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": null, "Nullable": null, "Optional": null, "NullableOptional": null }`
	input := &SampleInt32Input{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.Required.Validate(), "was expecting a validation error")
	assert.NoError(input.Nullable.Validate())         // should be valid
	assert.Error(input.Optional.Validate())           // should NOT be valid
	assert.NoError(input.NullableOptional.Validate()) // should be valid
	assert.Error(input.validate())
}

func TestInt32WithNotSet(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{}`
	input := &SampleInt32Input{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.Required.Validate())           // should NOT be valid
	assert.Error(input.Nullable.Validate())           // should NOT be valid
	assert.NoError(input.Optional.Validate())         // should be valid
	assert.NoError(input.NullableOptional.Validate()) // should be valid
	assert.Error(input.validate())
}
