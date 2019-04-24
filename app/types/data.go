package types

import (
	"encoding/json"
	"reflect"
)

type (
	// Int64 is an integer which can be set/not-set and null/not-null
	Data struct {
		Value interface{}
		Set   bool
		Null  bool
	}
)

// AllAttributeser is a generic interface for all set/null information about these custom types
type AllAttributeser interface {
	AllAttributes() (value interface{}, isNull bool, isSet bool)
}

// AllAttributes unwraps the wrapped value and its attributes
func (d Data) AllAttributes() (value interface{}, isNull, isSet bool) {
	return d.Value, d.Null, d.Set
}

// unmarshalJSON parse JSON data to the type fields
func unmarshalJSON(data []byte, isSet, isNull *bool, value interface{}, valueType reflect.Type) (err error) {
	*isSet = true
	*isNull = string(data) == jsonNull

	temp := reflect.New(valueType).Interface()
	err = json.Unmarshal(data, temp)
	if err == nil {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(temp).Elem())
	}
	return
}
