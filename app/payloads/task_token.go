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
	ItemID             *string `json:"idItem,omitempty"` // always is nil?
	AttemptID          string  `json:"idAttempt"`
	ItemURL            string  `json:"itemUrl"`
	LocalItemID        string  `json:"idItemLocal"`
	PlatformName       string  `json:"platformName" valid:"stringlength(1|200)"`
	RandomSeed         string  `json:"randomSeed"`
	TaskID             *string `json:"idTask,omitempty"` // always is nil?
	HintsAllowed       *bool   `json:"bHintsAllowed,omitempty"`
	HintPossible       *bool   `json:"bHintPossible,omitempty"`
	HintsRequested     *string `json:"sHintsRequested,omitempty"`
	HintsGivenCount    *string `json:"nbHintsGiven,omitempty"`
	AccessSolutions    *bool   `json:"bAccessSolutions,omitempty"`
	ReadAnswers        *bool   `json:"bReadAnswers,omitempty"`
	Login              *string `json:"sLogin,omitempty"`
	SubmissionPossible *bool   `json:"bSubmissionPossible,omitempty"`
	SupportedLangProg  *string `json:"sSupportedLangProg,omitempty"`
	IsAdmin            *bool   `json:"bIsAdmin,omitempty"`

	Converted TaskTokenConverted

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// TaskTokenConverted contains converted field values of TaskToken payload
type TaskTokenConverted struct {
	UserID      int64
	LocalItemID int64
	AttemptID   int64
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

	tt.Converted.AttemptID, err = strconv.ParseInt(tt.AttemptID, 10, 64)
	if err != nil {
		return errors.New("wrong idAttempt")
	}
	return nil
}

func ptrBool(b bool) *bool { return &b }

var _ Binder = (*TaskToken)(nil)
