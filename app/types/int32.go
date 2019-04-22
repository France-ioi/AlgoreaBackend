package types

import (
	"encoding/json"
	"errors"
	"unsafe"
)

// Important notes on all these custom types:
// All types here are optional because you cannot ask the Go unmarshaller
// to fail on value not set. However the 'set' flag will be not set if the value has
// not been given.
// For failing on optional, it has be done at the struct validation level

type (
	// Int32 is an integer which can be set/not-set and null/not-null
	Int32 struct {
		Value int32
		Set   bool
		Null  bool
	}
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
	n := &Int32{}
	n.Value = v
	n.Set = true
	n.Null = false
	return n
}

// UnmarshalJSON parse JSON data to the type
func (i *Int32) UnmarshalJSON(data []byte) (err error) {
	i.Set = true // If this method was called, the value was set.
	i.Null = *(*string)(unsafe.Pointer(&data)) == jsonNull
	var temp int32
	err = json.Unmarshal(data, &temp)
	if err == nil {
		i.Value = temp
	}
	return
}

// AllAttributes unwrap the wrapped value and its attributes
func (i Int32) AllAttributes() (value interface{}, isNull, isSet bool) {
	return i.Value, i.Null, i.Set
}

// Validate checks that the subject matches "required" (set and not-null)
func (i *RequiredInt32) Validate() error {
	if !i.Set || i.Null {
		return errors.New("must be given and not null")
	}
	return nil
}

// Validate checks that the subject matches "nullable" (must be set)
func (i *NullableInt32) Validate() error {
	if !i.Set {
		return errors.New("must be given")
	}
	return nil
}

// Validate checks that the subject matches "optional" (not-null)
func (i *OptionalInt32) Validate() error {
	if i.Null {
		return errors.New("must not be null")
	}
	return nil
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullInt32) Validate() error {
	return nil
}
