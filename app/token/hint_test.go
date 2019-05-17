package token

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/jose.v1/crypto"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/app/tokentest"
)

func TestHint_UnmarshalString(t *testing.T) {
	hint := Hint{}
	err := payloads.ParseMap(payloadstest.HintPayloadFromTaskPlatform, &hint)
	assert.NoError(t, err)

	hint.PrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(tokentest.TaskPlatformPrivateKey)
	assert.NoError(t, err)

	hint.PublicKey, err = crypto.ParseRSAPublicKeyFromPEM(tokentest.TaskPlatformPublicKey)
	assert.NoError(t, err)

	marshaled, err := hint.MarshalJSON()
	assert.NoError(t, err)

	var marshaledString string
	assert.NoError(t, json.Unmarshal(marshaled, &marshaledString))

	result := Hint{PublicKey: hint.PublicKey, PrivateKey: hint.PrivateKey}
	err = result.UnmarshalString(marshaledString)
	assert.NoError(t, err)

	hint.Date = result.Date
	if hint.UserID != nil {
		convertedUserID, _ := strconv.ParseInt(*hint.UserID, 10, 64)
		hint.Converted.UserID = &convertedUserID
	}
	assert.Equal(t, hint, result)
}
