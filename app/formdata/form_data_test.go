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

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestFormData_IsSet(t *testing.T) {
	formData := &FormData{usedKeys: map[string]bool{"usedField": true}}
	assert.True(t, formData.IsSet("usedField"))
	assert.False(t, formData.IsSet("otherField"))
}

func TestFormData_IsValid(t *testing.T) {
	formData := &FormData{fieldErrors: FieldErrors{"fieldWithErrors": []string{"someError"}}}
	assert.True(t, formData.IsValid("someField"))
	assert.False(t, formData.IsValid("fieldWithErrors"))
}

func TestFormData_decodeMapIntoStruct_PanicsWhenMapstructureNewDecoderFails(t *testing.T) {
	f := &FormData{}
	defer func() {
		p := recover()
		assert.NotNil(t, p)
		assert.Equal(t, errors.New("result must be a pointer"), p)
	}()
	f.decodeMapIntoStruct(map[string]interface{}{})
}

func TestFormData_RegisterTranslation_PanicsOnError(t *testing.T) {
	f := NewFormData(&struct{}{})
	defer func() {
		p := recover()
		assert.NotNil(t, p)
		assert.IsType(t, (*ut.ErrConflictingTranslation)(nil), p)
	}()
	f.RegisterTranslation("", "")
	f.RegisterTranslation("", "")
}

func TestFormData_RegisterTranslation_SetsArgumentsForErrorMessages(t *testing.T) {
	f := NewFormData(&struct {
		ID int64 `validate:"test=value"`
	}{})
	f.RegisterValidation("test", func(validator.FieldLevel) bool {
		return false
	})
	f.RegisterTranslation("test", "failed for field %[2]s with parameter %[3]s (tag=%[1]s)")
	err := f.ParseMapData(map[string]interface{}{"id": int64(1)})
	assert.Equal(t, err, FieldErrors{
		"ID": []string{"failed for field ID with parameter value (tag=test)"},
	})
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
				Layout:     "2006-01-02T15:04:05Z07:00",
				Value:      "2019-05-30T14:00:00ZZ",
				LayoutElem: "",
				ValueElem:  "Z",
				Message:    ": extra text: Z",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hook := stringToDatabaseTimeUTCHookFunc(time.RFC3339)
			converted, err := mapstructure.DecodeHookExec(hook, tt.typeFrom, tt.typeTo, tt.data)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, converted)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}
