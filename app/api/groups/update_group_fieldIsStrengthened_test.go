package groups

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockFieldLevel struct{}

func (m *MockFieldLevel) Top() reflect.Value {
	panic("implement me")
}

func (m *MockFieldLevel) Parent() reflect.Value {
	panic("implement me")
}

func (m *MockFieldLevel) StructFieldName() string {
	panic("implement me")
}

func (m *MockFieldLevel) Path() string {
	panic("implement me")
}

func (m *MockFieldLevel) StructPath() string {
	panic("implement me")
}

func (m *MockFieldLevel) Param() string {
	panic("implement me")
}

func (m *MockFieldLevel) ExtractType(reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool) {
	panic("implement me")
}

func (m *MockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	panic("implement me")
}

func (m *MockFieldLevel) Field() reflect.Value {
	panic("implement me")
}

func (m *MockFieldLevel) FieldName() string {
	return "notRequireField"
}

func Test_fieldIsStrengthened_shouldReturnFalseForNonRequireField(t *testing.T) {
	// This test is only included for 100% coverage, to reach the end of the function.

	mockFieldLevel := new(MockFieldLevel)

	assert.False(t, fieldIsStrengthened(mockFieldLevel, true, &groupUpdateInput{}))
}
