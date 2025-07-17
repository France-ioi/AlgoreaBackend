package token

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
)

// Token represents a token. It contains a payload of type P,
// a public key for validation, and a private key for signing.
// The idea is to use this struct to represent various types of tokens
// to validate and sign them easily.
type Token[P any] struct {
	Payload    P
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// UnmarshalJSON unmarshals the token from JSON.
// It expects the JSON to be a JSON string containing the encoded token.
// The token is parsed and validated using the public key.
// If the token is valid, it populates the Payload field with the parsed data.
func (t *Token[P]) UnmarshalJSON(raw []byte) error {
	var err error

	var buffer string
	err = json.Unmarshal(raw, &buffer)
	if err != nil {
		return err
	}

	return t.UnmarshalString(buffer)
}

// UnmarshalString unmarshals the token from a string.
// It expects the string to be an encoded token.
// The token is parsed and validated using the public key.
// If the token is valid, it populates the Payload field with the parsed data.
func (t *Token[P]) UnmarshalString(raw string) error {
	var err error

	tokenPayload, err := ParseAndValidate([]byte(raw), t.PublicKey)
	if err != nil {
		return err
	}

	return payloads.ParseMap(tokenPayload, &t.Payload)
}

var _ json.Unmarshaler = (*Token[interface{}])(nil)

// MarshalJSON marshals the token into JSON.
// It generates a signed token from the Payload field using the private key,
// and returns it as a JSON string.
func (t *Token[P]) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", Generate(payloads.ConvertIntoMap(t.Payload), t.PrivateKey))), nil
}

// MarshalString marshals the token into a string.
// It generates a signed token from the Payload field using the private key,
// and returns it as a string.
func (t *Token[P]) MarshalString() (string, error) {
	return string(Generate(payloads.ConvertIntoMap(t.Payload), t.PrivateKey)), nil
}

var _ json.Marshaler = (*Token[interface{}])(nil)

// UnmarshalStringer is the interface implemented by types
// that can unmarshal a string description of themselves.
// For example, a token's string description is `{ENCODED_TOKEN}`
// while a token's JSON description is `"{ENCODED_TOKEN}"`.
type UnmarshalStringer interface {
	UnmarshalString(s string) error
}

// MarshalStringer is the interface implemented by types
// that can marshal themselves into a string.
// For example, a token's string description is `{ENCODED_TOKEN}`
// while a token's JSON description is `"{ENCODED_TOKEN}"`.
type MarshalStringer interface {
	MarshalString() (string, error)
}

// Signer is the interface implemented by types
// that can sign themselves returning a token in a string.
type Signer interface {
	Sign(privateKey *rsa.PrivateKey) (string, error)
}

// Sign returns a signed token as a string.
func (t *Token[P]) Sign(privateKey *rsa.PrivateKey) (string, error) {
	t.PrivateKey = privateKey
	return t.MarshalString()
}

var _ Signer = (*Token[interface{}])(nil)

var (
	_ UnmarshalStringer = (*Token[interface{}])(nil)
	_ MarshalStringer   = (*Token[interface{}])(nil)
)

// ConvertIntoMap converts the token's payload into a map.
func (t *Token[P]) ConvertIntoMap() map[string]interface{} {
	return payloads.ConvertIntoMap(t.Payload)
}

var _ payloads.ConverterIntoMap = (*Token[interface{}])(nil)
