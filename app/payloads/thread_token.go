package payloads

import "crypto/rsa"

// ThreadToken represents data inside a thread token.
// swagger:model ThreadToken
type ThreadToken struct {
	// Format dd-mm-yyyy
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// required:true
	ItemID string `json:"item_id"`
	// required:true
	ParticipantID string `json:"participant_id"`
	// Current user.
	// required:true
	UserID string `json:"user_id"`
	// Whether the thread is from the current user.
	// required:true
	IsMine bool `json:"is_mine"`
	// Whether the current user can post new content.
	// required:true
	CanWatch bool `json:"can_watch"`
	// Whether the current user can post new content on the thread
	// required:true
	CanWrite bool `json:"can_write"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp string `json:"exp"`

	// swagger:ignore
	PublicKey *rsa.PublicKey
	// swagger:ignore
	PrivateKey *rsa.PrivateKey
}
