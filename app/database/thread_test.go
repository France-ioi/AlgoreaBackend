package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsThreadOpenStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"waiting_for_trainer", true},
		{"waiting_for_participant", true},
		{"", false},
		{"closed", false},
		{"not_started", false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			assert.Equal(t, tt.want, IsThreadOpenStatus(tt.status))
		})
	}
}

func Test_IsClosedStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"waiting_for_trainer", false},
		{"waiting_for_participant", false},
		{"", true},
		{"closed", true},
		{"not_started", true},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			isClosed := IsThreadClosedStatus(tt.status)
			assert.Equal(t, tt.want, isClosed)
			assert.Equal(t, isClosed, !IsThreadOpenStatus(tt.status))
		})
	}
}
