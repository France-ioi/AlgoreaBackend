package payloads

import "crypto/rsa"

// AnswerToken represents data inside an answer token
type AnswerToken struct {
	// Nullable fields are of pointer types
	Date           string  `json:"date" valid:"matches(^[0-3][0-9]-[0-1][0-9]-\\d{4}$)"` // dd-mm-yyyy
	UserID         string  `json:"idUser"`
	ItemID         *string `json:"idItem"` // always is nil?
	AttemptID      *string `json:"idAttempt"`
	ItemURL        string  `json:"itemUrl"`
	LocalItemID    string  `json:"idItemLocal"`
	PlatformName   string  `json:"platformName" valid:"stringlength(1|200)"`
	RandomSeed     string  `json:"randomSeed"`
	HintsRequested *string `json:"sHintsRequested"`
	HintsGiven     string  `json:"nbHintsGiven"`
	Answer         string  `json:"sAnswer"`
	UserAnswerID   string  `json:"idUserAnswer"`

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}
