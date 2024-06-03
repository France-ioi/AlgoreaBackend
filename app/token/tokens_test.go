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

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/app/tokentest"
)

var marshalAndSignTests = []struct {
	name        string
	currentTime time.Time
	structType  reflect.Type
	payloadMap  map[string]interface{}
	payloadType reflect.Type
}{
	{
		name:        "task token",
		currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
		structType:  reflect.TypeOf(Task{}),
		payloadMap:  payloadstest.TaskPayloadFromAlgoreaPlatform,
		payloadType: reflect.TypeOf(payloads.TaskToken{}),
	},
	{
		name:        "answer token",
		currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
		structType:  reflect.TypeOf(Answer{}),
		payloadMap:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
		payloadType: reflect.TypeOf(payloads.AnswerToken{}),
	},
	{
		name:        "hint token",
		currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
		structType:  reflect.TypeOf(Hint{}),
		payloadMap:  payloadstest.HintPayloadFromTaskPlatform,
		payloadType: reflect.TypeOf(payloads.HintToken{}),
	},
	{
		name:        "score token",
		currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
		structType:  reflect.TypeOf(Score{}),
		payloadMap:  payloadstest.ScorePayloadFromGrader,
		payloadType: reflect.TypeOf(payloads.ScoreToken{}),
	},
	{
		name:        "thread token",
		currentTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		structType:  reflect.TypeOf(Thread{}),
		payloadMap:  payloadstest.ThreadPayloadFromAlgoreaPlatformOriginal,
		payloadType: reflect.TypeOf(payloads.ThreadToken{}),
	},
}

func TestToken_MarshalJSON(t *testing.T) {
	for _, test := range marshalAndSignTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
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

func TestToken_Sign(t *testing.T) {
	for _, test := range marshalAndSignTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
			defer monkey.UnpatchAll()
			privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
			assert.NoError(t, err)
			publicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
			assert.NoError(t, err)

			payloadRefl := reflect.New(test.payloadType)
			payloadRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			payload := payloadRefl.Interface()
			assert.NoError(t, payloads.ParseMap(test.payloadMap, payload))
			tokenStruct := reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface().(Signer)
			token, err := tokenStruct.Sign(privateKey)
			assert.NoError(t, err)

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(UnmarshalStringer)
			assert.NoError(t, result.UnmarshalString(token))
			assert.Equal(t, reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface(), result)
		})
	}
}

func TestToken_MarshalString(t *testing.T) {
	for _, test := range marshalAndSignTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return test.currentTime })
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
			tokenStruct := reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface()
			token, err := tokenStruct.(MarshalStringer).MarshalString()
			assert.NoError(t, err)
			tokenJSON, err := tokenStruct.(json.Marshaler).MarshalJSON()
			assert.NoError(t, err)

			assert.Equal(t, string(tokenJSON), fmt.Sprintf("%q", token))

			resultRefl := reflect.New(test.structType)
			resultRefl.Elem().FieldByName("PublicKey").Set(reflect.ValueOf(publicKey))
			resultRefl.Elem().FieldByName("PrivateKey").Set(reflect.ValueOf(privateKey))
			result := resultRefl.Interface().(UnmarshalStringer)
			assert.NoError(t, result.UnmarshalString(token))
			assert.Equal(t, reflect.ValueOf(payload).Convert(reflect.PtrTo(test.structType)).Interface(), result)
		})
	}
}

func TestAbstract_UnmarshalJSON_HandlesError(t *testing.T) {
	assert.Equal(t, errors.New("not a compact JWS"), (&Task{}).UnmarshalJSON([]byte(`""`)))
}
