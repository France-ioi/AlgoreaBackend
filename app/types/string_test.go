package types

import (
	"encoding/json"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

type SampleStrInput struct {
	Required         RequiredString
	Nullable         NullableString
	Optional         OptionalString
	OptionalNullable OptNullString
}

func (v *SampleStrInput) Validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}

func TestNewString(t *testing.T) {
	assert := assertlib.New(t)
	var value = "Foo"
	n := NewString(value)
	val, null, set := n.AllAttributes()
	assert.Equal(value, n.Value)
	assert.Equal(value, val)
	assert.True(n.Set)
	assert.True(set)
	assert.False(n.Null)
	assert.False(null)
}

func TestStrValid(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "The Pragmatic Programmer", "Nullable": "From Journeyman to Master", ` +
		`"Optional": "Andy Hunt", "OptionalNullable": "John Doe" }`
	input := &SampleStrInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Equal("The Pragmatic Programmer", input.Required.Value)
	assert.Equal("From Journeyman to Master", input.Nullable.Value)
	assert.Equal("Andy Hunt", input.Optional.Value)
	assert.Equal("John Doe", input.OptionalNullable.Value)
	assert.NoError(input.Validate())
}
