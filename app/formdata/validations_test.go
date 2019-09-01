package formdata

import (
	"reflect"
	"testing"

	validator "gopkg.in/go-playground/validator.v9"
)

func Test_validateDuration(t *testing.T) {
	tests := []struct {
		name string
		fl   validator.FieldLevel
		want bool
	}{
		{name: "wrong format", fl: &FieldLevel{FieldValue: "12:34"}, want: false},
		{name: "invalid hours", fl: &FieldLevel{FieldValue: "ab:59:59"}, want: false},
		{name: "negative hours", fl: &FieldLevel{FieldValue: "-1:59:59"}, want: false},
		{name: "too many hours", fl: &FieldLevel{FieldValue: "839:59:59"}, want: false},
		{name: "invalid minutes", fl: &FieldLevel{FieldValue: "99:ab:59"}, want: false},
		{name: "negative minutes", fl: &FieldLevel{FieldValue: "99:-1:59"}, want: false},
		{name: "too many minutes", fl: &FieldLevel{FieldValue: "99:60:59"}, want: false},
		{name: "invalid seconds", fl: &FieldLevel{FieldValue: "99:59:ab"}, want: false},
		{name: "negative seconds", fl: &FieldLevel{FieldValue: "99:59:-1"}, want: false},
		{name: "too many seconds", fl: &FieldLevel{FieldValue: "99:59:60"}, want: false},
		{name: "max possible value", fl: &FieldLevel{FieldValue: "838:59:59"}, want: true},
		{name: "min possible value", fl: &FieldLevel{FieldValue: "00:00:00"}, want: true},
		{name: "short notation", fl: &FieldLevel{FieldValue: "0:0:0"}, want: true},
		{name: "nil", fl: &FieldLevel{FieldValue: (*string)(nil)}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := validateDuration(tt.fl); got != tt.want {
				t.Errorf("validateDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateDMYDate(t *testing.T) {
	tests := []struct {
		name string
		fl   validator.FieldLevel
		want bool
	}{
		{name: "wrong format", fl: &FieldLevel{FieldValue: "12-34"}, want: false},
		{name: "zero day", fl: &FieldLevel{FieldValue: "00-01-2019"}, want: true},
		{name: "day is too big", fl: &FieldLevel{FieldValue: "32-01-2019"}, want: true},
		{name: "zero month", fl: &FieldLevel{FieldValue: "01-00-2019"}, want: true},
		{name: "month is too big", fl: &FieldLevel{FieldValue: "01-13-2019"}, want: true},
		{name: "zero year", fl: &FieldLevel{FieldValue: "01-01-0000"}, want: true},
		{name: "max possible value", fl: &FieldLevel{FieldValue: "01-11-2019"}, want: true},
		{name: "min possible value", fl: &FieldLevel{FieldValue: "01-01-0000"}, want: true},
		{name: "with letter", fl: &FieldLevel{FieldValue: "a1-01-0000"}, want: false},
		{name: "nil", fl: &FieldLevel{FieldValue: (*string)(nil)}, want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := validateDMYDate(tt.fl); got != tt.want {
				t.Errorf("validateDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// FieldLevel contains all the information and helper functions
// to validate a field
type FieldLevel struct {
	FieldValue interface{}
}

func (fl *FieldLevel) Top() reflect.Value      { return reflect.ValueOf(nil) }
func (fl *FieldLevel) Parent() reflect.Value   { return reflect.ValueOf(nil) }
func (fl *FieldLevel) Field() reflect.Value    { return reflect.ValueOf(fl.FieldValue) }
func (fl *FieldLevel) FieldName() string       { return "" }
func (fl *FieldLevel) StructFieldName() string { return "" }
func (fl *FieldLevel) Path() string            { return "" }
func (fl *FieldLevel) StructPath() string      { return "" }
func (fl *FieldLevel) Param() string           { return "" }
func (fl *FieldLevel) ExtractType(reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return reflect.ValueOf(nil), reflect.Ptr, true
}
func (fl *FieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.ValueOf(nil), reflect.Ptr, true
}

var _ validator.FieldLevel = &FieldLevel{}
