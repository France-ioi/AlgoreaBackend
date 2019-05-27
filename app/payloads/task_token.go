package payloads

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"strconv"

	"github.com/France-ioi/AlgoreaBackend/app/formdata"
)

// TaskToken represents data inside a task token
type TaskToken struct {
	// Nullable fields are of pointer types
	Date               string             `json:"date" valid:"matches(^[0-3][0-9]-[0-1][0-9]-\\d{4}$)"` // dd-mm-yyyy
	UserID             string             `json:"idUser"`
	ItemID             *string            `json:"idItem,omitempty"` // always is nil?
	AttemptID          *string            `json:"idAttempt,omitempty"`
	ItemURL            string             `json:"itemUrl"`
	LocalItemID        string             `json:"idItemLocal"`
	PlatformName       string             `json:"platformName" valid:"stringlength(1|200)"`
	RandomSeed         string             `json:"randomSeed"`
	TaskID             *string            `json:"idTask,omitempty"`        // always is nil?
	HintsAllowed       *string            `json:"bHintsAllowed,omitempty"` // "0" or "1"
	HintPossible       *bool              `json:"bHintPossible,omitempty"`
	HintsRequested     *string            `json:"sHintsRequested,omitempty"`
	HintsGivenCount    *string            `json:"nbHintsGiven,omitempty"`
	AccessSolutions    *formdata.Anything `json:"bAccessSolutions,omitempty"` // "0" or "1" / 0 or 1
	ReadAnswers        *bool              `json:"bReadAnswers,omitempty"`
	Login              *string            `json:"sLogin,omitempty"`
	SubmissionPossible *bool              `json:"bSubmissionPossible,omitempty"`
	SupportedLangProg  *string            `json:"sSupportedLangProg,omitempty"`
	IsAdmin            *string            `json:"bIsAdmin,omitempty"` // "0" or "1"

	Converted TaskTokenConverted

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// TaskTokenConverted contains converted field values of TaskToken payload
type TaskTokenConverted struct {
	UserID          int64
	LocalItemID     int64
	AttemptID       *int64
	AccessSolutions *bool
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
	if tt.AccessSolutions != nil {
		switch {
		case bytes.Equal([]byte("0"), tt.AccessSolutions.Bytes()), bytes.Equal([]byte(`"0"`), tt.AccessSolutions.Bytes()):
			tt.Converted.AccessSolutions = ptrBool(false)
		case bytes.Equal([]byte("1"), tt.AccessSolutions.Bytes()), bytes.Equal([]byte(`"1"`), tt.AccessSolutions.Bytes()):
			tt.Converted.AccessSolutions = ptrBool(true)
		default:
			return errors.New("wrong bAccessSolutions")
		}
	}
	return nil
}

func ptrBool(b bool) *bool { return &b }

var _ Binder = (*TaskToken)(nil)
