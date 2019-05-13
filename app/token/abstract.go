package token

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"reflect"

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

	return t.UnmarshalString(buffer)
}

func (t *abstract) UnmarshalString(raw string) error {
	var err error

	publicKey := reflect.ValueOf(t.Payload).Elem().FieldByName("PublicKey").Interface().(*rsa.PublicKey)
	tokenPayload, err := ParseAndValidate([]byte(raw), publicKey)
	if err != nil {
		return err
	}

	return payloads.ParseMap(tokenPayload, t.Payload)
}

var _ json.Unmarshaler = (*abstract)(nil)

func (t *abstract) MarshalJSON() ([]byte, error) {
	privateKey := reflect.ValueOf(t.Payload).Elem().FieldByName("PrivateKey").Interface().(*rsa.PrivateKey)
	return []byte(fmt.Sprintf("%q", Generate(payloads.ConvertIntoMap(t.Payload), privateKey))), nil
}

var _ json.Marshaler = (*abstract)(nil)
