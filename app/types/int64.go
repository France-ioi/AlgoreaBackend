package types

import (
  "encoding/json"
  "errors"
)

// Important notes on all these custom types:
// All types here are optional because you cannot ask the Go unmarshaller
// to fail on value not set. However the 'set' flag will be not set if the value has
// not been given.
// For failing on optional, it has be done at the struct validation level

type (
  // Int64 is an integer which can be set/not-set and null/not-null
  Int64 struct {
    Value int64
    Set   bool
    Null  bool
  }
  // RequiredInt64 must be set and not null
  RequiredInt64 struct{ Int64 }
  // NullableInt64 must be set and can be null
  NullableInt64 struct{ Int64 }
  // OptionalInt64 can be not set. If set, cannot be null.
  OptionalInt64 struct{ Int64 }
  // OptNullInt64 can be not set or null
  OptNullInt64 struct{ Int64 }
)

// NewInt64 creates a Int64 which is not-null and set with the given value
func NewInt64(v int64) *Int64 {
  n := &Int64{}
  n.Value = v
  n.Set = true
  n.Null = false
  return n
}

// UnmarshalJSON parse JSON data to the type
func (i *Int64) UnmarshalJSON(data []byte) (err error) {
  i.Set = true // If this method was called, the value was set.
  i.Null = (string(data) == "null")
  var temp int64
  err = json.Unmarshal(data, &temp)
  if err == nil {
    i.Value = temp
  }
  return
}

// Validate checks that the subject matches "required" (set and not-null)
func (i *RequiredInt64) Validate() error {
  if !i.Set || i.Null {
    return errors.New("must be given and not null")
  }
  return nil
}

// Validate checks that the subject matches "nullable" (must be set)
func (i *NullableInt64) Validate() error {
  if !i.Set {
    return errors.New("must be given")
  }
  return nil
}

// Validate checks that the subject matches "optional" (not-null)
func (i *OptionalInt64) Validate() error {
  if i.Null {
    return errors.New("must not be null")
  }
  return nil
}

// Validate checks that the subject matches "optnull" (always true)
func (i *OptNullInt64) Validate() error {
  return nil
}
