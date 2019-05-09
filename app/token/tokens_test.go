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
			structType:          reflect.TypeOf(TaskToken{}),
			token:               []byte(fmt.Sprintf("%q", taskTokenFromAlgoreaPlatform)),
			expectedPayloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform,
			expectedPayloadType: reflect.TypeOf(payloads.TaskTokenPayload{}),
		},
		{
			name:                "answer token",
			structType:          reflect.TypeOf(AnswerToken{}),
			token:               []byte(fmt.Sprintf("%q", answerTokenFromAlgoreaPlatform)),
			expectedPayloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
			expectedPayloadType: reflect.TypeOf(payloads.AnswerTokenPayload{}),
		},
		{
			name:                 "invalid JSON string",
			structType:           reflect.TypeOf(AnswerToken{}),
			token:                []byte(""),
			expectedErrorMessage: "invalid token: unexpected end of JSON input",
			expectedPayloadType:  reflect.TypeOf(payloads.AnswerTokenPayload{}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now,
				func() time.Time { return time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC) })
			defer monkey.UnpatchAll()

			var err error
			platformPublicKey, err = crypto.ParseRSAPublicKeyFromPEM(algoreaPlatformPublicKey)
			platformName = testPlatformName
			assert.NoError(t, err)

			expectedPayload := reflect.New(test.expectedPayloadType).Interface()
			assert.NoError(t, payloads.ParseMap(test.expectedPayloadMap, expectedPayload))

			payload := reflect.New(test.structType).Interface().(json.Unmarshaler)
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
			structType:  reflect.TypeOf(TaskToken{}),
			payloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform,
			payloadType: reflect.TypeOf(payloads.TaskTokenPayload{}),
		},
		{
			name:        "answer token",
			structType:  reflect.TypeOf(AnswerToken{}),
			payloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
			payloadType: reflect.TypeOf(payloads.AnswerTokenPayload{}),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC) })
			defer monkey.UnpatchAll()
			var err error
			platformPrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(algoreaPlatformPrivateKey)
			assert.NoError(t, err)
			platformName = "test_dmitry"

			payload := reflect.New(test.payloadType).Interface()
			assert.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStruct := reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface().(json.Marshaler)
			token, err := tokenStruct.MarshalJSON()
			assert.NoError(t, err)

			result := reflect.New(test.structType).Interface().(json.Unmarshaler)
			assert.NoError(t, result.UnmarshalJSON(token))
			assert.Equal(t, result, reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface())
		})
	}
}

func TestAbstract_UnmarshalJSON_HandlesError(t *testing.T) {
	assert.Equal(t, errors.New("invalid token: not a compact JWS"), (&TaskToken{}).UnmarshalJSON([]byte(`""`)))
}
