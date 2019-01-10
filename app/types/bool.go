package types

import (
	"encoding/json"
	"errors"
)

// Doc is mainly in the "int64" file :-)

type (
	// Bool is a bool which can be set/not-set and null/not-null
	Bool struct {
		Value bool
		Set   bool
		Null  bool
	}
	// RequiredBool must be set and not null
	RequiredBool struct{ Bool }
	// NullableBool must be set and can be null
	NullableBool struct{ Bool }
	// OptionalBool can be not set. If set, cannot be null.
	OptionalBool struct{ Bool }
	// OptNullBool can be not set or null
	OptNullBool struct{ Bool }
)

// NewBool creates a Bool which is not-null and set with the given value
func NewBool(s bool) *Bool {
	n := &Bool{}
	n.Value = s
	n.Set = true
	n.Null = false
	return n
}

// UnmarshalJSON parse JSON data to the type
func (s *Bool) UnmarshalJSON(data []byte) (err error) {
	s.Set = true // If this method was called, the value was set.
	s.Null = (string(data) == jsonNull)
	var temp bool
	err = json.Unmarshal(data, &temp)
	if err == nil {
		s.Value = temp
	}
	return
}

// Validate checks that the subject matches "required" (set and not-null)
func (s *RequiredBool) Validate() error {
	if !s.Set || s.Null {
		return errors.New("must be given and not null")
	}
	return nil
}

// Validate checks that the subject matches "nullable" (must be set)
func (s *NullableBool) Validate() error {
	if !s.Set {
		return errors.New("must be given")
	}
	return nil
}

// Validate checks that the subject matches "optional" (not-null)
func (s *OptionalBool) Validate() error {
	if s.Null {
		return errors.New("must not be null")
	}
	return nil
}

// Validate checks that the subject matches "optnull" (always true)
func (s *OptNullBool) Validate() error {
	return nil
}
