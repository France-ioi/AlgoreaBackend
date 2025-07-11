package token

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/SermoDigital/jose/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

func TestHint_UnmarshalString(t *testing.T) {
	hint := Token[payloads.HintToken]{}
	err := payloads.ParseMap(payloadstest.HintPayloadFromTaskPlatform, &hint.Payload)
	assert.NoError(t, err)

	hint.PrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(tokentest.TaskPlatformPrivateKey)
	assert.NoError(t, err)

	hint.PublicKey, err = crypto.ParseRSAPublicKeyFromPEM(tokentest.TaskPlatformPublicKey)
	assert.NoError(t, err)

	marshaled, err := hint.MarshalJSON()
	assert.NoError(t, err)

	var marshaledString string
	assert.NoError(t, json.Unmarshal(marshaled, &marshaledString))

	result := Token[payloads.HintToken]{PublicKey: hint.PublicKey, PrivateKey: hint.PrivateKey}
	err = result.UnmarshalString(marshaledString)
	assert.NoError(t, err)

	hint.Payload.Date = result.Payload.Date
	hint.Payload.Converted.UserID, _ = strconv.ParseInt(hint.Payload.UserID, 10, 64)
	assert.Equal(t, hint, result)
}
