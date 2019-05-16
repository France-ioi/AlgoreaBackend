package payloads

import "encoding/json"

// Anything represents a value of any type serialized as JSON
type Anything []byte

// UnmarshalJSON of Anything just copies the JSON data
func (a *Anything) UnmarshalJSON(raw []byte) error {
	*a = Anything(raw)
	return nil
}

// MarshalJSON of Anything copies the stored JSON data back
func (a Anything) MarshalJSON() ([]byte, error) {
	return []byte(a), nil
}

var _ = json.Unmarshaler(func(a Anything) *Anything { return &a }([]byte{}))
var _ = json.Marshaler(Anything([]byte{}))
