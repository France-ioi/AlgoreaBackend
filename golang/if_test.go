package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfElse(t *testing.T) {
	assert.Equal(t, 1, IfElse(true, 1, 2))
	assert.Equal(t, 2, IfElse(false, 1, 2))
	assert.Equal(t, "a", IfElse(true, "a", "b"))
	assert.Equal(t, "b", IfElse(false, "a", "b"))
	assert.True(t, IfElse(true, true, false))
}

func TestIf(t *testing.T) {
	str := "str"
	strPtr := &str

	assert.Equal(t, 1, If(true, 1))
	assert.Equal(t, 0, If(false, 1))
	assert.Equal(t, "a", If(true, "a"))
	assert.Equal(t, "", If(false, "a"))
	assert.True(t, If(true, true))
	assert.False(t, If(false, true))
	assert.Equal(t, (*string)(nil), If(false, strPtr))
}

func TestLazyIfElse_TrueValueFuncIsCalled(t *testing.T) {
	var trueValueFuncCalled, falseValueFuncCalled bool
	assert.Equal(t, 1,
		LazyIfElse(true, func() int {
			trueValueFuncCalled = true
			return 1
		}, func() int {
			falseValueFuncCalled = true
			return 2
		}))
	assert.True(t, trueValueFuncCalled)
	assert.False(t, falseValueFuncCalled)
}

func TestLazyIfElse_FalseValueFuncIsCalled(t *testing.T) {
	var trueValueFuncCalled, falseValueFuncCalled bool
	assert.Equal(t, 2,
		LazyIfElse(false, func() int {
			trueValueFuncCalled = true
			return 1
		}, func() int {
			falseValueFuncCalled = true
			return 2
		}))
	assert.False(t, trueValueFuncCalled)
	assert.True(t, falseValueFuncCalled)
}

func TestLazyIf_TrueValueFuncIsCalled(t *testing.T) {
	var trueValueFuncCalled bool
	assert.Equal(t, 1,
		LazyIf(true, func() int {
			trueValueFuncCalled = true
			return 1
		}))
	assert.True(t, trueValueFuncCalled)
}

func TestLazyIf_TrueValueFuncIsNotCalled(t *testing.T) {
	var trueValueFuncCalled bool
	assert.Equal(t, "",
		LazyIf(false, func() string {
			trueValueFuncCalled = true
			return "a"
		}))
	assert.False(t, trueValueFuncCalled)
}

func TestLazyIf_TrueValueFuncIsNotCalled_ZeroPointer(t *testing.T) {
	var trueValueFuncCalled bool
	assert.Equal(t, (*string)(nil),
		LazyIf(false, func() *string {
			trueValueFuncCalled = true
			result := "a"
			return &result
		}))
	assert.False(t, trueValueFuncCalled)
}
