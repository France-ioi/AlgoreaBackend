package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jose.v1/crypto"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/app/tokentest"
)

func Test_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name                 string
		structType           reflect.Type
		token                []byte
		expectedPayloadMap   map[string]interface{}
		expectedPayloadType  reflect.Type
		expectedErrorMessage string
	}{
		{
			name:                "task token",
			structType:          reflect.TypeOf(Task{}),
			token:               []byte(fmt.Sprintf("%q", taskTokenFromAlgoreaPlatform)),
			expectedPayloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform,
			expectedPayloadType: reflect.TypeOf(payloads.TaskToken{}),
		},
		{
			name:                "answer token",
			structType:          reflect.TypeOf(Answer{}),
			token:               []byte(fmt.Sprintf("%q", answerTokenFromAlgoreaPlatform)),
			expectedPayloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
			expectedPayloadType: reflect.TypeOf(payloads.AnswerToken{}),
		},
		{
			name:                 "invalid JSON string",
			structType:           reflect.TypeOf(Answer{}),
			token:                []byte(""),
			expectedErrorMessage: "unexpected end of JSON input",
			expectedPayloadType:  reflect.TypeOf(payloads.AnswerToken{}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now,
				func() time.Time { return time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC) })
			defer monkey.UnpatchAll()

			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			assert.NoError(t, err)

			expectedPayloadRefl := reflect.New(test.expectedPayloadType)
			expectedPayloadRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			expectedPayload := expectedPayloadRefl.Interface()
			assert.NoError(t, payloads.ParseMap(test.expectedPayloadMap, expectedPayload))

			payloadRefl := reflect.New(test.structType)
			payloadRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			payload := payloadRefl.Interface().(json.Unmarshaler)
			err = payload.UnmarshalJSON(test.token)
			if test.expectedErrorMessage == "" {
				assert.NoError(t, err)
			} else {
				errMessage := ""
				if err != nil {
					errMessage = err.Error()
				}
				assert.Equal(t, test.expectedErrorMessage, errMessage)
			}
			assert.Equal(t, expectedPayload,
				reflect.ValueOf(payload).Convert(reflect.PtrTo(test.expectedPayloadType)).Interface())
		})
	}
}

func TestTaskToken_MarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		structType  reflect.Type
		payloadMap  map[string]interface{}
		payloadType reflect.Type
	}{
		{
			name:        "task token",
			structType:  reflect.TypeOf(Task{}),
			payloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform,
			payloadType: reflect.TypeOf(payloads.TaskToken{}),
		},
		{
			name:        "answer token",
			structType:  reflect.TypeOf(Answer{}),
			payloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
			payloadType: reflect.TypeOf(payloads.AnswerToken{}),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC) })
			defer monkey.UnpatchAll()
			privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
			assert.NoError(t, err)
			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			assert.NoError(t, err)

			payloadRefl := reflect.New(test.payloadType)
			payloadRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			payloadRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			payload := payloadRefl.Interface()
			assert.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStruct := reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface().(json.Marshaler)
			token, err := tokenStruct.MarshalJSON()
			assert.NoError(t, err)

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(json.Unmarshaler)
			assert.NoError(t, result.UnmarshalJSON(token))
			assert.Equal(t, reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface(), result)
		})
	}
}

func TestAbstract_UnmarshalJSON_HandlesError(t *testing.T) {
	assert.Equal(t, errors.New("not a compact JWS"), (&Task{}).UnmarshalJSON([]byte(`""`)))
}
