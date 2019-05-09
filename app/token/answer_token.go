package token

import (
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// AnswerToken represents an answer token
type AnswerToken payloads.AnswerToken

// UnmarshalJSON unmarshals the answer token from JSON
func (tt *AnswerToken) UnmarshalJSON(raw []byte) error {
	return (&abstract{tt}).UnmarshalJSON(raw)
}

var _ json.Unmarshaler = (*AnswerToken)(nil)

// MarshalJSON marshals the answer token into JSON
func (tt *AnswerToken) MarshalJSON() ([]byte, error) {
	return (&abstract{tt}).MarshalJSON()
}

var _ json.Marshaler = (*AnswerToken)(nil)
