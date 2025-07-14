package payloads

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

func TestTaskToken_Bind(t *testing.T) {
	tests := []struct {
		name          string
		taskToken     TaskToken
		wantErr       error
		wantConverted TaskTokenConverted
	}{
		{
			name:          "okay",
			taskToken:     TaskToken{UserID: "10", LocalItemID: "20", AttemptID: "100/1"},
			wantConverted: TaskTokenConverted{UserID: 10, LocalItemID: 20, ParticipantID: int64(100), AttemptID: int64(1)},
		},
		{
			name:      "wrong idUser",
			taskToken: TaskToken{UserID: "abc", LocalItemID: "20", AttemptID: "100"},
			wantErr:   errors.New("wrong idUser"),
		},
		{
			name:      "wrong idItemLocal",
			taskToken: TaskToken{UserID: "10", LocalItemID: "abc", AttemptID: "100"},
			wantErr:   errors.New("wrong idItemLocal"),
		},
		{
			name:      "wrong idAttempt",
			taskToken: TaskToken{UserID: "10", LocalItemID: "20", AttemptID: "abc"},
			wantErr:   errors.New("wrong idAttempt"),
		},
		{
			name:      "wrong idAttempt (only one number)",
			taskToken: TaskToken{UserID: "10", LocalItemID: "20", AttemptID: "123"},
			wantErr:   errors.New("wrong idAttempt"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.taskToken.Bind()
			if tt.wantErr == nil {
				require.NoError(t, got)
				assert.Equal(t, tt.wantConverted, tt.taskToken.Converted)
			} else {
				assert.Equal(t, tt.wantErr, got)
			}
		})
	}
}

func TestTaskToken_MarshalJSON(t *testing.T) {
	tt := &TaskToken{
		UserID:          "10",
		AttemptID:       "200",
		AccessSolutions: golang.Ptr(true),
	}
	result, err := json.Marshal(ConvertIntoMap(tt))
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"bAccessSolutions":true,"date":"","idAttempt":"200","idItemLocal":"","idUser":"10","itemUrl":"","platformName":"","randomSeed":""}`,
		string(result))
}
