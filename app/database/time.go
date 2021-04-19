package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Time is the same as time.Time, but it can be assigned from MySQL datetime string representation (implements Scanner interface),
// can convert itself to sql/driver.Value (implements Valuer interface) and marshal itself as JSON (implements json.Marshaler)
//
// swagger:strfmt date-time
type Time time.Time

// Scan assigns a value from a database driver value ([]byte)
func (t *Time) Scan(src interface{}) (err error) {
	return t.ScanString(string(src.([]byte)))
}

const timeFormat = "2006-01-02 15:04:05.999999"

// ScanString assigns a value from string with a database driver value
func (t *Time) ScanString(str string) (err error) {
	// Based on go-sql-driver/mysql.parseDateTime (see https://github.com/go-sql-driver/mysql/blob/master/utils.go#L109)
	base := "0000-00-00 00:00:00.0000000"
	switch len(str) {
	case 10, 19, 21, 22, 23, 24, 25, 26: // up to "YYYY-MM-DD HH:MM:SS.MMMMMM"
		if str == base[:len(str)] {
			return nil
		}
		var parsedTime time.Time
		parsedTime, err = time.Parse(timeFormat[:len(str)], str)
		if err == nil {
			*t = Time(parsedTime)
		}
		return
	default:
		err = fmt.Errorf("invalid time string: %s", str)
		return
	}
}

// Value returns a database driver Value (*time.Time)
func (t *Time) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return (*time.Time)(t).UTC().Format(timeFormat), nil
}

// MarshalJSON returns the JSON encoding of t
func (t *Time) MarshalJSON() ([]byte, error) {
	return (*time.Time)(t).MarshalJSON()
}

var (
	_ = sql.Scanner(&Time{})
	_ = driver.Valuer(&Time{})
	_ = json.Marshaler(&Time{})
)
