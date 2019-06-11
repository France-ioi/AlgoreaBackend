package formdata

import (
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
