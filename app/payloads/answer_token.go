package payloads

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
)

// AnswerToken represents data inside an answer token.
// idAttempt is required.
type AnswerToken struct {
	// Nullable fields are of pointer types
	Date            string  `json:"date"            validate:"dmy-date"` // dd-mm-yyyy
	UserID          string  `json:"idUser"`
	ItemID          *string `json:"idItem"` // always is nil?
	AttemptID       string  `json:"idAttempt"`
	ItemURL         string  `json:"itemUrl"`
	LocalItemID     string  `json:"idItemLocal"`
	PlatformName    string  `json:"platformName"    validate:"min=1,max=200"` // 1 <= length <= 200
	RandomSeed      string  `json:"randomSeed"`
	HintsRequested  *string `json:"sHintsRequested"`
	HintsGivenCount string  `json:"nbHintsGiven"`
	Answer          string  `json:"sAnswer"`
	UserAnswerID    string  `json:"idUserAnswer"`

	Converted  AnswerTokenConverted
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// AnswerTokenConverted contains converted field values of AnswerToken payload.
type AnswerTokenConverted struct {
	UserID        int64
	UserAnswerID  int64
	LocalItemID   int64
	ParticipantID int64
	AttemptID     int64
}

// Bind validates a score token and converts some needed field values (called by ParseMap).
func (tt *AnswerToken) Bind() error {
	var err error
	tt.Converted.UserID, err = strconv.ParseInt(tt.UserID, 10, 64)
	if err != nil {
		return errors.New("wrong idUser")
	}
	tt.Converted.UserAnswerID, err = strconv.ParseInt(tt.UserAnswerID, 10, 64)
	if err != nil {
		return errors.New("wrong idUserAnswer")
	}
	_, err = fmt.Sscanf(tt.AttemptID, "%d/%d", &tt.Converted.ParticipantID, &tt.Converted.AttemptID)
	if err != nil {
		return errors.New("wrong idAttempt")
	}
	tt.Converted.LocalItemID, err = strconv.ParseInt(tt.LocalItemID, 10, 64)
	if err != nil {
		return errors.New("wrong idItemLocal")
	}
	return nil
}

var _ Binder = (*AnswerToken)(nil)
