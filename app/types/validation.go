package types

import (
	"errors"
)

// OptionalType is an abstract type which is extended by other types with a value
type OptionalType struct {
	Value interface{}
	Set   bool
	Null  bool
}

// Validatable is the interface indicating the type implementing it supports data validation.
type validatable interface {
	// Validate validates the data and returns an error if validation fails.
	Validate() error
}

// Validate checks a set of `Validatable` values and returns the first encountered error, or nil
func Validate(values ...validatable) error {
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON is the generic part of the unmarshalling
func (v *OptionalType) UnmarshalJSON(data []byte) error {
	v.Set = true // If this method was called, the value was set.
	v.Null = (string(data) == "null")
	return nil
}

func (v *OptionalType) validateRequired() error {
	if !v.Set || v.Null {
		return errors.New("must be given and not null")
	}
	return nil
}

func (v *OptionalType) validateNullable() error {
	if !v.Set {
		return errors.New("must be given")
	}
	return nil
}

func (v *OptionalType) validateOptional() error {
	if v.Null {
		return errors.New("must not be null")
	}
	return nil
}
