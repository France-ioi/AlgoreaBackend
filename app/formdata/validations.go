package formdata

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/France-ioi/validator"
)

func validateNull(fl validator.FieldLevel) bool {
	field := fl.Field()
	if !field.IsValid() || field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}
	return false
}

func validateDuration(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}
	//nolint:mnd // we want exactly 3 parts (hours:minutes:seconds), but allow the 4th part in case of error
	hms := strings.SplitN(field.String(), ":", 4)
	//nolint:mnd // fail when the number of parts is not 3 (hours:minutes:seconds)
	if len(hms) != 3 {
		return false
	}

	return validateHMSForDuration(hms)
}

func validateHMSForDuration(hms []string) bool {
	hours, err := strconv.Atoi(hms[0])
	if err != nil {
		return false
	}
	minutes, err := strconv.Atoi(hms[1])
	if err != nil {
		return false
	}
	seconds, err := strconv.Atoi(hms[2])
	if err != nil {
		return false
	}
	return hours >= 0 && hours <= 838 && minutes >= 0 && minutes <= 59 && seconds >= 0 && seconds <= 59
}

var dmyDateRegexp = regexp.MustCompile(`^[0-3]\d-[01]\d-\d{4}$`)

func validateDMYDate(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}
	return dmyDateRegexp.MatchString(field.String())
}
