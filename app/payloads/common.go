// Package payloads defines data structures to be used as tokens.
package payloads

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

// Binder is an interface for managing payloads.
type Binder interface {
	Bind() error
}

// ParseMap converts a map into a structure and validates fields.
func ParseMap(raw map[string]interface{}, target interface{}) error {
	if err := formdata.NewFormData(target).ParseMapData(raw); err != nil {
		typeName := reflect.TypeOf(target).Elem().Name()
		return fmt.Errorf("invalid %s: %s", typeName, err.Error())
	}

	if binder, ok := target.(Binder); ok {
		return binder.Bind()
	}

	return nil
}

// ConverterIntoMap is an interface implemented by types
// that can convert themselves into a map.
// payloads.ConvertIntoMap uses this interface to convert structs into maps.
type ConverterIntoMap interface {
	ConvertIntoMap() map[string]interface{}
}

// ConvertIntoMap converts a struct into a map
// Fields without a `json` tag or having '-' as a json field name are skipped.
func ConvertIntoMap(source interface{}) map[string]interface{} {
	if converter, ok := source.(ConverterIntoMap); ok {
		return converter.ConvertIntoMap()
	}

	sourceValue := reflect.ValueOf(source)
	for sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	sourceType := sourceValue.Type()
	fieldsCount := sourceValue.NumField()
	out := make(map[string]interface{}, fieldsCount)
	for fieldNumber := 0; fieldNumber < fieldsCount; fieldNumber++ {
		field := sourceType.Field(fieldNumber)
		jsonName, omitEmpty := getJSONFieldNameAndOmitEmpty(&field)
		if jsonName == "-" {
			continue
		}
		fieldValue := sourceValue.Field(fieldNumber)
		if !fieldValue.CanInterface() { // skip unexported fields
			continue
		}
		fieldValue = resolvePointer(fieldValue)
		if !omitEmpty || fieldValue.Type().Kind() != reflect.Ptr || !fieldValue.IsNil() {
			if shouldConvert(fieldValue) {
				out[jsonName] = ConvertIntoMap(fieldValue.Addr().Interface())
			} else {
				out[jsonName] = fieldValue.Interface()
			}
		}
	}
	return out
}

func shouldConvert(fieldValue reflect.Value) bool {
	return fieldValue.Kind() == reflect.Struct &&
		(fieldValue.Type().Name() != "Anything" ||
			fieldValue.Type().PkgPath() != "github.com/France-ioi/AlgoreaBackend/v2/app/formdata")
}

func resolvePointer(fieldValue reflect.Value) reflect.Value {
	for fieldValue.IsValid() && fieldValue.Type().Kind() == reflect.Ptr && !fieldValue.IsNil() {
		fieldValue = fieldValue.Elem()
	}
	return fieldValue
}

func getJSONFieldNameAndOmitEmpty(structField *reflect.StructField) (string, bool) {
	jsonTagParts := strings.Split(structField.Tag.Get("json"), ",")
	name := jsonTagParts[0]
	if name == "" {
		name = "-"
	}
	var omitEmpty bool
	for i := 1; i < len(jsonTagParts); i++ {
		if jsonTagParts[i] == "omitempty" {
			omitEmpty = true
			break
		}
	}
	return name, omitEmpty
}
