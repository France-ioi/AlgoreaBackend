package service

import (
	"errors"
	"net/http"
)

// ResolveWatchedGroupID returns the watched_group_id parameter (if given) and checks if the current user has rights
// to watch for its members.
func (srv *Base) ResolveWatchedGroupID(httpReq *http.Request) (watchedGroupID int64, ok bool, apiError APIError) {
	if len(httpReq.URL.Query()["watched_group_id"]) == 0 {
		return 0, false, NoError
	}

	var err error
	watchedGroupID, err = ResolveURLQueryGetInt64Field(httpReq, "watched_group_id")
	if err != nil {
		return 0, false, ErrInvalidRequest(err)
	}

	found := srv.GetUser(httpReq).CanWatchGroupMembers(srv.GetStore(httpReq), watchedGroupID)
	if !found {
		return 0, false, ErrForbidden(errors.New("no rights to watch for watched_group_id"))
	}
	return watchedGroupID, true, NoError
}
