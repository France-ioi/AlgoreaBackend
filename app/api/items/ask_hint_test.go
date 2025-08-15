package items

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestAskHintRequest_UnmarshalJSON(t *testing.T) {
	expectedTaskToken := token.Token[payloads.TaskToken]{}
	_ = payloads.ParseMap(payloadstest.TaskPayloadFromAlgoreaPlatform(), &expectedTaskToken.Payload)
	expectedTaskToken.Payload.Converted.UserID = 556371821693219925
	expectedTaskToken.Payload.Converted.LocalItemID = 901756573345831409
	expectedTaskToken.Payload.Converted.AttemptID = 100
	expectedTaskToken.Payload.Converted.ParticipantID = 556371821693219925
	expectedHintToken := token.Token[payloads.HintToken]{}
	expectedHintToken.Payload.Converted.UserID = 556371821693219925
	_ = payloads.ParseMap(payloadstest.HintPayloadFromTaskPlatform(), &expectedHintToken.Payload)
	monkey.Patch(time.Now,
		func() time.Time { return time.Date(2019, 5, 2, 12, 0, 0, 0, time.UTC) })
	defer monkey.UnpatchAll()

	type platform struct {
		publicKey string
	}
	tests := []struct {
		name     string
		platform *platform
		raw      []byte
		expected AskHintRequest
		wantErr  error
		mockDB   bool
		itemID   int64
	}{
		{
			name:    "invalid json",
			raw:     []byte(""),
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name:    "missing task_token",
			raw:     []byte(`{}`),
			wantErr: errors.New("missing task_token"),
		},
		{
			name:    "invalid task_token",
			raw:     []byte(`{"task_token":"ABC.DEF.ABC"}`),
			wantErr: errors.New("invalid task_token: invalid character '\\x00' looking for beginning of value"),
		},
		{
			name: "missing hint_requested token",
			raw: []byte(fmt.Sprintf(`{"task_token": %q}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()))),
			wantErr: errors.New("missing hint_requested"),
		},
		{
			name: "missing platform",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": "ABC.DEF.ABC"}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()))),
			wantErr: errors.New("cannot find the platform for item 901756573345831409"),
			mockDB:  true,
			itemID:  901756573345831409,
		},
		{
			name: "wrong platform's public key",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": "ABC.DEF.ABC"}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()))),
			wantErr:  errors.New("invalid hint_requested: wrong platform's key"),
			mockDB:   true,
			itemID:   901756573345831409,
			platform: &platform{publicKey: "zzz"},
		},
		{
			name: "hint_requested is not a string, but it should be a token",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": 1234}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()))),
			wantErr:  errors.New("invalid hint_requested: json: cannot unmarshal number into Go value of type string"),
			mockDB:   true,
			itemID:   901756573345831409,
			platform: &platform{publicKey: tokentest.AlgoreaPlatformPublicKey},
		},
		{
			name: "invalid hint_requested token",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": "ABC.DEF.ABC"}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()))),
			wantErr:  errors.New("invalid hint_requested: invalid character '\\x00' looking for beginning of value"),
			mockDB:   true,
			platform: &platform{publicKey: tokentest.AlgoreaPlatformPublicKey},
			itemID:   901756573345831409,
		},
		{
			name: "everything is okay",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": %q}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()),
				token.Generate(payloadstest.HintPayloadFromTaskPlatform(), tokentest.TaskPlatformPrivateKeyParsed()),
			)),
			mockDB:   true,
			platform: &platform{publicKey: tokentest.TaskPlatformPublicKey},
			itemID:   901756573345831409,
			expected: AskHintRequest{
				TaskToken: &expectedTaskToken,
				HintToken: &expectedHintToken,
			},
		},
		{
			name: "plain hint_requested should not be accepted",
			raw: []byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": {"idUser":"556371821693219925","askedHint":"123"}}`,
				token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()),
			)),
			mockDB:   true,
			itemID:   901756573345831409,
			platform: &platform{publicKey: tokentest.AlgoreaPlatformPublicKey},
			wantErr:  errors.New("invalid hint_requested: json: cannot unmarshal object into Go value of type string"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := database.NewDBMock()
			defer func() { _ = db.Close() }()

			if tt.mockDB {
				var publicKeyPtr *string
				if tt.platform != nil {
					publicKeyPtr = &tt.platform.publicKey
				}
				mockPlatformPublicKeyLoading(mock, tt.itemID, publicKeyPtr)
			}

			askHintRequest := &AskHintRequest{
				store:     database.NewDataStore(db),
				publicKey: tokentest.AlgoreaPlatformPublicKeyParsed(),
			}
			err := askHintRequest.UnmarshalJSON(tt.raw)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				if err == nil {
					assert.Equal(t, tt.wantErr, err)
				} else {
					assert.Equal(t, tt.wantErr.Error(), err.Error())
				}
			}
			if err == nil {
				askHintRequest.store = nil
				if askHintRequest.HintToken != nil {
					askHintRequest.HintToken.PublicKey = nil
				}
				if askHintRequest.TaskToken != nil {
					askHintRequest.TaskToken.PublicKey = nil
				}
				askHintRequest.publicKey = nil
				assert.Equal(t, &tt.expected, askHintRequest)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func mockPlatformPublicKeyLoading(mock sqlmock.Sqlmock, itemID int64, publicKey *string) {
	mockQuery := mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT public_key "+
		"FROM `platforms` JOIN items ON items.platform_id = platforms.id WHERE (items.id = ?) LIMIT 1") + "$").
		WithArgs(itemID)

	if publicKey != nil {
		if *publicKey == "" {
			publicKey = nil
		}
		mockQuery.
			WillReturnRows(mock.NewRows([]string{"public_key"}).AddRow(publicKey))
	} else {
		mockQuery.
			WillReturnRows(mock.NewRows([]string{"public_key"}))
	}
}

func TestAskHintRequest_UnmarshalJSON_DBError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("error")
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT public_key "+
		"FROM `platforms` JOIN items ON items.platform_id = platforms.id WHERE (items.id = ?) LIMIT 1") + "$").
		WithArgs(901756573345831409).WillReturnError(expectedError)

	r := &AskHintRequest{
		store:     database.NewDataStore(db),
		publicKey: tokentest.AlgoreaPlatformPublicKeyParsed(),
	}
	assert.PanicsWithError(t, expectedError.Error(), func() {
		_ = r.UnmarshalJSON([]byte(fmt.Sprintf(`{"task_token": %q, "hint_requested": %q}`,
			token.Generate(payloadstest.TaskPayloadFromAlgoreaPlatform(), tokentest.AlgoreaPlatformPrivateKeyParsed()),
			token.Generate(payloadstest.HintPayloadFromTaskPlatform(), tokentest.TaskPlatformPrivateKeyParsed()),
		)))
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}
