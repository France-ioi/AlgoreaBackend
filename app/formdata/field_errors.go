package formdata

// FieldErrors represents multiple errors for form fields
type FieldErrors map[string][]string

func (e FieldErrors) Error() string {
	return "invalid input data"
}
