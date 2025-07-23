package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/SermoDigital/jose/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

type marshalAndSignTest struct {
	name        string
	currentTime time.Time
	structType  reflect.Type
	payloadMap  map[string]interface{}
	payloadType reflect.Type
}

func marshalAndSignTests() []marshalAndSignTest {
	return []marshalAndSignTest{
		{
			name:        "task token",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			structType:  reflect.TypeOf(Token[payloads.TaskToken]{}),
			payloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform(),
			payloadType: reflect.TypeOf(payloads.TaskToken{}),
		},
		{
			name:        "answer token",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			structType:  reflect.TypeOf(Token[payloads.AnswerToken]{}),
			payloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform(),
			payloadType: reflect.TypeOf(payloads.AnswerToken{}),
		},
		{
			name:        "hint token",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			structType:  reflect.TypeOf(Token[payloads.HintToken]{}),
			payloadMap:  payloadstest.HintPayloadFromTaskPlatform(),
			payloadType: reflect.TypeOf(payloads.HintToken{}),
		},
		{
			name:        "score token",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			structType:  reflect.TypeOf(Token[payloads.ScoreToken]{}),
			payloadMap:  payloadstest.ScorePayloadFromGrader(),
			payloadType: reflect.TypeOf(payloads.ScoreToken{}),
		},
		{
			name:        "thread token",
			currentTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			structType:  reflect.TypeOf(Token[payloads.ThreadToken]{}),
			payloadMap:  payloadstest.ThreadPayloadFromAlgoreaPlatformOriginal(),
			payloadType: reflect.TypeOf(payloads.ThreadToken{}),
		},
	}
}

func TestToken_MarshalJSON(t *testing.T) {
	for _, test := range marshalAndSignTests() {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
			defer monkey.UnpatchAll()
			privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
			require.NoError(t, err)
			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			require.NoError(t, err)

			payloadRefl := reflect.New(test.payloadType)
			payload := payloadRefl.Interface()
			require.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStructRefl := reflect.New(test.structType)
			tokenStructRefl.Elem().FieldByName("Payload").Set(payloadRefl.Elem())
			tokenStructRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			tokenStructRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			tokenStruct := tokenStructRefl.Interface()
			token, err := tokenStruct.(json.Marshaler).MarshalJSON()
			require.NoError(t, err)

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(json.Unmarshaler)
			require.NoError(t, result.UnmarshalJSON(token))
			assert.Equal(t, tokenStruct, result)
		})
	}
}

func TestToken_Sign(t *testing.T) {
	for _, test := range marshalAndSignTests() {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
			defer monkey.UnpatchAll()
			privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
			require.NoError(t, err)
			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			require.NoError(t, err)

			tokenStructRefl := reflect.New(test.structType)
			payloadRefl := reflect.New(test.payloadType)
			payload := payloadRefl.Interface()
			require.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStructRefl.Elem().FieldByName("Payload").Set(payloadRefl.Elem())
			tokenStructRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			tokenStruct := tokenStructRefl.Interface()
			token, err := tokenStruct.(Signer).Sign(privateKey)
			require.NoError(t, err)

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(UnmarshalStringer)
			require.NoError(t, result.UnmarshalString(token))
			assert.Equal(t, tokenStruct, result)
		})
	}
}

func TestToken_MarshalString(t *testing.T) {
	for _, test := range marshalAndSignTests() {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
			defer monkey.UnpatchAll()
			privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
			require.NoError(t, err)
			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			require.NoError(t, err)

			payloadRefl := reflect.New(test.payloadType)
			tokenStructRefl := reflect.New(test.structType)
			tokenStructRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			tokenStructRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			tokenStruct := tokenStructRefl.Interface()
			payload := payloadRefl.Interface()
			require.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStructRefl.Elem().FieldByName("Payload").Set(payloadRefl.Elem())
			token, err := tokenStruct.(MarshalStringer).MarshalString()
			require.NoError(t, err)
			tokenJSON, err := tokenStruct.(json.Marshaler).MarshalJSON()
			require.NoError(t, err)

			assert.JSONEq(t, string(tokenJSON), fmt.Sprintf("%q", token))

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(UnmarshalStringer)
			require.NoError(t, result.UnmarshalString(token))
			assert.Equal(t, tokenStruct, result)
		})
	}
}

func TestToken_UnmarshalJSON_HandlesError(t *testing.T) {
	assert.Equal(t, errors.New("not a compact JWS"), (&Token[payloads.TaskToken]{}).UnmarshalJSON([]byte(`""`)))
}
