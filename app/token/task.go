package token

import (
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Task represents a task token
type Task payloads.TaskToken

// UnmarshalJSON unmarshals the task token from JSON
func (tt *Task) UnmarshalJSON(raw []byte) error { return unmarshalJSON(raw, (*payloads.TaskToken)(tt)) }

// UnmarshalString unmarshals the task token from a string
func (tt *Task) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.TaskToken)(tt))
}

// MarshalJSON marshals the task token into JSON
func (tt *Task) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

var _ json.Unmarshaler = (*Task)(nil)
var _ json.Marshaler = (*Task)(nil)
var _ UnmarshalStringer = (*Task)(nil)
