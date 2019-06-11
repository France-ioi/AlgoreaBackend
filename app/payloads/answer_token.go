package payloads

import "crypto/rsa"

// AnswerToken represents data inside an answer token.
// idAttempt is required.
type AnswerToken struct {
	// Nullable fields are of pointer types
	Date            string  `json:"date" validate:"dmy-date"` // dd-mm-yyyy
	UserID          string  `json:"idUser"`
	ItemID          *string `json:"idItem"` // always is nil?
	AttemptID       string  `json:"idAttempt"`
	ItemURL         string  `json:"itemUrl"`
	LocalItemID     string  `json:"idItemLocal"`
	PlatformName    string  `json:"platformName" validate:"min=1,max=200"`
	RandomSeed      string  `json:"randomSeed"`
	HintsRequested  *string `json:"sHintsRequested"`
	HintsGivenCount string  `json:"nbHintsGiven"`
	Answer          string  `json:"sAnswer"`
	UserAnswerID    string  `json:"idUserAnswer"`

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}
