package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// JSON represents a DB value storing a valid JSON object.
// As JSON implements sql.Scaner interface, it is automatically parsed into map[string]interface{} on data load.
// As JSON implements driver.Valuer interface, it is automatically converted to JSON on data save.
//
// swagger:type object
type JSON map[string]interface{}

// Scan assigns a value from a database driver value (unmarshalls given bytes).
func (j *JSON) Scan(src interface{}) (err error) {
	//nolint:forcetypeassert // panic if src is not []byte (this means that the database returned a value of unexpected type)
	return json.Unmarshal(src.([]byte), j)
}

// Value returns a database driver Value of JSON (marshaled JSON bytes).
func (j *JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil //nolint:nilnil // return nil for nil receiver
	}
	return json.Marshal(*j)
}

var (
	_ = sql.Scanner(&JSON{})
	_ = driver.Valuer(&JSON{})
)
