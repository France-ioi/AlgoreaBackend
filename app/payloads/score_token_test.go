package payloads

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreToken_Bind(t *testing.T) {
	tests := []struct {
		name          string
		taskToken     ScoreToken
		wantErr       error
		wantConverted ScoreTokenConverted
	}{
		{
			name:          "okay",
			taskToken:     ScoreToken{UserID: "10", UserAnswerID: "20", Score: "10.12"},
			wantConverted: ScoreTokenConverted{UserID: 10, UserAnswerID: 20, Score: 10.12},
		},
		{
			name:      "wrong idUser",
			taskToken: ScoreToken{UserID: "abc", UserAnswerID: "20", Score: "100"},
			wantErr:   errors.New("wrong idUser"),
		},
		{
			name:      "wrong idUserAnswer",
			taskToken: ScoreToken{UserID: "10", UserAnswerID: "abc", Score: "10.12"},
			wantErr:   errors.New("wrong idUserAnswer"),
		},
		{
			name:      "wrong score",
			taskToken: ScoreToken{UserID: "10", UserAnswerID: "20", Score: "abc"},
			wantErr:   errors.New("wrong score"),
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
