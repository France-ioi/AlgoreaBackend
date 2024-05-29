package payloads

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnswerToken_Bind(t *testing.T) {
	tests := []struct {
		name          string
		answerToken   AnswerToken
		wantErr       error
		wantConverted AnswerTokenConverted
	}{
		{
			name: "okay",
			answerToken: AnswerToken{
				UserID:       "10",
				UserAnswerID: "20",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantConverted: AnswerTokenConverted{
				UserID:        10,
				UserAnswerID:  20,
				ParticipantID: 1,
				AttemptID:     2,
				LocalItemID:   42,
			},
		},
		{
			name: "wrong idUser",
			answerToken: AnswerToken{
				UserID:       "abc",
				UserAnswerID: "20",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUser"),
		},
		{
			name: "wrong idUserAnswer",
			answerToken: AnswerToken{
				UserID:       "10",
				UserAnswerID: "abc",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUserAnswer"),
		},
		{
			name: "wrong idUserAnswer",
			answerToken: AnswerToken{
				UserID:       "10",
				UserAnswerID: "abc",
				AttemptID:    "1/2",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idUserAnswer"),
		},
		{
			name: "wrong idAttempt",
			answerToken: AnswerToken{
				UserID:       "10",
				UserAnswerID: "20",
				AttemptID:    "abc",
				LocalItemID:  "42",
			},
			wantErr: errors.New("wrong idAttempt"),
		},
		{
			name: "wrong idItemLocal",
			answerToken: AnswerToken{
				UserID:       "10",
				UserAnswerID: "20",
				AttemptID:    "1/2",
				LocalItemID:  "abc",
			},
			wantErr: errors.New("wrong idItemLocal"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.answerToken.Bind()
			if tt.wantErr == nil {
				assert.NoError(t, got)
				assert.Equal(t, tt.wantConverted, tt.answerToken.Converted)
			} else {
				assert.Equal(t, got, tt.wantErr)
			}
		})
	}
}
