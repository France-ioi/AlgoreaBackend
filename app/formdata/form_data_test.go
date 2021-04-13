package formdata

import (
	"errors"
	"testing"

	"github.com/France-ioi/validator"
	ut "github.com/go-playground/universal-translator"
	"github.com/stretchr/testify/assert"
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
