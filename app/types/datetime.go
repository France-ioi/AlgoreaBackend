package types

import (
	"reflect"
	"time"
)

// Doc is mainly in the "value" file :-)

type (
	// Datetime is a date+time format which can be set/not-set and null/not-null
	Datetime struct{ Data }
	// RequiredDatetime must be set and not null
	RequiredDatetime struct{ Datetime }
	// NullableDatetime must be set and can be null
	NullableDatetime struct{ Datetime }
	// OptionalDatetime can be not set. If set, cannot be null.
	OptionalDatetime struct{ Datetime }
	// OptNullDatetime can be not set or null
	OptNullDatetime struct{ Datetime }
)

// NewDatetime creates a Datetime which is not-null and set with the given Value
func NewDatetime(t time.Time) *Datetime {
	n := &Datetime{Data{Value: t, Set: true, Null: false}}
	return n
}

// UnmarshalJSON parse JSON data to the type
func (t *Datetime) UnmarshalJSON(data []byte) (err error) {
	return unmarshalJSON(data, &t.Set, &t.Null, &t.Value, reflect.TypeOf(time.Time{}))
}

// Validate checks that the subject matches "required" (set and not-null)
func (t *RequiredDatetime) Validate() error {
	return validateRequired(t.Set, t.Null)
}

// Validate checks that the subject matches "nullable" (must be set)
func (t *NullableDatetime) Validate() error {
	return validateNullable(t.Set)
}

// Validate checks that the subject matches "optional" (not-null)
func (t *OptionalDatetime) Validate() error {
	return validateOptional(t.Null)
}

// Validate checks that the subject matches "optnull" (always true)
func (t *OptNullDatetime) Validate() error {
	return nil
}
