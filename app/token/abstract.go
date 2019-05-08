package token

import (
	"encoding/json"
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

type abstract struct {
	Payload interface{}
}

func (t *abstract) UnmarshalJSON(raw []byte) error {
	var err error

	var buffer string
	err = json.Unmarshal(raw, &buffer)
	if err != nil {
		return err
	}

	tokenPayload, err := ParseAndValidate([]byte(buffer))
	if err != nil {
		return err
	}

	return payloads.ParseMap(tokenPayload, t.Payload)
}

var _ json.Unmarshaler = (*abstract)(nil)

func (t *abstract) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", Generate(payloads.ConvertIntoMap(t.Payload)))), nil
}

var _ json.Marshaler = (*abstract)(nil)
