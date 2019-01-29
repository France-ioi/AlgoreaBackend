package types

import (
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestValidateErrorMsg(t *testing.T) {
	assert := assertlib.New(t)
	listSizeErrorMsg := "the number of names should match the number of values for validation"

	exists := RequiredBool{Bool{true, true, false}}
	name := OptionalString{String{"John Doe", true, false}}
	wrongExists := RequiredBool{Bool{true, false, false}}

	assert.EqualError(Validate([]string{"exists"}, &exists, &name), listSizeErrorMsg)
	assert.EqualError(Validate([]string{"exists", "name", "address"}, &exists, &name), listSizeErrorMsg)
	assert.EqualError(Validate([]string{"exists", "name"}, &wrongExists, &name), "wrong value for 'exists': must be given and not null")
}
