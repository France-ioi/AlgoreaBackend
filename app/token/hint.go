package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Hint represents a hint token.
type Hint payloads.HintToken

// UnmarshalString unmarshals the task token from a string.
func (tt *Hint) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.HintToken)(tt))
}

// UnmarshalJSON unmarshals the answer token from JSON.
func (tt *Hint) UnmarshalJSON(raw []byte) error { return unmarshalJSON(raw, (*payloads.HintToken)(tt)) }

// MarshalJSON marshals the answer token into JSON.
func (tt *Hint) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the hint token into a string.
func (tt *Hint) MarshalString() (string, error) { return marshalString(tt) }

// Sign returns a signed token as a string.
func (tt *Hint) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*Hint)(nil)
	_ json.Marshaler    = (*Hint)(nil)
	_ UnmarshalStringer = (*Hint)(nil)
	_ MarshalStringer   = (*Hint)(nil)
	_ Signer            = (*Hint)(nil)
)
