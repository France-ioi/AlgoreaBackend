package golang

// Ptr returns a pointer to a literal.
// Use for initializations.
// e.g., var pb *bool = Ptr(true).
func Ptr[T any](v T) *T {
	return &v
}
