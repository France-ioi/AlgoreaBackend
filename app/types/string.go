package types

import (
	"encoding/json"
)

// Doc is mainly in the "int64" file :-)

type (
	// String is an abstract type for the types below
	String struct {
		Value int64
		validatableType
	}
	// RequiredString must be set and not null
	RequiredString struct{ String }
	// NullableString must be set and can be null
	NullableString struct{ String }
	// OptionalString can be not set. If set, cannot be null.
	OptionalString struct{ String }
	// OptNullString can be not set or null
	OptNullString struct{ String }
)

// UnmarshalJSON parse JSON data to the type
func (i *String) UnmarshalJSON(data []byte) error {
	i.validatableType.UnmarshalJSON(data)
	var temp int64
	err := json.Unmarshal(data, &temp)
	if err == nil {
		i.Value = temp
	}
	return err
}

// Validate checks that the subject matches "required" (set and not-null)
func (i *RequiredString) Validate() error {
	return i.validatableType.validateRequired()
}

// Validate checks that the subject matches "nullable" (must be set)
func (i *NullableString) Validate() error {
	return i.validatableType.validateNullable()
}

// Validate checks that the subject matches "optional" (not-null)
func (i *OptionalString) Validate() error {
	return i.validatableType.validateOptional()
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullString) Validate() error {
	return nil
}
