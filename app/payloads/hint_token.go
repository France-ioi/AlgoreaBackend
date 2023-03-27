package payloads

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/France-ioi/AlgoreaBackend/app/formdata"
)

// HintToken represents data inside a hint token.
type HintToken struct {
	Date        string            `json:"date" validate:"dmy-date"` // dd-mm-yyyy
	UserID      string            `json:"idUser,omitempty"`
	ItemID      *string           `json:"idItem,omitempty"`
	LocalItemID string            `json:"idItemLocal"`
	ItemURL     string            `json:"itemUrl"`
	AttemptID   string            `json:"idAttempt"`
	AskedHint   formdata.Anything `json:"askedHint"`

	Converted HintTokenConverted

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// HintTokenConverted contains converted field values of HintToken payload.
type HintTokenConverted struct {
	UserID int64
}

// UnmarshalJSON unmarshals the hint token payload from JSON.
func (tt *HintToken) UnmarshalJSON(raw []byte) error {
	preparsedHintToken := map[string]formdata.Anything{}
	if err := json.Unmarshal(raw, &preparsedHintToken); err != nil {
		return err
	}
	parsedHintToken := make(map[string]interface{}, len(preparsedHintToken))
	for key := range preparsedHintToken {
		if key == "askedHint" {
			parsedHintToken[key] = preparsedHintToken[key]
		} else {
			var value interface{}
			_ = json.Unmarshal(preparsedHintToken[key].Bytes(), &value)
			parsedHintToken[key] = value
		}
	}
	return ParseMap(parsedHintToken, tt)
}

// Bind validates a hint token and converts some needed field values (called by ParseMap).
func (tt *HintToken) Bind() error {
	var err error
	tt.Converted.UserID, err = strconv.ParseInt(tt.UserID, 10, 64)
	if err != nil {
		return errors.New("wrong idUser")
	}
	return nil
}

var _ Binder = (*HintToken)(nil)
