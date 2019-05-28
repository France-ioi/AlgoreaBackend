package payloads

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskToken_Bind(t *testing.T) {
	attemptID := "100"
	attemptIDInt64 := int64(100)
	attemptIDInt64Ptr := &attemptIDInt64
	wrongAttemptID := "abc"
	tests := []struct {
		name          string
		taskToken     TaskToken
		wantErr       error
		wantConverted TaskTokenConverted
	}{
		{
			name:          "okay",
			taskToken:     TaskToken{UserID: "10", LocalItemID: "20", AttemptID: &attemptID},
			wantConverted: TaskTokenConverted{UserID: 10, LocalItemID: 20, AttemptID: attemptIDInt64Ptr},
		},
		{
			name:      "wrong idUser",
			taskToken: TaskToken{UserID: "abc", LocalItemID: "20", AttemptID: &attemptID},
			wantErr:   errors.New("wrong idUser"),
		},
		{
			name:      "wrong idItemLocal",
			taskToken: TaskToken{UserID: "10", LocalItemID: "abc", AttemptID: &attemptID},
			wantErr:   errors.New("wrong idItemLocal"),
		},
		{
			name:      "wrong idAttempt",
			taskToken: TaskToken{UserID: "10", LocalItemID: "20", AttemptID: &wrongAttemptID},
			wantErr:   errors.New("wrong idAttempt"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.taskToken.Bind()
			if tt.wantErr == nil {
				assert.NoError(t, got)
				assert.Equal(t, tt.wantConverted, tt.taskToken.Converted)
			} else {
				assert.Equal(t, got, tt.wantErr)
			}
		})
	}
}

func TestTaskToken_MarshalJSON(t *testing.T) {
	tt := &TaskToken{
		UserID:          "10",
		AccessSolutions: ptrBool(true),
	}
	result, err := json.Marshal(ConvertIntoMap(tt))
	assert.NoError(t, err)
	assert.Equal(t, []byte(
		`{"bAccessSolutions":true,"date":"","idItemLocal":"","idUser":"10","itemUrl":"","platformName":"","randomSeed":""}`,
	), result)
}
