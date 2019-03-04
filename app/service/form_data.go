package service

import (
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/France-ioi/govalidator"
	"github.com/France-ioi/mapstructure"
)

// FormData can parse JSON, validate it and construct a map for updating DB
type FormData struct {
	definitionStructure interface{}
	fieldErrors         FieldErrors
	metadata            mapstructure.Metadata
	usedKeys            map[string]bool
}

// NewFormData creates a new FormData object for given definitions
func NewFormData(definitionStructure interface{}) *FormData {
	return &FormData{
		definitionStructure: definitionStructure,
	}
}

// ParseJSONRequestData parses and validates JSON according to the structure definition
func (f *FormData) ParseJSONRequestData(r *http.Request) error {
	if err := f.decodeRequestJSONDataIntoStruct(r); err != nil {
		return err
	}

	f.checkProvidedFields()
	f.validateFieldValues()

	if len(f.fieldErrors) > 0 {
		return f.fieldErrors
	}

	return nil
}

// ConstructMapForDB constructs a map for updating DB. It uses both the definition and the JSON input
func (f *FormData) ConstructMapForDB() map[string]interface{} {
	result := map[string]interface{}{}
	f.addDBFieldsIntoMap(result, reflect.ValueOf(f.definitionStructure), "")
	return result
}

var mapstructTypeErrorRegexp = regexp.MustCompile(`^'([^']*)'\s+(.*)$`)
var mapstructDecodingErrorRegexp = regexp.MustCompile(`^error decoding '([^']*)':\s+(.*)$`)

func (f *FormData) decodeRequestJSONDataIntoStruct(r *http.Request) error {
	var rawData map[string]interface{}
	var err error
	if err = json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		return err
	}

	f.fieldErrors = make(FieldErrors)
	f.usedKeys = map[string]bool{}

	var decoder *mapstructure.Decoder
	decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           f.definitionStructure,
		DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339),
		ErrorUnused:      false, // we will check this on our own
		Metadata:         &f.metadata,
		TagName:          "json",
		ZeroFields:       true, // this marks keys with null values as used
		WeaklyTypedInput: false,
	})

	if err != nil {
		panic(err) // this error can only be caused by bugs in the code
	}

	if err = decoder.Decode(rawData); err != nil {
		mapstructureErr := err.(*mapstructure.Error)
		for _, fieldErrorString := range mapstructureErr.Errors { // Convert mapstructure's errors to our format
			if matches := mapstructTypeErrorRegexp.FindStringSubmatch(fieldErrorString); len(matches) > 0 {
				key := make([]byte, len(matches[1]))
				copy(key, matches[1])
				value := make([]byte, len(matches[2]))
				copy(value, matches[2])
				f.fieldErrors[string(key)] = append(f.fieldErrors[string(key)], string(value))
			} else if matches := mapstructDecodingErrorRegexp.FindStringSubmatch(fieldErrorString); len(matches) > 0 {
				key := make([]byte, len(matches[1]))
				copy(key, matches[1])
				f.fieldErrors[string(key)] = append(f.fieldErrors[string(key)], "decoding error: "+matches[2])
			} else {
				f.fieldErrors[""] = append(f.fieldErrors[""], fieldErrorString) // should never happen
			}
		}
	}
	return nil
}

func (f *FormData) validateFieldValues() {
	if _, err := govalidator.ValidateStruct(f.definitionStructure); err != nil {
		f.processGovalidatorErrors(err)
	}
}

func (f *FormData) processGovalidatorErrors(err error) {
	validatorErrors := err.(govalidator.Errors)
	for _, validatorError := range validatorErrors {
		if err, ok := validatorError.(govalidator.Error); ok {
			currentFieldType := reflect.TypeOf(f.definitionStructure)
			for pathIndex, pathElement := range err.Path {
				for currentFieldType.Kind() == reflect.Ptr {
					currentFieldType = currentFieldType.Elem()
				}
				field, _ := currentFieldType.FieldByName(pathElement)
				currentFieldType = field.Type
				jsonName := getJSONFieldName(field)
				if jsonName != "-" {
					err.Path[pathIndex] = jsonName
				}
			}

			path := strings.Join(err.Path, ".")
			if len(path) > 0 {
				path += "."
			}
			path += err.Name
			if f.usedKeys[path] {
				f.fieldErrors[path] = append(f.fieldErrors[path], err.Err.Error())
			}
		} else {
			f.processGovalidatorErrors(validatorError)
		}
	}
}

func (f *FormData) checkProvidedFields() {
	for _, unusedKey := range f.metadata.Unused {
		f.fieldErrors[unusedKey] = append(f.fieldErrors[unusedKey], "unexpected field")
	}
	for _, usedKey := range f.metadata.Keys {
		f.usedKeys[usedKey] = true
	}
}

func (f *FormData) addDBFieldsIntoMap(resultMap map[string]interface{}, reflValue reflect.Value, prefix string) {
	traverseStructure(func(field reflect.Value, structField reflect.StructField, jsonName string) bool {
		if _, ok := f.usedKeys[jsonName]; !ok {
			return false
		}

		dbName := structField.Name

		for _, str := range []string{structField.Tag.Get("sql"), structField.Tag.Get("gorm")} {
			tags := strings.Split(str, ";")
			for _, value := range tags {
				v := strings.Split(value, ":")
				key := strings.TrimSpace(v[0])
				if key == "-" {
					return false // skip this field
				}
				var value string
				if len(v) >= 2 {
					value = strings.Join(v[1:], ":")
				} else {
					value = key
				}
				if key == "column" {
					dbName = value
				}
			}
		}

		// For now, all the fields from nested structures will be set in the root map.
		// Yet the nested structures themselves will not be in the map
		if field.Kind() != reflect.Struct {
			resultMap[dbName] = field.Interface()
		}

		return true
	}, reflValue, "")
}

func traverseStructure(fn func(fieldValue reflect.Value, structField reflect.StructField, jsonName string) bool,
	reflValue reflect.Value, prefix string) {
	for reflValue.Kind() == reflect.Ptr {
		reflValue = reflValue.Elem()
	}
	numberOfFields := reflValue.NumField()

	for i := 0; i < numberOfFields; i++ {
		field := reflValue.Field(i)
		structField := reflValue.Type().Field(i)
		firstRune, _ := utf8.DecodeRuneInString(structField.Name)
		if !unicode.IsUpper(firstRune) { // skip unexported fields
			continue
		}

		jsonName := getJSONFieldName(structField)
		if jsonName == "-" { // skip fields ignored in json
			continue
		}
		if len(prefix) > 0 {
			jsonName = prefix + "." + jsonName
		}

		result := fn(field, structField, jsonName)

		if result && field.Kind() == reflect.Struct {
			traverseStructure(fn, field, jsonName)
		}
	}
}

func getJSONFieldName(structField reflect.StructField) string {
	jsonTagParts := strings.Split(structField.Tag.Get("json"), ",")
	if len(jsonTagParts[0]) == 0 {
		return "-"
	}
	return jsonTagParts[0]
}
