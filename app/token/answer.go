package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Answer represents an answer token
type Answer payloads.AnswerToken

// UnmarshalJSON unmarshals the answer token from JSON
func (tt *Answer) UnmarshalJSON(raw []byte) error {
	return unmarshalJSON(raw, (*payloads.AnswerToken)(tt))
}

// UnmarshalString unmarshals the task token from a string
func (tt *Answer) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.AnswerToken)(tt))
}

// MarshalJSON marshals the answer token into JSON
func (tt *Answer) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the answer token into a string
func (tt *Answer) MarshalString() (string, error) { return marshalString(tt) }

// Sign returns a signed token as a string
func (tt *Answer) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*Answer)(nil)
	_ json.Marshaler    = (*Answer)(nil)
	_ UnmarshalStringer = (*Answer)(nil)
)
