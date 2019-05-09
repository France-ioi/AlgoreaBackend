package token

import (
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// Task represents a task token
type Task payloads.TaskToken

// UnmarshalJSON unmarshals the task token from JSON
func (tt *Task) UnmarshalJSON(raw []byte) error {
	return (&abstract{tt}).UnmarshalJSON(raw)
}

var _ json.Unmarshaler = (*Task)(nil)

// MarshalJSON marshals the task token into JSON
func (tt *Task) MarshalJSON() ([]byte, error) {
	return (&abstract{tt}).MarshalJSON()
}

var _ json.Marshaler = (*Task)(nil)
