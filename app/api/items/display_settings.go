package items

import (
	"reflect"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

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
// its translation onto a FormData. itemCreate and itemUpdate share the exact
// same validator + message, so this helper keeps the duplication out of both
// handlers.
func registerDisplaySettingsValidator(formData *formdata.FormData) {
	formData.RegisterValidation("display_settings", constructDisplaySettingsValidator())
	formData.RegisterTranslation("display_settings", "display_settings should be a JSON object and cannot be null")
}
