package formdata

import (
	"errors"
	"testing"

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
