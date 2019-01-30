package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Doc is mainly in the "int64" file :-)

type (
	// String is a string which can be set/not-set and null/not-null
	String struct {
		Value string
		Set   bool
		Null  bool
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
func (s *String) UnmarshalJSON(data []byte) (err error) {
	s.Set = true // If this method was called, the value was set.
	s.Null = string(data) == jsonNull
	var temp string
	err = json.Unmarshal(data, &temp)
	if err == nil {
		s.Value = temp
	}
	return
}

// Scan converts GORM types
func (s *String) Scan(value interface{}) (err error) {
	if value == nil {
		s.Value, s.Set, s.Null = "", false, true
		return
	}

	switch v := value.(type) {
	case string:
	case []rune:
	case []uint8:
		s.Value, s.Set, s.Null = string(v), true, false
		return
	}

	s.Value, s.Set, s.Null = "", false, true
	return fmt.Errorf("failed to convert %T to String", value)
}

// AllAttributes unwrap the wrapped value and its attributes
func (s String) AllAttributes() (value interface{}, isNull bool, isSet bool) {
	return s.Value, s.Null, s.Set
}

// Validate checks that the subject matches "required" (set and not-null)
func (s *RequiredString) Validate() error {
	if !s.Set || s.Null {
		return errors.New("must be given and not null")
	}
	return nil
}

// Validate checks that the subject matches "nullable" (must be set)
func (s *NullableString) Validate() error {
	if !s.Set {
		return errors.New("must be given")
	}
	return nil
}

// Validate checks that the subject matches "optional" (not-null)
func (s *OptionalString) Validate() error {
	if s.Null {
		return errors.New("must not be null")
	}
	return nil
}

// Validate checks that the subject matches "optnull" (always true)
func (s *OptNullString) Validate() error {
	return nil
}
