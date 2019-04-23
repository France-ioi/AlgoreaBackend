package types

import (
	"encoding/json"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

type SampleIntInput struct {
	Required         RequiredInt64
	Nullable         NullableInt64
	Optional         OptionalInt64
	OptionalNullable OptNullInt64
}

func (v *SampleIntInput) validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}

func TestNewInt(t *testing.T) {
	assert := assertlib.New(t)
	var value int64 = 2147483645
	n := NewInt64(value)
	val, null, set := n.AllAttributes()
	assert.Equal(value, n.Value)
	assert.Equal(value, val)
	assert.True(n.Set)
	assert.True(set)
	assert.False(n.Null)
	assert.False(null)
}

func TestIntValid(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "2147483645", "Nullable": "22", "Optional": "-1", "OptionalNullable": "7" }`
	input := &SampleIntInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.EqualValues(2147483645, input.Required.Value)
	assert.EqualValues(22, input.Nullable.Value)
	assert.EqualValues(-1, input.Optional.Value)
	assert.EqualValues(7, input.OptionalNullable.Value)
	assert.NoError(input.validate())
}

func TestIntWithNonInt(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "not an int", "Nullable": "22", "Optional": "-1", "OptionalNullable": "7" }`
	input := &SampleIntInput{}
	assert.Error(json.Unmarshal([]byte(jsonInput), &input))
}

func TestIntWithDefault(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "0", "Nullable": "0", "Optional": "0", "OptionalNullable": "0" }`
	input := &SampleIntInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.NoError(input.validate())
}

func TestIntWithNull(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": null, "Nullable": null, "Optional": null, "OptionalNullable": null }`
	input := &SampleIntInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.Required.Validate(), "was expecting a validation error")
	assert.NoError(input.Nullable.Validate())         // should be valid
	assert.Error(input.Optional.Validate())           // should NOT be valid
	assert.NoError(input.OptionalNullable.Validate()) // should be valid
	assert.Error(input.validate())
}

func TestIntWithNotSet(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := emptyJSONStruct
	input := &SampleIntInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.Required.Validate())           // should NOT be valid
	assert.Error(input.Nullable.Validate())           // should NOT be valid
	assert.NoError(input.Optional.Validate())         // should be valid
	assert.NoError(input.OptionalNullable.Validate()) // should be valid
	assert.Error(input.validate())
}
