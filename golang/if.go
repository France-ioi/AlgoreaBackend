package golang

// IfElse returns trueValue if condition is true, otherwise falseValue.
func IfElse[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// If returns trueValue if condition is true, otherwise the zero value of T.
func If[T any](condition bool, trueValue T) T {
	if condition {
		return trueValue
	}
	return Zero[T]()
}

// LazyIfElse returns the result of trueValueFunc if condition is true, otherwise the result of falseValueFunc.
// The functions are only called if their result is needed.
func LazyIfElse[T any](condition bool, trueValueFunc, falseValueFunc func() T) T {
	if condition {
		return trueValueFunc()
	}
	return falseValueFunc()
}

// LazyIf returns the result of trueValueFunc if condition is true, otherwise the zero value of T.
// The function is only called if its result is needed.
func LazyIf[T any](condition bool, trueValueFunc func() T) T {
	if condition {
		return trueValueFunc()
	}
	return Zero[T]()
}
