package types

import "reflect"

type (
	// String is a string which can be set/not-set and null/not-null
	String struct{ Data }
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
	n := &String{Data{Value: s, Set: true, Null: false}}
	return n
}

// UnmarshalJSON parse JSON data to the type
func (s *String) UnmarshalJSON(data []byte) (err error) {
	return unmarshalJSON(data, &s.Set, &s.Null, &s.Value, reflect.TypeOf(""))
}

// Validate checks that the subject matches "required" (set and not-null)
func (s *RequiredString) Validate() error {
	return validateRequired(s.Set, s.Null)
}

// Validate checks that the subject matches "nullable" (must be set)
func (s *NullableString) Validate() error {
	return validateNullable(s.Set)
}

// Validate checks that the subject matches "optional" (not-null)
func (s *OptionalString) Validate() error {
	return validateOptional(s.Null)
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullString) Validate() error {
	return nil
}
