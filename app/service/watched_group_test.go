package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase_ResolveWatchedGroupID(t *testing.T) {
	tests := []struct {
		name                    string
		url                     string
		wantWatchedGroupID      int64
		wantWatchedGroupIDIsSet bool
		wantError               error
	}{
		{name: "no watched_group_id", url: "/dummy", wantWatchedGroupID: 0, wantWatchedGroupIDIsSet: false},
		{
			name:                    "invalid watched_group_id",
			url:                     "/dummy?watched_group_id=abc",
			wantWatchedGroupID:      0,
			wantWatchedGroupIDIsSet: false,
			wantError:               ErrInvalidRequest(errors.New("wrong value for watched_group_id (should be int64)")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.url, http.NoBody)
			watchedGroupID, watchedGroupIDIsSet, gotError := (&Base{}).ResolveWatchedGroupID(req)

			assert.Equal(t, tt.wantWatchedGroupID, watchedGroupID)
			assert.Equal(t, tt.wantWatchedGroupIDIsSet, watchedGroupIDIsSet)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}
