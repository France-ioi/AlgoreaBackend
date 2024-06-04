package token

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

// ProfileEdit represents a profile edit token.
type ProfileEdit payloads.ProfileEditToken

// UnmarshalJSON unmarshals the answer token from JSON.
func (tt *ProfileEdit) UnmarshalJSON(raw []byte) error {
	return unmarshalJSON(raw, (*payloads.ProfileEditToken)(tt))
}

// UnmarshalString unmarshals the task token from a string.
func (tt *ProfileEdit) UnmarshalString(raw string) error {
	return unmarshalString(raw, (*payloads.ProfileEditToken)(tt))
}

// MarshalJSON marshals the answer token into JSON.
func (tt *ProfileEdit) MarshalJSON() ([]byte, error) { return marshalJSON(tt) }

// MarshalString marshals the answer token into a string.
func (tt *ProfileEdit) MarshalString() (string, error) { return marshalString(tt) }

// Sign returns a signed token as a string.
func (tt *ProfileEdit) Sign(privateKey *rsa.PrivateKey) (string, error) {
	tt.PrivateKey = privateKey
	return tt.MarshalString()
}

var (
	_ json.Unmarshaler  = (*ProfileEdit)(nil)
	_ json.Marshaler    = (*ProfileEdit)(nil)
	_ UnmarshalStringer = (*ProfileEdit)(nil)
	_ MarshalStringer   = (*ProfileEdit)(nil)
	_ Signer            = (*ProfileEdit)(nil)
)
