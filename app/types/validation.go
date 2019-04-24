package types

import (
	"errors"
	"fmt"
)

const jsonNull = "null"

// NOTE: to be probably replace by JSON schema validation and remove

// Validatable is the interface indicating the type implementing it supports data validation.
type validatable interface {
	// Validate validates the data and returns an error if validation fails.
	Validate() error
}

// Validate checks a set of `Validatable` values and returns the first encountered error, or nil
func Validate(names []string, values ...validatable) error {
	if len(names) != len(values) {
		return errors.New("the number of names should match the number of values for validation")
	}
	for index, value := range values {
		if err := value.Validate(); err != nil {
			return fmt.Errorf("wrong value for '%s': %s", names[index], err.Error())
		}
	}
	return nil
}

// validateRequired checks that the subject matches "required" (set and not-null)
func validateRequired(isSet, isNull bool) error {
	if !isSet || isNull {
		return errors.New("must be given and not null")
	}
	return nil
}

// validateNullable checks that the subject matches "nullable" (must be set)
func validateNullable(isSet bool) error {
	if !isSet {
		return errors.New("must be given")
	}
	return nil
}

// validateOptional checks that the subject matches "optional" (not-null)
func validateOptional(isNull bool) error {
	if isNull {
		return errors.New("must not be null")
	}
	return nil
}
