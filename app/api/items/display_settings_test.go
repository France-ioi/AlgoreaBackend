package items

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

func TestStringFromDisplaySettings(t *testing.T) {
	tests := []struct {
		name string
		ds   database.JSON
		key  string
		def  string
		want string
	}{
		{name: "nil map returns default", ds: nil, key: "any", def: "fallback", want: "fallback"},
		{name: "absent key returns default", ds: database.JSON{}, key: "missing", def: "fallback", want: "fallback"},
		{name: "string value returned as-is", ds: database.JSON{"k": "value"}, key: "k", def: "fallback", want: "value"},
		{name: "non-string value falls back", ds: database.JSON{"k": 42.0}, key: "k", def: "fallback", want: "fallback"},
		{name: "nil JSON value falls back", ds: database.JSON{"k": nil}, key: "k", def: "fallback", want: "fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stringFromDisplaySettings(tt.ds, tt.key, tt.def))
		})
	}
}

func TestBoolFromDisplaySettings(t *testing.T) {
	tests := []struct {
		name string
		ds   database.JSON
		key  string
		def  bool
		want bool
	}{
		{name: "nil map returns default true", ds: nil, key: "any", def: true, want: true},
		{name: "nil map returns default false", ds: nil, key: "any", def: false, want: false},
		{name: "absent key returns default", ds: database.JSON{}, key: "missing", def: true, want: true},
		{name: "bool true returned as-is", ds: database.JSON{"k": true}, key: "k", def: false, want: true},
		{name: "bool false returned as-is", ds: database.JSON{"k": false}, key: "k", def: true, want: false},
		// `encoding/json` decodes JSON numbers as float64; the helper has to treat
		// non-zero numbers as `true` for legacy rows that store booleans numerically.
		{name: "float64 non-zero treated as true", ds: database.JSON{"k": 1.0}, key: "k", def: false, want: true},
		{name: "float64 zero treated as false", ds: database.JSON{"k": 0.0}, key: "k", def: true, want: false},
		{name: "string value falls back", ds: database.JSON{"k": "true"}, key: "k", def: false, want: false},
		{name: "nil JSON value falls back to default", ds: database.JSON{"k": nil}, key: "k", def: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, boolFromDisplaySettings(tt.ds, tt.key, tt.def))
		})
	}
}

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
