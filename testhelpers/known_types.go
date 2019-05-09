package testhelpers

import (
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
