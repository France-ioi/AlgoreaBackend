package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Thread represents a thread token.
type Thread payloads.ThreadToken

// UnmarshalJSON unmarshals the answer token from JSON.
func (tt *Thread) UnmarshalJSON(raw []byte) error {
	return unmarshalJSON(raw, (*payloads.ThreadToken)(tt))
}

// UnmarshalString unmarshals the task token from a string.
func (tt *Thread) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.ThreadToken)(tt))
}

// MarshalJSON marshals the answer token into JSON.
func (tt *Thread) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the answer token into a string.
func (tt *Thread) MarshalString() (string, error) { return marshalString(tt) }

// Sign returns a signed token as a string.
func (tt *Thread) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*Thread)(nil)
	_ json.Marshaler    = (*Thread)(nil)
	_ UnmarshalStringer = (*Thread)(nil)
	_ MarshalStringer   = (*Task)(nil)
	_ Signer            = (*Task)(nil)
)
