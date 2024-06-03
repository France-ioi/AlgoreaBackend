package payloads

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
)

// ScoreToken represents data inside a score token.
type ScoreToken struct {
	// Nullable fields are of pointer types
	Date         string  `json:"date" validate:"dmy-date"` // dd-mm-yyyy
	UserID       string  `json:"idUser"`
	ItemID       *string `json:"idItem,omitempty"`
	LocalItemID  string  `json:"idItemLocal"`
	AttemptID    string  `json:"idAttempt"`
	ItemURL      string  `json:"itemUrl"`
	Score        string  `json:"score"`
	UserAnswerID string  `json:"idUserAnswer"`
	Answer       *string `json:"sAnswer"`

	Converted  ScoreTokenConverted
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// ScoreTokenConverted contains converted field values of ScoreToken payload.
type ScoreTokenConverted struct {
	UserID        int64
	UserAnswerID  int64
	Score         float64
	LocalItemID   int64
	ParticipantID int64
	AttemptID     int64
}

// Bind validates a score token and converts some needed field values (called by ParseMap).
func (tt *ScoreToken) Bind() error {
	var err error
	tt.Converted.UserID, err = strconv.ParseInt(tt.UserID, 10, 64)
	if err != nil {
		return errors.New("wrong idUser")
	}
	tt.Converted.UserAnswerID, err = strconv.ParseInt(tt.UserAnswerID, 10, 64)
	if err != nil {
		return errors.New("wrong idUserAnswer")
	}
	tt.Converted.Score, err = strconv.ParseFloat(tt.Score, 64)
	if err != nil {
		return errors.New("wrong score")
	}
	if tt.AttemptID != "" {
		_, err = fmt.Sscanf(tt.AttemptID, "%d/%d", &tt.Converted.ParticipantID, &tt.Converted.AttemptID)
		if err != nil {
			return errors.New("wrong idAttempt")
		}
	}
	if tt.LocalItemID != "" {
		tt.Converted.LocalItemID, err = strconv.ParseInt(tt.LocalItemID, 10, 64)
		if err != nil {
			return errors.New("wrong idItemLocal")
		}
	}
	return nil
}

var _ Binder = (*ScoreToken)(nil)
