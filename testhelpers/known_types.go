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

var knownTypes = map[string]reflect.Type{
	"AnswersSubmitRequest":  reflect.TypeOf(&answers.SubmitRequest{}).Elem(),
	"AnswersSubmitResponse": reflect.TypeOf(&answersSubmitResponse{}).Elem(),
}

func getZeroStructPtr(typeName string) (interface{}, error) {
	if _, ok := knownTypes[typeName]; !ok {
		return nil, fmt.Errorf("unknown type: %q", typeName)
	}
	return reflect.New(knownTypes[typeName]).Interface(), nil
}
