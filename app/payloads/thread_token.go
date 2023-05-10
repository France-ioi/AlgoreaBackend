package payloads

import "crypto/rsa"

// ThreadToken represents data inside a thread token.
type ThreadToken struct {
	Date          string `json:"date" validate:"dmy-date"` // dd-mm-yyyy
	ItemID        string `json:"item_id"`
	ParticipantID string `json:"participant_id"`
	UserID        string `json:"user_id"`   // Current user.
	IsMine        bool   `json:"is_mine"`   // Whether the thread is from the current user.
	CanWatch      bool   `json:"can_watch"` // Whether the current user can post new content.
	CanWrite      bool   `json:"can_write"` // Whether the current user can post new content on the thread
	Exp           string `json:"exp"`       // Expiry date in the number of seconds since 01/01/1970 UTC.

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}
