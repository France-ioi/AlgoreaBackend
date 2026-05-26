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
// A NULL DB value (which the driver passes as `src == nil`) is decoded as a nil
// map. Without this short-circuit the unconditional `src.([]byte)` assertion
// would panic on NULL — defensive against e.g. a manual `UPDATE ... SET col = NULL`
// on a column whose schema is non-NULL.
func (j *JSON) Scan(src interface{}) (err error) {
	if src == nil {
		*j = nil
		return nil
	}
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

// OrEmpty returns a non-nil JSON map: the underlying map when the receiver
// holds one, and an empty (allocated) JSON map otherwise (including when the
// receiver pointer itself is nil). Use it on the response-construction path
// for columns documented as "always a JSON object, never NULL" (e.g.
// `items.display_settings`): the DB schema enforces NOT NULL, but `Scan`
// defensively decodes a stray NULL into a nil map, which `encoding/json`
// would then emit as `null` — silently violating the documented contract.
// OrEmpty closes that gap at the call site without changing `Scan`'s
// semantics (which other, legitimately nullable JSON columns like
// `users.profile` rely on for `null` round-tripping).
//
// Pointer receiver keeps method-set parity with `Scan`/`Value` (the
// `recvcheck` linter rejects a mixed receiver style on the same type).
func (j *JSON) OrEmpty() JSON {
	if j == nil || *j == nil {
		return JSON{}
	}
	return *j
}

var (
	_ = sql.Scanner(&JSON{})
	_ = driver.Valuer(&JSON{})
)
