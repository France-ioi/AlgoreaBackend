package token

import (
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Answer represents an answer token
type Answer payloads.AnswerToken

// UnmarshalJSON unmarshals the answer token from JSON
func (tt *Answer) UnmarshalJSON(raw []byte) error {
	return (&abstract{tt}).UnmarshalJSON(raw)
}

var _ json.Unmarshaler = (*Answer)(nil)

// MarshalJSON marshals the answer token into JSON
func (tt *Answer) MarshalJSON() ([]byte, error) {
	return (&abstract{tt}).MarshalJSON()
}

var _ json.Marshaler = (*Answer)(nil)
