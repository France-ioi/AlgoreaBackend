package items

import (
	"reflect"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

// stringFromDisplaySettings reads a string-typed key from the `display_settings`
// map and falls back to `def` when the key is absent or holds a non-string value.
//
// This is used to resolve legacy GET fields (e.g. `children_layout`) that the
// backend kept exposing for wire-compat after their dedicated DB columns were
// dropped: the value now lives opaquely inside `items.display_settings`, with
// the convention that defaults are *omitted* from the JSON.
func stringFromDisplaySettings(ds database.JSON, key, def string) string {
	if v, ok := ds[key].(string); ok {
		return v
	}
	return def
}

// boolFromDisplaySettings reads a bool-typed key from the `display_settings`
// map and falls back to `def` when the key is absent or holds a non-bool value.
//
// We accept JSON booleans (the canonical encoding new clients write) AND JSON
// numbers (the encoding any legacy/manually-inserted row might still hold for a
// boolean key — `encoding/json` decodes those into `float64`). This keeps the
// helper resilient to off-spec values without making the storage convention any
// less strict on the write side.
func boolFromDisplaySettings(ds database.JSON, key string, def bool) bool {
	switch v := ds[key].(type) {
	case bool:
		return v
	case float64:
		return v != 0
	}
	return def
}

// constructDisplaySettingsValidator rejects `null` for the `display_settings`
// field. The DB column is `NOT NULL`, so writing nil through gorm would fail.
//
// The framework only surfaces a validation error for fields that the client
// explicitly sent (see `FormData.processValidatorErrors`), so omitted requests
// pass even though the underlying field is also nil in that case.
func constructDisplaySettingsValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		// `DisplaySettings` is `*database.JSON` (pointer to a map). The validator
		// library extracts pointer kinds without diving when the pointer is nil,
		// and dives into the map otherwise — handle both.
		switch field.Kind() {
		case reflect.Ptr, reflect.Map:
			return !field.IsNil()
		default:
			return false
		}
	}
}

// registerDisplaySettingsValidator wires `constructDisplaySettingsValidator` and
// its translation onto a FormData. Callers in itemCreate and itemUpdate share
// the exact same validator + message, so this lives next to the helpers to keep
// every `display_settings`-related concern in one file.
func registerDisplaySettingsValidator(formData *formdata.FormData) {
	formData.RegisterValidation("display_settings", constructDisplaySettingsValidator())
	formData.RegisterTranslation("display_settings", "display_settings should be a JSON object and cannot be null")
}
