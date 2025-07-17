package payloads

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreToken_Bind(t *testing.T) {
	tests := []struct {
		name          string
		scoreToken    ScoreToken
		wantErr       error
		wantConverted ScoreTokenConverted
	}{
		{
			name: "okay",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "20",
				Score:        "10.12",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantConverted: ScoreTokenConverted{
				UserID:        10,
				UserAnswerID:  20,
				Score:         10.12,
				ParticipantID: 1,
				AttemptID:     2,
				LocalItemID:   42,
			},
		},
		{
			name: "wrong idUser",
			scoreToken: ScoreToken{
				UserID:       "abc",
				UserAnswerID: "20",
				Score:        "100",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUser"),
		},
		{
			name: "wrong idUserAnswer",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "abc",
				Score:        "10.12",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUserAnswer"),
		},
		{
			name: "wrong idUserAnswer",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "abc",
				Score:        "10.12",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUserAnswer"),
		},
		{
			name: "wrong score",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "20",
				Score:        "abc",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong score"),
		},
		{
			name: "wrong idAttempt",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "20",
				Score:        "10.12",
				AttemptID:    "abc",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idAttempt"),
		},
		{
			name: "wrong idItemLocal",
			scoreToken: ScoreToken{
				UserID:       "10",
				UserAnswerID: "20",
				Score:        "10.12",
				AttemptID:    "1/2",
				LocalItemID:  "abc",
			},
			wantErr: errors.New("wrong idItemLocal"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scoreToken.Bind()
			if tt.wantErr == nil {
				require.NoError(t, got)
				assert.Equal(t, tt.wantConverted, tt.scoreToken.Converted)
			} else {
				assert.Equal(t, tt.wantErr, got)
			}
		})
	}
}
