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

func (t *abstract) MarshalString() (string, error) {
	privateKey := reflect.ValueOf(t.Payload).Elem().FieldByName("PrivateKey").Interface().(*rsa.PrivateKey)
	return string(Generate(payloads.ConvertIntoMap(t.Payload), privateKey)), nil
}

var _ json.Marshaler = (*abstract)(nil)

// UnmarshalStringer is the interface implemented by types
// that can unmarshal a string description of themselves.
// For example, a token's string description is `{ENCODED_TOKEN}`
// while a token's JSON description is `"{ENCODED_TOKEN}"`
type UnmarshalStringer interface {
	UnmarshalString(string) error
}

// MarshalStringer is the interface implemented by types
// that can marshal themselves into a string.
// For example, a token's string description is `{ENCODED_TOKEN}`
// while a token's JSON description is `"{ENCODED_TOKEN}"`
type MarshalStringer interface {
	MarshalString() (string, error)
}

// Signer is the interface implemented by types
// that can sign themselves returning a token in a string.
type Signer interface {
	Sign(*rsa.PrivateKey) (string, error)
}

var _ UnmarshalStringer = (*abstract)(nil)
var _ MarshalStringer = (*abstract)(nil)

func marshalJSON(payload interface{}) ([]byte, error) {
	return (&abstract{Payload: payload}).MarshalJSON()
}

func unmarshalJSON(data []byte, payload interface{}) error {
	return (&abstract{Payload: payload}).UnmarshalJSON(data)
}

func marshalString(payload interface{}) (string, error) {
	return (&abstract{Payload: payload}).MarshalString()
}

func unmarshalString(data string, payload interface{}) error {
	return (&abstract{Payload: payload}).UnmarshalString(data)
}
