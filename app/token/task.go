package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Task represents a task token.
type Task payloads.TaskToken

// UnmarshalJSON unmarshals the task token from JSON.
func (tt *Task) UnmarshalJSON(raw []byte) error { return unmarshalJSON(raw, (*payloads.TaskToken)(tt)) }

// UnmarshalString unmarshals the task token from a string.
func (tt *Task) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.TaskToken)(tt))
}

// MarshalJSON marshals the task token into JSON.
func (tt *Task) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the task token into a string.
func (tt *Task) MarshalString() (string, error) { return marshalString(tt) }

// Sign returns a signed token as a string.
func (tt *Task) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*Task)(nil)
	_ json.Marshaler    = (*Task)(nil)
	_ UnmarshalStringer = (*Task)(nil)
	_ MarshalStringer   = (*Task)(nil)
	_ Signer            = (*Task)(nil)
)
