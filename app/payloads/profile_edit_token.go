package payloads

import "crypto/rsa"

// ProfileEditToken permits a requester user to edit the profile of a target user.
// swagger:model ProfileEditToken
type ProfileEditToken struct {
	// Format dd-mm-yyyy
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// User who requested the token.
	// required:true
	RequesterID string `json:"requester_id"`
	// User whose profile is to be edited.
	// required:true
	TargetID string `json:"target_id"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp string `json:"exp"`

	// swagger:ignore
	PublicKey *rsa.PublicKey
	// swagger:ignore
	PrivateKey *rsa.PrivateKey
}
