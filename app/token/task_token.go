package token

import (
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// TaskToken represents a task token
type TaskToken payloads.TaskTokenPayload

// UnmarshalJSON unmarshals the task token from JSON
func (tt *TaskToken) UnmarshalJSON(raw []byte) error {
	return (&abstract{tt}).UnmarshalJSON(raw)
}

var _ json.Unmarshaler = (*TaskToken)(nil)

// MarshalJSON marshals the task token into JSON
func (tt *TaskToken) MarshalJSON() ([]byte, error) {
	return (&abstract{tt}).MarshalJSON()
}

var _ json.Marshaler = (*TaskToken)(nil)
