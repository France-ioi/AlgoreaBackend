//go:build !prod

package testhelpers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/answers"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

type answersSubmitResponse struct {
	Data struct {
		AnswerToken token.Token[payloads.AnswerToken] `json:"answer_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`

	PublicKey *rsa.PublicKey
}

type answersSubmitResponseWrapper struct {
	Data struct {
		AnswerToken *string `json:"answer_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (resp *answersSubmitResponse) UnmarshalJSON(raw []byte) error {
	wrapper := answersSubmitResponseWrapper{}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	resp.Message = wrapper.Message
	resp.Success = wrapper.Success
	if wrapper.Data.AnswerToken != nil {
		resp.Data.AnswerToken.PublicKey = resp.PublicKey
		return (&resp.Data.AnswerToken).UnmarshalString(*wrapper.Data.AnswerToken)
	}
	return nil
}

type responseWithTaskToken struct {
	Data struct {
		TaskToken token.Token[payloads.TaskToken] `json:"task_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`

	PublicKey *rsa.PublicKey
}

type responseWithTaskTokenWrapper struct {
	Data struct {
		TaskToken *string `json:"task_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (resp *responseWithTaskToken) UnmarshalJSON(raw []byte) error {
	wrapper := responseWithTaskTokenWrapper{}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	resp.Message = wrapper.Message
	resp.Success = wrapper.Success
	if wrapper.Data.TaskToken != nil {
		resp.Data.TaskToken.PublicKey = resp.PublicKey
		return (&resp.Data.TaskToken).UnmarshalString(*wrapper.Data.TaskToken)
	}
	return nil
}

type saveGradeResponse struct {
	Data struct {
		TaskToken token.Token[payloads.TaskToken] `json:"task_token"`
		Validated bool                            `json:"validated"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`

	PublicKey *rsa.PublicKey
}

type saveGradeResponseWrapper struct {
	Data struct {
		TaskToken *string `json:"task_token"`
		Validated bool    `json:"validated"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (resp *saveGradeResponse) UnmarshalJSON(raw []byte) error {
	wrapper := saveGradeResponseWrapper{}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	resp.Message = wrapper.Message
	resp.Success = wrapper.Success
	resp.Data.Validated = wrapper.Data.Validated
	if wrapper.Data.TaskToken != nil {
		resp.Data.TaskToken.PublicKey = resp.PublicKey
		return (&resp.Data.TaskToken).UnmarshalString(*wrapper.Data.TaskToken)
	}
	return nil
}

type threadGetResponse struct {
	ParticipantID int64  `json:"participant_id"`
	ItemID        int64  `json:"item_id"`
	Status        string `json:"status"`

	ThreadToken token.Token[payloads.ThreadToken] `json:"token"`

	PublicKey *rsa.PublicKey
}

type threadGetResponseWrapper struct {
	ParticipantID int64  `json:"participant_id"`
	ItemID        int64  `json:"item_id"`
	Status        string `json:"status"`

	ThreadToken *string `json:"token"`
}

func (resp *threadGetResponse) UnmarshalJSON(raw []byte) error {
	wrapper := threadGetResponseWrapper{}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}

	resp.ItemID = wrapper.ItemID
	resp.ParticipantID = wrapper.ParticipantID
	resp.Status = wrapper.Status

	if wrapper.ThreadToken != nil {
		resp.ThreadToken.PublicKey = resp.PublicKey
		return (&resp.ThreadToken).UnmarshalString(*wrapper.ThreadToken)
	}

	return nil
}

//nolint:gochecknoglobals // knownTypes is a map of type names to their reflect.Type.
var knownTypes = map[string]reflect.Type{
	"AnswersSubmitRequest":      reflect.TypeOf(&answers.SubmitRequest{}).Elem(),
	"AnswersSubmitResponse":     reflect.TypeOf(&answersSubmitResponse{}).Elem(),
	"AskHintResponse":           reflect.TypeOf(&responseWithTaskToken{}).Elem(),
	"SaveGradeResponse":         reflect.TypeOf(&saveGradeResponse{}).Elem(),
	"GenerateTaskTokenResponse": reflect.TypeOf(&responseWithTaskToken{}).Elem(),
	"ThreadTokenResponse":       reflect.TypeOf(&threadGetResponse{}).Elem(),
}

func getZeroStructPtr(typeName string) (interface{}, error) {
	reflType, ok := knownTypes[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown type: %q", typeName)
	}
	return reflect.New(reflType).Interface(), nil
}
