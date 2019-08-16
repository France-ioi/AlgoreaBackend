package formdata

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

func validateDuration(fl validator.FieldLevel) bool {
	hms := strings.Split(fl.Field().Interface().(string), ":")
	if len(hms) != 3 {
		return false
	}

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
	return dmyDateRegexp.MatchString(fl.Field().Interface().(string))
}
