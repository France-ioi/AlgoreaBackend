package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase_ResolveWatchedGroupID(t *testing.T) {
	tests := []struct {
		name               string
		url                string
		wantWatchedGroupID int64
		wantOk             bool
		wantAPIError       APIError
	}{
		{name: "no watched_group_id", url: "/dummy", wantWatchedGroupID: 0, wantOk: false, wantAPIError: NoError},
		{
			name:               "invalid watched_group_id",
			url:                "/dummy?watched_group_id=abc",
			wantWatchedGroupID: 0,
			wantOk:             false,
			wantAPIError:       ErrInvalidRequest(errors.New("wrong value for watched_group_id (should be int64)")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, http.NoBody)
			watchedGroupID, ok, apiError := (&Base{}).ResolveWatchedGroupID(req)

			assert.Equal(t, tt.wantWatchedGroupID, watchedGroupID)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}
