package token

import (
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

var (
	_ json.Unmarshaler  = (*Answer)(nil)
	_ json.Marshaler    = (*Answer)(nil)
	_ UnmarshalStringer = (*Answer)(nil)
)
