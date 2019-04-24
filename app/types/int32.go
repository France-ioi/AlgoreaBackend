package types

import "reflect"

// Important notes on all these custom types:
// All types here are optional because you cannot ask the Go unmarshaller
// to fail on value not set. However the 'set' flag will be not set if the value has
// not been given.
// For failing on optional, it has be done at the struct validation level

type (
	// Int32 is an integer which can be set/not-set and null/not-null
	Int32 struct{ Data }
	// RequiredInt32 must be set and not null
	RequiredInt32 struct{ Int32 }
	// NullableInt32 must be set and can be null
	NullableInt32 struct{ Int32 }
	// OptionalInt32 can be not set. If set, cannot be null.
	OptionalInt32 struct{ Int32 }
	// OptNullInt32 can be not set or null
	OptNullInt32 struct{ Int32 }
)

// NewInt32 creates a Int32 which is not-null and set with the given value
func NewInt32(v int32) *Int32 {
	n := &Int32{Data{Value: v, Set: true, Null: false}}
	return n
}

// UnmarshalJSON parse JSON data to the type
func (i *Int32) UnmarshalJSON(data []byte) (err error) {
	return unmarshalJSON(data, &i.Set, &i.Null, &i.Value, reflect.TypeOf(int32(0)))
}

// Validate checks that the subject matches "required" (set and not-null)
func (i *RequiredInt32) Validate() error {
	return validateRequired(i.Set, i.Null)
}

// Validate checks that the subject matches "nullable" (must be set)
func (i *NullableInt32) Validate() error {
	return validateNullable(i.Set)
}

// Validate checks that the subject matches "optional" (not-null)
func (i *OptionalInt32) Validate() error {
	return validateOptional(i.Null)
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullInt32) Validate() error {
	return nil
}
