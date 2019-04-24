package types

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	assertlib "github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	timeVal := time.Date(2001, time.February, 3, 5, 6, 7, 890000000, time.UTC)

	tests := []struct {
		valueType     string
		funcToCall    func() AllAttributeser
		expectedValue interface{}
	}{
		{valueType: "Bool", funcToCall: func() AllAttributeser {
			return NewBool(true)
		}, expectedValue: true},
		{valueType: "Datetime", funcToCall: func() AllAttributeser {
			return NewDatetime(timeVal)
		}, expectedValue: timeVal},
		{valueType: "Int32", funcToCall: func() AllAttributeser {
			return NewInt32(2147483647)
		}, expectedValue: int32(2147483647)},
		{valueType: "Int64", funcToCall: func() AllAttributeser {
			return NewInt64(2147483645)
		}, expectedValue: int64(2147483645)},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.valueType, func(t *testing.T) {
			assert := assertlib.New(t)

			n := testCase.funcToCall()
			val, null, set := n.AllAttributes()
			assert.Equal(testCase.expectedValue, reflect.ValueOf(n).Elem().FieldByName("Value").Interface())
			assert.Equal(testCase.expectedValue, val)
			assert.True(reflect.ValueOf(n).Elem().FieldByName("Set").Interface().(bool))
			assert.True(set)
			assert.False(reflect.ValueOf(n).Elem().FieldByName("Null").Interface().(bool))
			assert.False(null)
		})
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name                          string
		target                        validatable
		jsonInput                     string
		expectedRequiredValue         interface{}
		expectedNullableValue         interface{}
		expectedOptionalValue         interface{}
		expectedOptionalNullableValue interface{}
	}{
		{
			name:                          "Bool",
			target:                        &SampleBoolInput{},
			jsonInput:                     `{ "Required": true, "Nullable": false, "Optional": true, "OptionalNullable": true}`,
			expectedRequiredValue:         true,
			expectedNullableValue:         false,
			expectedOptionalValue:         true,
			expectedOptionalNullableValue: true,
		},
		{
			name:                          "Int32",
			target:                        &SampleInt32Input{},
			jsonInput:                     `{ "Required": 2147483647, "Nullable": 22, "Optional": -1, "OptionalNullable": 7 }`,
			expectedRequiredValue:         int32(2147483647),
			expectedNullableValue:         int32(22),
			expectedOptionalValue:         int32(-1),
			expectedOptionalNullableValue: int32(7),
		},
		{
			name:                          "Int64",
			target:                        &SampleInt64Input{},
			jsonInput:                     `{ "Required": "2147483645", "Nullable": "22", "Optional": "-1", "OptionalNullable": "7" }`,
			expectedRequiredValue:         int64(2147483645),
			expectedNullableValue:         int64(22),
			expectedOptionalValue:         int64(-1),
			expectedOptionalNullableValue: int64(7),
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert := assertlib.New(t)

			assert.NoError(json.Unmarshal([]byte(testCase.jsonInput), testCase.target))
			reflTarget := reflect.ValueOf(testCase.target).Elem()
			assert.Equal(testCase.expectedRequiredValue,
				reflTarget.FieldByName("Required").FieldByName("Value").Interface())
			assert.Equal(testCase.expectedNullableValue,
				reflTarget.FieldByName("Nullable").FieldByName("Value").Interface())
			assert.Equal(testCase.expectedOptionalValue,
				reflTarget.FieldByName("Optional").FieldByName("Value").Interface())
			assert.Equal(testCase.expectedOptionalNullableValue,
				reflTarget.FieldByName("OptionalNullable").FieldByName("Value").Interface())
			assert.NoError(testCase.target.Validate())
		})
	}
}

func TestWithIncorrectValues(t *testing.T) {
	tests := []struct {
		name      string
		target    interface{}
		jsonInput string
	}{
		{
			name:      "Bool",
			target:    &SampleBoolInput{},
			jsonInput: `{ "Required": 1234, "Nullable": true, "Optional": false, "OptionalNullable": true }`,
		},
		{
			name:   "DateTime",
			target: &SampleDTInput{},
			jsonInput: `{ "Required": "2001", "Nullable": "2002-01-01T23:11:11Z", "Optional": "2001-09-02T12:30Z", ` +
				`"OptionalNullable": "2042-12-31T23:59:59Z" }`,
		},
		{
			name:      "Int32",
			target:    &SampleInt32Input{},
			jsonInput: `{ "Required": "not an int", "Nullable": 22, "Optional": -1, "OptionalNullable": 7 }`,
		},
		{
			name:      "Int64",
			target:    &SampleInt64Input{},
			jsonInput: `{ "Required": "not an int", "Nullable": "22", "Optional": "-1", "OptionalNullable": "7" }`,
		},
		{
			name:      "String",
			target:    &SampleStrInput{},
			jsonInput: `{ "Required": 1234, "Nullable": "From Journeyman to Master", "Optional": "Andy Hunt", "OptionalNullable": "John Doe" }`,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert := assertlib.New(t)
			assert.Error(json.Unmarshal([]byte(testCase.jsonInput), testCase.target))
		})
	}
}

func TestWithDefault(t *testing.T) {
	tests := []struct {
		name      string
		target    validatable
		jsonInput string
	}{
		{
			name:      "Bool",
			target:    &SampleBoolInput{},
			jsonInput: `{ "Required": false, "Nullable": false, "Optional": false, "OptionalNullable": false}`,
		},
		{
			name:      "Int32",
			target:    &SampleInt32Input{},
			jsonInput: `{ "Required": 0, "Nullable": 0, "Optional": 0, "OptionalNullable": 0 }`,
		},
		{
			name:      "Int64",
			target:    &SampleInt64Input{},
			jsonInput: `{ "Required": "0", "Nullable": "0", "Optional": "0", "OptionalNullable": "0" }`,
		},
		{
			name:      "String",
			target:    &SampleStrInput{},
			jsonInput: `{ "Required": "", "Nullable": "", "Optional": "", "OptionalNullable": "" }`,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert := assertlib.New(t)

			assert.NoError(json.Unmarshal([]byte(testCase.jsonInput), testCase.target))
			assert.NoError(testCase.target.Validate())
		})
	}
}

func TestWithNull(t *testing.T) {
	tests := []struct {
		name   string
		target validatable
	}{
		{name: "Bool", target: &SampleBoolInput{}},
		{name: "Int32", target: &SampleInt32Input{}},
		{name: "Int64", target: &SampleInt64Input{}},
		{name: "String", target: &SampleStrInput{}},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert := assertlib.New(t)

			assert.NoError(json.Unmarshal([]byte(allNullsJSONStruct), testCase.target))
			reflTarget := reflect.ValueOf(testCase.target).Elem()
			assert.Error(reflTarget.FieldByName("Required").Addr().Interface().(validatable).Validate(),
				"was expecting a validation error")
			assert.NoError(reflTarget.FieldByName("Nullable").Addr().Interface().(validatable).Validate())         // should be valid
			assert.Error(reflTarget.FieldByName("Optional").Addr().Interface().(validatable).Validate())           // should NOT be valid
			assert.NoError(reflTarget.FieldByName("OptionalNullable").Addr().Interface().(validatable).Validate()) // should be valid
			assert.Error(testCase.target.Validate())
		})
	}
}

func TestWithNotSet(t *testing.T) {
	tests := []struct {
		name   string
		target validatable
	}{
		{name: "Bool", target: &SampleBoolInput{}},
		{name: "Int32", target: &SampleInt32Input{}},
		{name: "Int64", target: &SampleInt64Input{}},
		{name: "String", target: &SampleStrInput{}},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			assert := assertlib.New(t)

			assert.NoError(json.Unmarshal([]byte(emptyJSONStruct), testCase.target))
			reflTarget := reflect.ValueOf(testCase.target).Elem()
			assert.Error(reflTarget.FieldByName("Required").Addr().Interface().(validatable).Validate())           // should NOT be valid
			assert.Error(reflTarget.FieldByName("Nullable").Addr().Interface().(validatable).Validate())           // should NOT be valid
			assert.NoError(reflTarget.FieldByName("Optional").Addr().Interface().(validatable).Validate())         // should be valid
			assert.NoError(reflTarget.FieldByName("OptionalNullable").Addr().Interface().(validatable).Validate()) // should be valid
			assert.Error(testCase.target.Validate())
		})
	}
}
