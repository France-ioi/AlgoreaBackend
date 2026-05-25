package items

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

// displaySettingsTestForm is a minimal struct that mirrors the field type used
// by the production Item struct so we can exercise the validator end-to-end
// through the actual FormData parse path (the path that fires on every
// itemCreate/itemUpdate request).
type displaySettingsTestForm struct {
	DisplaySettings *database.JSON `json:"display_settings" validate:"display_settings"`
}

func TestRegisterDisplaySettingsValidator(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		wantErrs formdata.FieldErrorsError
	}{
		{
			name:  "field omitted passes (validator only runs on set fields with non-nil pointer)",
			input: map[string]interface{}{},
		},
		{
			name:  "empty object passes",
			input: map[string]interface{}{"display_settings": map[string]interface{}{}},
		},
		{
			name:  "populated object passes",
			input: map[string]interface{}{"display_settings": map[string]interface{}{"children_layout": "Grid"}},
		},
		{
			name:     "explicit null is rejected",
			input:    map[string]interface{}{"display_settings": nil},
			wantErrs: formdata.FieldErrorsError{"display_settings": []string{"display_settings should be a JSON object and cannot be null"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &displaySettingsTestForm{}
			fd := formdata.NewFormData(form)
			registerDisplaySettingsValidator(fd)

			err := fd.ParseMapData(tt.input)
			if tt.wantErrs == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.wantErrs, err)
			}
		})
	}
}

// nonPtrNonMapDisplaySettingsForm uses the `display_settings` validator on a
// field whose kind is neither pointer nor map, to cover the validator's
// `default: return false` fallback. The production struct never exposes the
// validator on such a field, but the fallback exists as a defensive guard and
// must remain reachable behavior.
type nonPtrNonMapDisplaySettingsForm struct {
	DisplaySettings string `json:"display_settings" validate:"display_settings"`
}

func TestConstructDisplaySettingsValidator_RejectsNonPtrNonMapField(t *testing.T) {
	form := &nonPtrNonMapDisplaySettingsForm{}
	fd := formdata.NewFormData(form)
	registerDisplaySettingsValidator(fd)

	err := fd.ParseMapData(map[string]interface{}{"display_settings": "not-a-map"})
	assert.Equal(t, formdata.FieldErrorsError{
		"display_settings": []string{"display_settings should be a JSON object and cannot be null"},
	}, err)
}
