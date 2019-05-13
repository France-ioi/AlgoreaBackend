package payloads

import (
	"crypto/rsa"
	"errors"
	"strconv"
)

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
	HintsGivenCount    string  `json:"nbHintsGiven"`
	AccessSolutions    string  `json:"bAccessSolutions"` // "0" or "1"
	ReadAnswers        bool    `json:"bReadAnswers"`
	Login              string  `json:"sLogin"`
	SubmissionPossible bool    `json:"bSubmissionPossible"`
	SupportedLangProg  string  `json:"sSupportedLangProg"`
	IsAdmin            string  `json:"bIsAdmin"` // "0" or "1"

	Converted TaskTokenConverted

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// TaskTokenConverted contains converted field values of TaskToken payload
type TaskTokenConverted struct {
	UserID      int64
	LocalItemID int64
	AttemptID   *int64
}

// Bind validates a task token and converts some needed field values.
func (tt *TaskToken) Bind() error {
	var err error
	tt.Converted.UserID, err = strconv.ParseInt(tt.UserID, 10, 64)
	if err != nil {
		return errors.New("wrong idUser")
	}
	tt.Converted.LocalItemID, err = strconv.ParseInt(tt.LocalItemID, 10, 64)
	if err != nil {
		return errors.New("wrong idItemLocal")
	}

	if tt.AttemptID != nil {
		var attemptID int64
		attemptID, err = strconv.ParseInt(*tt.AttemptID, 10, 64)
		if err != nil {
			return errors.New("wrong idAttempt")
		}
		tt.Converted.AttemptID = &attemptID
	}
	return nil
}

var _ Binder = (*TaskToken)(nil)
