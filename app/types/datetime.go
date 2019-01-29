package types

import (
	"encoding/json"
	"errors"
	"time"
)

// Doc is mainly in the "int64" file :-)

type (
	// Datetime is a date+time format which can be set/not-set and null/not-null
	Datetime struct {
		Value time.Time
		Set   bool
		Null  bool
	}
	// RequiredDatetime must be set and not null
	RequiredDatetime struct{ Datetime }
	// NullableDatetime must be set and can be null
	NullableDatetime struct{ Datetime }
	// OptionalDatetime can be not set. If set, cannot be null.
	OptionalDatetime struct{ Datetime }
	// OptNullDatetime can be not set or null
	OptNullDatetime struct{ Datetime }
)

// NewDatetime creates a Datetime which is not-null and set with the given value
func NewDatetime(t time.Time) *Datetime {
	n := &Datetime{}
	n.Value = t
	n.Set = true
	n.Null = false
	return n
}

// UnmarshalJSON parse JSON data to the type
func (t *Datetime) UnmarshalJSON(data []byte) (err error) {
	t.Set = true // If this method was called, the value was set.
	t.Null = string(data) == jsonNull
	var temp time.Time
	err = json.Unmarshal(data, &temp)
	if err == nil {
		t.Value = temp
	}
	return
}

// AllAttributes unwrap the wrapped value and its attributes
func (t Datetime) AllAttributes() (value interface{}, isNull bool, isSet bool) {
	return t.Value, t.Null, t.Set
}

// Validate checks that the subject matches "required" (set and not-null)
func (t *RequiredDatetime) Validate() error {
	if !t.Set || t.Null {
		return errors.New("must be given and not null")
	}
	return nil
}

// Validate checks that the subject matches "nullable" (must be set)
func (t *NullableDatetime) Validate() error {
	if !t.Set {
		return errors.New("must be given")
	}
	return nil
}

// Validate checks that the subject matches "optional" (not-null)
func (t *OptionalDatetime) Validate() error {
	if t.Null {
		return errors.New("must not be null")
	}
	return nil
}

// Validate checks that the subject matches "optnull" (always true)
func (t *OptNullDatetime) Validate() error {
	return nil
}
