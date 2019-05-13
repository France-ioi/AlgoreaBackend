package payloads

import "crypto/rsa"

// TaskToken represents data inside a task token
type TaskToken struct {
	// Nullable fields are of pointer types
	Date               string  `json:"date" valid:"matches(^[0-3][0-9]-[0-1][0-9]-\\d{4}$)"` // dd-mm-yyyy
	UserID             string  `json:"idUser"`
	ItemID             *string `json:"idItem"` // always is nil?
	AttemptID          *string `json:"idAttempt"`
	ItemURL            string  `json:"itemUrl"`
	LocalItemID        string  `json:"idItemLocal"`
	PlatformName       string  `json:"platformName" valid:"stringlength(1|200)"`
	RandomSeed         string  `json:"randomSeed"`
	TaskID             *string `json:"idTask"`        // always is nil?
	HintsAllowed       string  `json:"bHintsAllowed"` // "0" or "1"
	HintPossible       bool    `json:"bHintPossible"`
	HintsRequested     *string `json:"sHintsRequested"`
	HintsGiven         string  `json:"nbHintsGiven"`
	AccessSolutions    string  `json:"bAccessSolutions"` // "0" or "1"
	ReadAnswers        bool    `json:"bReadAnswers"`
	Login              string  `json:"sLogin"`
	SubmissionPossible bool    `json:"bSubmissionPossible"`
	SupportedLangProg  string  `json:"sSupportedLangProg"`
	IsAdmin            string  `json:"bIsAdmin"` // "0" or "1"

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}
