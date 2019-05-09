package testhelpers

import (
	"fmt"
	"reflect"
)

var knownTypes = map[string]reflect.Type{}

func getZeroStructPtr(typeName string) (interface{}, error) {
	if _, ok := knownTypes[typeName]; !ok {
		return nil, fmt.Errorf("unknown type: %q", typeName)
	}
	return reflect.New(knownTypes[typeName]).Interface(), nil
}
