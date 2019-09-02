package formdata

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/France-ioi/validator"
)

func validateDuration(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}
	hms := strings.Split(field.String(), ":")
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

var dmyDateRegexp = regexp.MustCompile(`^[0-3][0-9]-[0-1][0-9]-[\d]{4}$`)

func validateDMYDate(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}
	return dmyDateRegexp.MatchString(field.String())
}
