// +build !prod

package testhelpers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/France-ioi/AlgoreaBackend/app/api/answers"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

type answersSubmitResponse struct {
	Data struct {
		AnswerToken token.Answer `json:"answer_token"`
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

type askHintResponse struct {
	Data struct {
		TaskToken token.Task `json:"task_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`

	PublicKey *rsa.PublicKey
}

type askHintResponseWrapper struct {
	Data struct {
		TaskToken *string `json:"task_token"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (resp *askHintResponse) UnmarshalJSON(raw []byte) error {
	wrapper := askHintResponseWrapper{}
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

type getTaskTokenResponse struct {
	TaskToken token.Task `json:"task_token"`

	PublicKey *rsa.PublicKey
}

type getTaskTokenResponseWrapper struct {
	TaskToken *string `json:"task_token"`
}

func (resp *getTaskTokenResponse) UnmarshalJSON(raw []byte) error {
	wrapper := getTaskTokenResponseWrapper{}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return err
	}
	if wrapper.TaskToken != nil {
		resp.TaskToken.PublicKey = resp.PublicKey
		return (&resp.TaskToken).UnmarshalString(*wrapper.TaskToken)
	}
	return nil
}

type saveGradeResponse struct {
	Data struct {
		TaskToken token.Task `json:"task_token"`
		Validated bool       `json:"validated"`
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

var knownTypes = map[string]reflect.Type{
	"AnswersSubmitRequest":  reflect.TypeOf(&answers.SubmitRequest{}).Elem(),
	"AnswersSubmitResponse": reflect.TypeOf(&answersSubmitResponse{}).Elem(),
	"AskHintResponse":       reflect.TypeOf(&askHintResponse{}).Elem(),
	"SaveGradeResponse":     reflect.TypeOf(&saveGradeResponse{}).Elem(),
	"GetTaskTokenResponse":  reflect.TypeOf(&getTaskTokenResponse{}).Elem(),
}

func getZeroStructPtr(typeName string) (interface{}, error) {
	if _, ok := knownTypes[typeName]; !ok {
		return nil, fmt.Errorf("unknown type: %q", typeName)
	}
	return reflect.New(knownTypes[typeName]).Interface(), nil
}
