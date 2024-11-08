//go:build !unit

package token_test

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestToken_UnmarshalDependingOnItemPlatform(t *testing.T) {
	expectedParsedPayload := payloads.HintToken{}
	_ = payloads.ParseMap(payloadstest.HintPayloadFromTaskPlatform, &expectedParsedPayload)
	expectedToken := (*token.Hint)(&expectedParsedPayload)
	tests := []struct {
		name                   string
		itemID                 int64
		token                  []byte
		tokenFieldName         string
		fixtures               []string
		target                 interface{}
		expected               interface{}
		expectedHasPlatformKey bool
		expectedErr            error
	}{
		{
			name:                   "no platform",
			itemID:                 404,
			token:                  []byte(""),
			expectedHasPlatformKey: false,
			expectedErr:            errors.New("cannot find the platform for item 404"),
		},
		{
			name:   "missing token",
			itemID: 50,
			fixtures: []string{
				`platforms: [{id: 11, regexp: "http://taskplatform1.mblockelet.info/task.html.*",
						public_key: ` + fmt.Sprintf("%q", tokentest.TaskPlatformPublicKey) + `}]`,
				`items: [{id: 50, platform_id: 11, url: "http://taskplatform1.mblockelet.info/task.html?taskId=403449543672183936",
				          default_language_tag: fr}]`,
			},
			token:                  nil,
			tokenFieldName:         "hint_requested",
			target:                 reflect.New(reflect.TypeOf(&token.Hint{})).Interface(),
			expected:               nil,
			expectedHasPlatformKey: true,
			expectedErr:            errors.New("missing hint_requested"),
		},
		{
			name:   "invalid token",
			itemID: 50,
			fixtures: []string{
				`platforms: [{id: 10, regexp: "http://taskplatform2.mblockelet.info/task.html\\.*",
						public_key: ` + fmt.Sprintf("%q", tokentest.TaskPlatformPublicKey) + `}]`,
				`items: [{id: 50, platform_id: 10, url: "http://taskplatform2.mblockelet.info/task.html?taskId=403449543672183936",
				          default_language_tag: fr}]`,
			},
			token:                  []byte(""),
			tokenFieldName:         "hint_requested",
			target:                 reflect.New(reflect.TypeOf(&token.Hint{})).Interface(),
			expected:               nil,
			expectedHasPlatformKey: true,
			expectedErr:            errors.New("invalid hint_requested: unexpected end of JSON input"),
		},
		{
			name:   "invalid public key",
			itemID: 50,
			fixtures: []string{
				`platforms: [{id: 10, regexp: "^http://taskplatform3\\.mblockelet\\.info/task\\.html\\.*",
						public_key: dasdfa}]`,
				`items: [{id: 50, platform_id: 10, url: "http://taskplatform3.mblockelet.info/task.html?taskId=403449543672183936",
				          default_language_tag: fr}]`,
			},
			token:                  []byte("dsafafd"),
			tokenFieldName:         "score_token",
			target:                 reflect.New(reflect.TypeOf(&token.Hint{})).Interface(),
			expected:               nil,
			expectedHasPlatformKey: true,
			expectedErr:            errors.New("invalid score_token: wrong platform's key"),
		},
		{
			name:   "everything is okay",
			itemID: 50,
			fixtures: []string{
				`platforms: [{id: 10, regexp: "^http://taskplatform4\\.mblockelet.info/task.html.*$",
						public_key: ` + fmt.Sprintf("%q", tokentest.TaskPlatformPublicKey) + `}]`,
				`items: [{id: 50, platform_id: 10, url: "http://taskplatform4.mblockelet.info/task.html?taskId=403449543672183936",
				          default_language_tag: fr}]`,
			},
			token: []byte(fmt.Sprintf("%q", token.Generate(payloadstest.HintPayloadFromTaskPlatform,
				tokentest.TaskPlatformPrivateKeyParsed))),
			tokenFieldName:         "hint_requested",
			target:                 reflect.New(reflect.TypeOf(&token.Hint{})).Interface(),
			expected:               expectedToken,
			expectedHasPlatformKey: true,
			expectedErr:            nil,
		},
		{
			name:   "platform doesn't use tokens",
			itemID: 50,
			fixtures: []string{
				`platforms: [{id: 10, regexp: "^http://taskplatform5\\.mblockelet\\.info/task.html.*$"}]`,
				`items: [{id: 50, platform_id: 10, url: "http://taskplatform5.mblockelet.info/task.html?taskId=403449543672183936",
				          default_language_tag: fr}]`,
			},
			token:                  []byte(`{}`),
			tokenFieldName:         "hint_requested",
			target:                 reflect.New(reflect.TypeOf(&token.Hint{})).Interface(),
			expected:               (*token.Hint)(nil),
			expectedHasPlatformKey: false,
			expectedErr:            nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(tt.fixtures...)
			defer func() { _ = db.Close() }()
			store := database.NewDataStore(db)
			hasPlatformKey, err := token.UnmarshalDependingOnItemPlatform(store, tt.itemID, tt.target, tt.token, tt.tokenFieldName)
			assert.Equal(t, tt.expectedHasPlatformKey, hasPlatformKey)
			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				if tt.expected != nil && tt.target != nil && !reflect.ValueOf(tt.target).Elem().IsNil() {
					reflect.ValueOf(tt.expected).Elem().FieldByName("Date").Set(
						reflect.ValueOf(tt.target).Elem().Elem().FieldByName("Date"))
					reflect.ValueOf(tt.target).Elem().Elem().FieldByName("PublicKey").Set(reflect.ValueOf((*rsa.PublicKey)(nil)))
				}
				assert.Equal(t, tt.expected, reflect.ValueOf(tt.target).Elem().Interface())
			}
		})
	}
}
