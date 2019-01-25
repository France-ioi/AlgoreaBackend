package types

const jsonNull = "null"

// NullableOptional is a generic interface for all set/null information about these custom types
type NullableOptional interface {
	AllAttributes() (value interface{}, isNull bool, isSet bool)
}

// NOTE: to be probably replace by JSON schema validation and remove

// Validatable is the interface indicating the type implementing it supports data validation.
type validatable interface {
	// Validate validates the data and returns an error if validation fails.
	Validate() error
}

// Validate checks a set of `Validatable` values and returns the first encountered error, or nil
func Validate(values ...validatable) (err error) {
	for _, value := range values {
		if err = value.Validate(); err != nil {
			return
		}
	}
	return
}
