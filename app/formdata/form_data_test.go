package formdata

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/France-ioi/mapstructure"
	"github.com/France-ioi/validator"
	ut "github.com/go-playground/universal-translator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func TestFormData_IsSet(t *testing.T) {
	formData := &FormData{usedKeys: map[string]bool{"usedField": true}}
	assert.True(t, formData.IsSet("usedField"))
	assert.False(t, formData.IsSet("otherField"))
}

func TestFormData_IsValid(t *testing.T) {
	formData := &FormData{fieldErrors: FieldErrorsError{"fieldWithErrors": []string{"someError"}}}
	assert.True(t, formData.IsValid("someField"))
	assert.False(t, formData.IsValid("fieldWithErrors"))
}

func TestFormData_AllowUnknownFields(t *testing.T) {
	formData := NewFormData(&struct {
		ID int64 `json:"id"`
	}{})
	formData.AllowUnknownFields()
	assert.NoError(t, formData.ParseMapData(map[string]interface{}{"my_id": "123"}))
}

func TestFormData_decodeMapIntoStruct_PanicsWhenMapstructureNewDecoderFails(t *testing.T) {
	formData := &FormData{}
	defer func() {
		p := recover()
		assert.NotNil(t, p)
		assert.Equal(t, errors.New("result must be a pointer"), p)
	}()
	formData.decodeMapIntoStruct(map[string]interface{}{})
}

func TestFormData_RegisterTranslation_OverridesConflictingTranslationsSilently(t *testing.T) {
	f := NewFormData(&struct{}{})
	f.RegisterTranslation("", "")
	assert.NotPanics(t, func() {
		f.RegisterTranslation("", "")
	})
}

func TestFormData_RegisterTranslation_PanicsOnError(t *testing.T) {
	formData := NewFormData(&struct{}{})
	defer func() {
		p := recover()
		assert.NotNil(t, p)
		assert.IsType(t, (*ut.ErrMissingBracket)(nil), p)
	}()
	formData.RegisterTranslation("", "{")
}

func TestFormData_RegisterTranslation_SetsArgumentsForErrorMessages(t *testing.T) {
	formData := NewFormData(&struct {
		ID int64 `validate:"test=value"`
	}{})
	formData.RegisterValidation("test", func(validator.FieldLevel) bool {
		return false
	})
	formData.RegisterTranslation("test", "failed for field %[2]s with parameter %[3]s (tag=%[1]s)")
	err := formData.ParseMapData(map[string]interface{}{"id": int64(1)})
	assert.Equal(t, FieldErrorsError{
		"ID": []string{"failed for field ID with parameter value (tag=test)"},
	}, err)
}

func TestFormData_ValidatorSkippingUnsetFields(t *testing.T) {
	formData := NewFormData(&struct {
		ID     *int64 `validate:"test"`
		Nested *struct {
			ID *int64 `validate:"test"`
		} `validate:"test"`
	}{})
	formData.RegisterValidation("test", formData.ValidatorSkippingUnsetFields(func(validator.FieldLevel) bool {
		return false
	}))
	formData.RegisterTranslation("test", "failed")
	err := formData.ParseMapData(map[string]interface{}{})
	require.NoError(t, err)
	err = formData.ParseMapData(map[string]interface{}{"id": nil, "nested": nil})
	assert.Equal(t, FieldErrorsError{"ID": []string{"failed"}, "Nested": []string{"failed"}}, err)
	err = formData.ParseMapData(map[string]interface{}{"id": nil, "nested": map[string]interface{}{"id": 1}})
	assert.Equal(t, FieldErrorsError{"ID": []string{"failed"}, "Nested": []string{"failed"}, "Nested.ID": []string{"failed"}}, err)
}

func TestFormData_ValidatorSkippingUnchangedFields(t *testing.T) {
	type nestedStruct struct {
		ID *int64 `validate:"test"`
	}
	type testStruct struct {
		ID     *int64        `validate:"test"`
		Nested *nestedStruct `validate:"test"`
	}

	formData := NewFormData(&testStruct{})
	formData.RegisterValidation("test", formData.ValidatorSkippingUnchangedFields(func(validator.FieldLevel) bool {
		return false
	}))
	formData.RegisterTranslation("test", "failed")

	err := formData.ParseMapData(map[string]interface{}{})
	require.NoError(t, err)
	formData.SetOldValues(&testStruct{})
	err = formData.ParseMapData(map[string]interface{}{})
	require.NoError(t, err)
	err = formData.ParseMapData(map[string]interface{}{"id": nil, "nested": map[string]interface{}{"id": 1}})
	assert.Equal(t, FieldErrorsError{"Nested": []string{"failed"}, "Nested.ID": []string{"failed"}}, err)
	i := int64(10)
	j := int64(20)
	formData.SetOldValues(&testStruct{ID: &i, Nested: &nestedStruct{ID: &j}})
	err = formData.ParseMapData(map[string]interface{}{})
	require.NoError(t, err)
	err = formData.ParseMapData(map[string]interface{}{"id": nil, "nested": map[string]interface{}{"id": nil}})
	assert.Equal(t, FieldErrorsError{"ID": []string{"failed"}, "Nested": []string{"failed"}, "Nested.ID": []string{"failed"}}, err)
	err = formData.ParseMapData(map[string]interface{}{"id": 10, "nested": map[string]interface{}{"id": 20}})
	assert.Equal(t, FieldErrorsError{"Nested": []string{"failed"}}, err)
	formData.SetOldValues(nil)
	err = formData.ParseMapData(map[string]interface{}{"id": 10, "nested": map[string]interface{}{"id": 20}})
	assert.Equal(t, FieldErrorsError{"ID": []string{"failed"}, "Nested": []string{"failed"}, "Nested.ID": []string{"failed"}}, err)
}

func Test_stringToDatabaseTimeUTCHookFunc(t *testing.T) {
	tests := []struct {
		name     string
		typeFrom reflect.Type
		typeTo   reflect.Type
		data     interface{}
		want     interface{}
		wantErr  error
	}{
		{
			name:     "string to database.Time (parse)",
			typeFrom: reflect.TypeOf("string"),
			typeTo:   reflect.TypeOf((*database.Time)(nil)).Elem(),
			data:     "2019-05-30T14:00:00+03:00",
			want:     database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC)),
		},
		{
			name:     "invalid string to database.Time (error)",
			typeFrom: reflect.TypeOf("string"),
			typeTo:   reflect.TypeOf((*database.Time)(nil)).Elem(),
			data:     "2019-05-30T14:00:00ZZ",
			want:     database.Time(time.Time{}),
			wantErr: &time.ParseError{
				Layout:     time.RFC3339,
				Value:      "2019-05-30T14:00:00ZZ",
				LayoutElem: "",
				ValueElem:  "Z",
				Message:    ": extra text: \"Z\"",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hook := stringToDatabaseTimeUTCHookFunc(time.RFC3339)
			converted, err := mapstructure.DecodeHookExec(hook, tt.typeFrom, tt.typeTo, tt.data)
			if tt.wantErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.want, converted)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestFormData_fieldPathInValidator(t *testing.T) {
	type TestNestedNestedStruct struct {
		Field *string `json:"field" validate:"custom|custom1"`
	}
	type TestNestedStruct struct {
		TestNestedNestedStruct `json:"nested_nested,squash"`
	}
	type testStruct struct {
		TestNestedStruct `json:"nested,squash"`
	}

	var path, path1 string
	var usedKeysPath, usedKeysPath1 string
	var isSet, isSet1 bool
	formData := NewFormData(&testStruct{})
	formData.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
		path = fl.Path()
		usedKeysPath = formData.getUsedKeysPathFromValidatorPath(path)
		isSet = formData.IsSet(usedKeysPath)
		return false
	})
	formData.RegisterValidation("custom1", func(fl validator.FieldLevel) bool {
		path1 = fl.Path()
		usedKeysPath1 = formData.getUsedKeysPathFromValidatorPath(path)
		isSet1 = formData.IsSet(usedKeysPath)
		return true
	})
	_ = formData.ParseMapData(map[string]interface{}{"field": ""})
	assert.Equal(t, "testStruct.<squash>.<squash>.field", path)
	assert.Equal(t, "field", usedKeysPath)
	assert.True(t, isSet)
	assert.Equal(t, "testStruct.<squash>.<squash>.field", path1)
	assert.Equal(t, "field", usedKeysPath1)
	assert.True(t, isSet1)
}
