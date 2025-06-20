package formdata

// FieldErrorsError represents multiple errors for form fields.
type FieldErrorsError map[string][]string

func (e FieldErrorsError) Error() string {
	return "invalid input data"
}
