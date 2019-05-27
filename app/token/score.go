package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Score represents a score token
type Score payloads.ScoreToken

// UnmarshalJSON unmarshals the score token from JSON
func (tt *Score) UnmarshalJSON(raw []byte) error {
	return unmarshalJSON(raw, (*payloads.ScoreToken)(tt))
}

// MarshalJSON marshals the score token into JSON
func (tt *Score) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the score token into a string
func (tt *Score) MarshalString() (string, error) { return marshalString(tt) }

// UnmarshalString unmarshals the score token from a string
func (tt *Score) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.ScoreToken)(tt))
}

// Sign returns a signed score token as a string
func (tt *Score) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*Score)(nil)
	_ json.Marshaler    = (*Score)(nil)
	_ UnmarshalStringer = (*Score)(nil)
	_ MarshalStringer   = (*Score)(nil)
	_ Signer            = (*Score)(nil)
)
