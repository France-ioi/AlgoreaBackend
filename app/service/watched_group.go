package service

import (
	"errors"
	"net/http"
)

// ResolveWatchedGroupID returns the watched_group_id parameter (if given) and checks if the current user has rights
// to watch for its members.
func (srv *Base) ResolveWatchedGroupID(httpReq *http.Request) (watchedGroupID int64, watchedGroupIDSet bool, apiError APIError) {
	if len(httpReq.URL.Query()["watched_group_id"]) == 0 {
		return 0, false, NoError
	}

	var err error
	watchedGroupID, err = ResolveURLQueryGetInt64Field(httpReq, "watched_group_id")
	if err != nil {
		return 0, false, ErrInvalidRequest(err)
	}
	var found bool
	found, err = srv.GetStore(httpReq).ActiveGroupAncestors().ManagedByUser(srv.GetUser(httpReq)).
		Where("groups_ancestors_active.child_group_id = ?", watchedGroupID).
		Where("can_watch_members").HasRows()
	if err != nil {
		return 0, false, ErrUnexpected(err)
	}
	if !found {
		return 0, false, ErrForbidden(errors.New("no rights to watch for watched_group_id"))
	}
	return watchedGroupID, true, NoError
}
