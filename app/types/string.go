package types

import (
	"encoding/json"
)

// Doc is mainly in the "int64" file :-)

type (
	// String is a string which can be set/not-set and null/not-null
	String struct {
		Value string
		OptionalType
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

// NewString creates a String which is not-null and set with the given value
func NewString(s string) *String {
	n := &String{}
	n.Value = s
	n.Set = true
	n.Null = false
	return n
}

// UnmarshalJSON parse JSON data to the type
func (i *String) UnmarshalJSON(data []byte) error {
	i.OptionalType.UnmarshalJSON(data)
	var temp string
	err := json.Unmarshal(data, &temp)
	if err == nil {
		i.Value = temp
	}
	return err
}

// Validate checks that the subject matches "required" (set and not-null)
func (i *RequiredString) Validate() error {
	return i.OptionalType.validateRequired()
}

// Validate checks that the subject matches "nullable" (must be set)
func (i *NullableString) Validate() error {
	return i.OptionalType.validateNullable()
}

// Validate checks that the subject matches "optional" (not-null)
func (i *OptionalString) Validate() error {
	return i.OptionalType.validateOptional()
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullString) Validate() error {
	return nil
}
