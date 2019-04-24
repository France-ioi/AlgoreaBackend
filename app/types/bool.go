package types

import "reflect"

// Doc is mainly in the "int64" file :-)

type (
	// Bool is a bool which can be set/not-set and null/not-null
	Bool struct{ Data }
	// RequiredBool must be set and not null
	RequiredBool struct{ Bool }
	// NullableBool must be set and can be null
	NullableBool struct{ Bool }
	// OptionalBool can be not set. If set, cannot be null.
	OptionalBool struct{ Bool }
	// OptNullBool can be not set or null
	OptNullBool struct{ Bool }
)

// NewBool creates a Bool which is not-null and set with the given Value
func NewBool(s bool) *Bool {
	n := &Bool{Data{Value: s, Set: true, Null: false}}
	return n
}

// UnmarshalJSON parse JSON data to the type
func (s *Bool) UnmarshalJSON(data []byte) (err error) {
	return unmarshalJSON(data, &s.Set, &s.Null, &s.Value, reflect.TypeOf(true))
}

// Validate checks that the subject matches "required" (set and not-null)
func (s *RequiredBool) Validate() error {
	return validateRequired(s.Set, s.Null)
}

// Validate checks that the subject matches "nullable" (must be set)
func (s *NullableBool) Validate() error {
	return validateNullable(s.Set)
}

// Validate checks that the subject matches "optional" (not-null)
func (s *OptionalBool) Validate() error {
	return validateOptional(s.Null)
}

// Validate checks that the subject matches "optnull" (always true)
func (s *OptNullBool) Validate() error {
	return nil
}
