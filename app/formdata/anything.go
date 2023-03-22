package formdata

import (
	"encoding/json"
)

// Anything represents a value of any type serialized as JSON.
type Anything struct {
	raw []byte
}

// AnythingFromString creates an instance of Anything with data from the given string.
func AnythingFromString(s string) *Anything {
	return &Anything{raw: []byte(s)}
}

// AnythingFromBytes creates an instance of Anything with data from the given bytes slice.
func AnythingFromBytes(bytes []byte) *Anything {
	return &Anything{raw: bytes}
}

// Bytes returns stored bytes.
func (a Anything) Bytes() []byte {
	return a.raw
}

// UnmarshalJSON of Anything just copies the JSON data.
func (a *Anything) UnmarshalJSON(raw []byte) error {
	a.raw = make([]byte, len(raw))
	copy(a.raw, raw)
	return nil
}

// MarshalJSON of Anything copies the stored JSON data back.
func (a Anything) MarshalJSON() ([]byte, error) {
	if len(a.raw) == 0 {
		return []byte("null"), nil
	}
	return a.raw, nil
}

var _ = json.Unmarshaler(&Anything{})
var _ = json.Marshaler(Anything{})
