package payloads

import (
	"crypto/rsa"
	"errors"
	"strconv"
)

// ScoreToken represents data inside a score token
type ScoreToken struct {
	// Nullable fields are of pointer types
	Date         string  `json:"date" valid:"matches(^[0-3][0-9]-[0-1][0-9]-\\d{4}$)"` // dd-mm-yyyy
	UserID       string  `json:"idUser"`
	ItemID       *string `json:"idItem,omitempty"`
	AttemptID    *string `json:"idAttempt,omitempty"`
	ItemURL      string  `json:"itemUrl"`
	Score        string  `json:"score"`
	UserAnswerID string  `json:"idUserAnswer"`
	Answer       *string `json:"sAnswer"`

	Converted  ScoreTokenConverted
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// ScoreTokenConverted contains converted field values of ScoreToken payload
type ScoreTokenConverted struct {
	UserID       int64
	UserAnswerID int64
	Score        float64
}

// Bind validates a score token and converts some needed field values (called by ParseMap)
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
	return nil
}

var _ Binder = (*HintToken)(nil)
