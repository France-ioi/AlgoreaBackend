package token

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jose.v1/crypto"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/payloadstest"
)

var taskTokenFromAlgoreaPlatform = []byte(
	"eyJhbGciOiJSUzUxMiJ9.eyJzTG9naW4iOiJ0ZXN0IiwiYklzQWRtaW4iOiIwIiwiaWRV" +
		"c2VyIjoiNTU2MzcxODIxNjkzMjE5OTI1IiwiaWRJdGVtIjpudWxsLCJzSGludHNSZXF" +
		"1ZXN0ZWQiOm51bGwsImJIaW50c0FsbG93ZWQiOiIwIiwic1N1cHBvcnRlZExhbmdQcm" +
		"9nIjoiKiIsImJBY2Nlc3NTb2x1dGlvbnMiOiIxIiwiaXRlbVVybCI6Imh0dHA6XC9cL" +
		"3Rhc2twbGF0Zm9ybS5tYmxvY2tlbGV0LmluZm9cL3Rhc2suaHRtbD90YXNrSWQ9NDAz" +
		"NDQ5NTQzNjcyMTgzOTM2IiwiaWRJdGVtTG9jYWwiOiI5MDE3NTY1NzMzNDU4MzE0MDk" +
		"iLCJiU3VibWlzc2lvblBvc3NpYmxlIjp0cnVlLCJpZEF0dGVtcHQiOm51bGwsIm5iSG" +
		"ludHNHaXZlbiI6IjAiLCJiSGludFBvc3NpYmxlIjp0cnVlLCJpZFRhc2siOm51bGwsI" +
		"mJSZWFkQW5zd2VycyI6dHJ1ZSwicmFuZG9tU2VlZCI6IjU1NjM3MTgyMTY5MzIxOTky" +
		"NSIsInBsYXRmb3JtTmFtZSI6InRlc3RfZG1pdHJ5IiwiZGF0ZSI6IjAyLTA1LTIwMTk" +
		"ifQ.jyAWVPyW442LQAcAAG8F8NddmKtRJLaNhpSsR7WZIDrkXro6G25ZL4oFxzQuMZp" +
		"k2xMkkSRyK3bjM0uOOQ0F6yDZOkJ3TiSbJe-tROUdcPP3xgGsoc8eOK2_KNLoXg49u9" +
		"Jtg-C1Yru04pXF9nEsm2FLB9n-Rg-cLCmPxbVCm_U")

var answerTokenFromAlgoreaPlatform = []byte(
	"eyJhbGciOiJSUzUxMiJ9.eyJzQW5zd2VyIjoie1wiaWRTdWJtaXNzaW9uXCI6XCI4OTkx" +
		"NDYzMDkyMDM4NTUwNzRcIixcImxhbmdQcm9nXCI6XCJweXRob25cIixcInNvdXJjZUN" +
		"vZGVcIjpcInByaW50KG1pbihpbnQoaW5wdXQoKSksIGludChpbnB1dCgpKSwgaW50KG" +
		"lucHV0KCkpKSlcIn0iLCJpZFVzZXIiOiI1NTYzNzE4MjE2OTMyMTk5MjUiLCJpZEl0Z" +
		"W0iOm51bGwsImlkQXR0ZW1wdCI6bnVsbCwiaXRlbVVybCI6Imh0dHA6XC9cL3Rhc2tw" +
		"bGF0Zm9ybS5tYmxvY2tlbGV0LmluZm9cL3Rhc2suaHRtbD90YXNrSWQ9NDAzNDQ5NTQ" +
		"zNjcyMTgzOTM2IiwiaWRJdGVtTG9jYWwiOiI5MDE3NTY1NzMzNDU4MzE0MDkiLCJpZF" +
		"VzZXJBbnN3ZXIiOiIyNTE1MTAwMjcxMzg3MjY4NTciLCJwbGF0Zm9ybU5hbWUiOiJ0Z" +
		"XN0X2RtaXRyeSIsInJhbmRvbVNlZWQiOiI1NTYzNzE4MjE2OTMyMTk5MjUiLCJzSGlu" +
		"dHNSZXF1ZXN0ZWQiOm51bGwsIm5iSGludHNHaXZlbiI6IjAiLCJkYXRlIjoiMDItMDU" +
		"tMjAxOSJ9.GriSv4nj0M0CHPuUSAWs31Wv-VPAm494rGL6RrAnrmg5Q5DNBhT8_RGua" +
		"pU5rhaTUHuWr3iwWZYEVqWVrFbuDbKmkKrwCCA6-j6NinWqzGG61EaunxpKXYDQjFOn" +
		"uH8E1PWMKrC6OLk-5J4NUE5qGn87WKbOpbzuzwcUdWJV77o")

var algoreaPlatformPublicKey = []byte(
	"-----BEGIN PUBLIC KEY-----\n" +
		"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDfsh3Rj/IAQ75LB7c8riFTYrgS\n" +
		"0FCDwZhYIPYgmqVWGPK7JX5KnrcTYqxr0e6nqD5e4anMIVyMUn7g+W9ULLa5QrFr\n" +
		"aJw7il+r1XPyadsPGe2C+YVqbSv33TRxTL03mzvlsLL+JlNvM7j0iJ/KGclLPHUz\n" +
		"fiE7YZDILwmultaYFQIDAQAB\n" +
		"-----END PUBLIC KEY-----")

var algoreaPlatformPrivateKey = []byte(
	"-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIICXAIBAAKBgQDfsh3Rj/IAQ75LB7c8riFTYrgS0FCDwZhYIPYgmqVWGPK7JX5K\n" +
		"nrcTYqxr0e6nqD5e4anMIVyMUn7g+W9ULLa5QrFraJw7il+r1XPyadsPGe2C+YVq\n" +
		"bSv33TRxTL03mzvlsLL+JlNvM7j0iJ/KGclLPHUzfiE7YZDILwmultaYFQIDAQAB\n" +
		"AoGALEiomonykJbYnyXh4oNeWZGbey3+Inc634d28jFrNcYul1nuzHrrJ01LcPTY\n" +
		"WBx4bHQkFyMrnSPftk3q+jD34wpCEiBMFJmZk/Exj8ypRvN9K4+oJtMjvx3tcuyB\n" +
		"fnFRvf1J2sTL7F499xv+/UHAIGfyIvyYHLg/SV+aBaHDJmkCQQD3VqeDRTiMul5p\n" +
		"hDc4RbNLgWS3u1KT2U615OcTJZsFVzHuL6LhkxKLsc+rUWNurY0vOkwz4Bra2CpZ\n" +
		"klb/pVFvAkEA54eCYQ3UHUq+HUGFAX7fPokunjf9V+khU5PfvkzFI1O6DbvT5VCe\n" +
		"H4RVzM787lOy17TyIMvGqSIcLbf1hyekuwJAbUT6IlM9ZWaceS8xGgoo6K2Uals2\n" +
		"Yxz42gDzWREfCF/6Lgkbg15vLgny/fOp4uaHXhr6OVzDYHVpWEL/bleBvwJADwAS\n" +
		"jGMu+O7cvlx+V4h2wkB1Cr8p5MYv6JBOELA8nXtRNI6UveipNfWG8Yv/ixlVHvCU\n" +
		"N1e8eTzCgpvGhokk/QJBAJv1h/9jNOB9H9GIf3sB0cRLzH6po6aQX1gEYRZP6hIw\n" +
		"KGHLOGPIBt1FHY5Z0WtQ4vaFtwOEPj5BCPLGP9cvLIs=\n" +
		"-----END RSA PRIVATE KEY-----")

const testPlatformName = "test_dmitry"

func TestParseAndValidate(t *testing.T) {
	tests := []struct {
		name        string
		currentTime time.Time
		token       []byte
		wantPayload map[string]interface{}
		wantError   error
	}{
		{
			name:        "a task token generated by AlgoreaPlatform",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			token:       taskTokenFromAlgoreaPlatform,
			wantPayload: payloadstest.TaskPayloadFromAlgoreaPlatform,
			wantError:   nil,
		},
		{
			name:        "a task token generated by AlgoreaPlatform has expired",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC).Add(+36 * time.Hour),
			token:       taskTokenFromAlgoreaPlatform,
			wantPayload: nil,
			wantError:   errors.New("the token has expired"),
		},
		{
			name:        "a task token generated by AlgoreaPlatform has not started",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC).Add(-36*time.Hour - 1),
			token:       taskTokenFromAlgoreaPlatform,
			wantPayload: nil,
			wantError:   errors.New("the token has expired"),
		},
		{
			name:        "an answer token generated by AlgoreaPlatform",
			currentTime: time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			token:       answerTokenFromAlgoreaPlatform,
			wantPayload: payloadstest.AnswerPayloadFromAlgoreaPlatform,
			wantError:   nil,
		},
		{
			name: "invalid token",
			token: []byte("eyJhbGciOiJSUzUxMiJ9.eyJzTG9naW4iOiJtYmxvY2tlbGV0IiwiYklzQWRtaW4iOiIwIiwiaWRVc2V" +
				"yIjoiODMxOTM1MTM0OTEzMzQ0OSIsImlkSXRlbSI6bnVsbCwic0hpbnRzUmVxdWVzdGVkIjoiW3tcInJvdG9ySW5kZXh" +
				"cIjowLFwiY2VsbFJhbmtcIjowfV0iLCJiSGludHNBbGxvd2VkIjoiMCIsInNTdXBwb3J0ZWRMYW5nUHJvZyI6IioiLCJ" +
				"iQWNjZXNzU29sdXRpb25zIjoiMSIsIml0ZW1VcmwiOiJodHRwczpcL1wvc3RhdGljLWl0ZW1zLmFsZ29yZWEub3JnXC8" +
				"yMDE4LWVuaWdtYVwvP3Rhc2tJRD1odHRwJTNBJTJGJTJGY29uY291cnMtYWxraW5kaS5mciUyRnRhc2tzJTJGMjAxOCU" +
				"yRmVuaWdtYSZ2ZXJzaW9uPTEiLCJpZEl0ZW1Mb2NhbCI6IjE5NzcxNjA0MDYyMTk0OTg0NSIsImJTdWJtaXNzaW9uUG9" +
				"zc2libGUiOnRydWUsImlkQXR0ZW1wdCI6IjI2NzQ5NDAzMDM4MjUzMjc0NSIsIm5iSGludHNHaXZlbiI6IjEiLCJiSGl" +
				"udFBvc3NpYmxlIjp0cnVlLCJpZFRhc2siOm51bGwsImJSZWFkQW5zd2VycyI6dHJ1ZSwicmFuZG9tU2VlZCI6IjI2NzQ" +
				"5NDAzMDM4MjUzMjc0NSIsInBsYXRmb3JtTmFtZSI6Imh0dHA6XC9cL2FsZ29yZWEucGVtLmRldiIsImRhdGUiOiIwMi0" +
				"wNS0yMDE5In0.2Ay1D3adWMhKLldMSWoVftE8584HGkKzSNMFHx-YgCC8TsFSnIANGYCH2VGwbubt5tw8EMif4NMqplM" +
				"e1ROK81N6nk-wPH-cxW-N9qwZvGFFh7PfgDBIQiuYbk-DHid9gGTf4oIOkb-6lD9GjPe4QNZM9zhVWarC-5xzTbWbdUg"),
			wantError: errors.New("invalid token: crypto/rsa: verification error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return tt.currentTime })
			defer monkey.UnpatchAll()
			var err error
			platformPublicKey, err = crypto.ParseRSAPublicKeyFromPEM(algoreaPlatformPublicKey)
			platformName = testPlatformName
			assert.NoError(t, err)
			payload, apiErr := ParseAndValidate(tt.token)
			assert.Equal(t, tt.wantError, apiErr)
			assert.Equal(t, tt.wantPayload, payload)
		})
	}
}

func Test_GenerateToken(t *testing.T) {
	tests := []struct {
		name         string
		platformName string
		currentTime  time.Time
		payload      map[string]interface{}
	}{
		{
			name:         "task payload",
			platformName: testPlatformName,
			currentTime:  time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			payload:      payloadstest.TaskPayloadFromAlgoreaPlatform,
		},
		{
			name:         "answer payload",
			platformName: testPlatformName,
			currentTime:  time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC),
			payload:      payloadstest.AnswerPayloadFromAlgoreaPlatform,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			monkey.Patch(time.Now, func() time.Time { return tt.currentTime })
			defer monkey.UnpatchAll()

			patchedPayload := make(map[string]interface{}, len(tt.payload))
			for k := range tt.payload {
				patchedPayload[k] = tt.payload[k]
			}
			delete(patchedPayload, "date")
			delete(patchedPayload, "platformName")

			var err error
			platformPrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(algoreaPlatformPrivateKey)
			assert.NoError(t, err)
			platformPublicKey, err = crypto.ParseRSAPublicKeyFromPEM(algoreaPlatformPublicKey)
			assert.NoError(t, err)
			platformName = tt.platformName
			token := Generate(patchedPayload)
			payload, err := ParseAndValidate(token)
			assert.NoError(t, err)
			assert.Equal(t, tt.payload, payload)
		})
	}
}

func Test_GenerateToken_PanicsOnError(t *testing.T) {
	platformPrivateKey = &rsa.PrivateKey{D: &big.Int{}, PublicKey: rsa.PublicKey{N: &big.Int{}}}
	defer func() {
		e := recover()
		assert.Equal(t, errors.New("crypto/rsa: message too long for RSA public key size"), e)
	}()
	Generate(map[string]interface{}{})
}

func Test_Initialize_LoadsKeysFromFile(t *testing.T) {
	tmpFilePublic, err := createTmpPublicKeyFile(algoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	assert.NoError(t, err)

	tmpFilePrivate, err := createTmpPrivateKeyFile(algoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	assert.NoError(t, err)

	expectedPrivateKey, err := crypto.ParseRSAPrivateKeyFromPEM(algoreaPlatformPrivateKey)
	assert.NoError(t, err)
	expectedPublicKey, err := crypto.ParseRSAPublicKeyFromPEM(algoreaPlatformPublicKey)
	assert.NoError(t, err)

	platformPrivateKey = nil
	platformPublicKey = nil
	platformName = ""
	err = Initialize(&config.Root{Platform: config.Platform{
		PrivateKeyFile: tmpFilePrivate.Name(),
		PublicKeyFile:  tmpFilePublic.Name(),
		Name:           "my platform",
	}})
	assert.NoError(t, err)
	assert.Equal(t, expectedPrivateKey, platformPrivateKey)
	assert.Equal(t, expectedPublicKey, platformPublicKey)
	assert.Equal(t, platformName, "my platform")
}

func Test_Initialize_CannotLoadPublicKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile(algoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	assert.NoError(t, err)

	err = Initialize(&config.Root{Platform: config.Platform{
		PrivateKeyFile: tmpFilePrivate.Name(),
		PublicKeyFile:  "nosuchfile.pem",
		Name:           "my platform",
	}})
	assert.IsType(t, &os.PathError{}, err)
}

func Test_Initialize_CannotLoadPrivateKey(t *testing.T) {
	tmpFilePublic, err := createTmpPublicKeyFile(algoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	assert.NoError(t, err)

	err = Initialize(&config.Root{Platform: config.Platform{
		PrivateKeyFile: "nosuchfile.pem",
		PublicKeyFile:  tmpFilePublic.Name(),
		Name:           "my platform",
	}})
	assert.IsType(t, &os.PathError{}, err)
}

func Test_Initialize_CannotParsePublicKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile(algoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	assert.NoError(t, err)

	tmpFilePublic, err := createTmpPublicKeyFile([]byte{})
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	assert.NoError(t, err)

	err = Initialize(&config.Root{Platform: config.Platform{
		PrivateKeyFile: tmpFilePrivate.Name(),
		PublicKeyFile:  tmpFilePublic.Name(),
		Name:           "my platform",
	}})
	assert.Equal(t, errors.New("invalid key: Key must be PEM encoded PKCS1 or PKCS8 private key"), err)
}

func Test_Initialize_CannotParsePrivateKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile([]byte{})
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	assert.NoError(t, err)

	tmpFilePublic, err := createTmpPublicKeyFile(algoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	assert.NoError(t, err)

	err = Initialize(&config.Root{Platform: config.Platform{
		PrivateKeyFile: tmpFilePrivate.Name(),
		PublicKeyFile:  tmpFilePublic.Name(),
		Name:           "my platform",
	}})
	assert.Equal(t, errors.New("invalid key: Key must be PEM encoded PKCS1 or PKCS8 private key"), err)
}

func Test_prepareFileName(t *testing.T) {
	assert.Equal(t, "/", prepareFileName("/"))

	preparedFileName := prepareFileName("")
	assert.True(t, strings.HasPrefix(preparedFileName, "/"))
	assert.Equal(t, "/app/token/../../", preparedFileName[len(preparedFileName)-len("/app/token/../../"):])

	preparedFileName = prepareFileName("some.file")
	assert.True(t, strings.HasPrefix(preparedFileName, "/"))
	assert.Equal(t, "/app/token/../../some.file",
		preparedFileName[len(preparedFileName)-len("/app/token/../../some.file"):])
}

func createTmpPublicKeyFile(key []byte) (*os.File, error) {
	tmpFilePublic, err := ioutil.TempFile("", "testPublicKey.pem")
	if err != nil {
		return tmpFilePublic, err
	}

	_, err = tmpFilePublic.Write(key)
	return tmpFilePublic, err
}

func createTmpPrivateKeyFile(key []byte) (*os.File, error) {
	tmpFilePrivate, err := ioutil.TempFile("", "testPrivateKey.pem")
	if err != nil {
		return tmpFilePrivate, err
	}

	_, err = tmpFilePrivate.Write(key)
	return tmpFilePrivate, err
}
