package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

type participantMiddlewareKey int

const ctxParticipant participantMiddlewareKey = iota

// GetStorer is an interface allowing to get a data store bound to the context of the given request.
type GetStorer interface {
	GetStore(r *http.Request) *database.DataStore
}

// ParticipantMiddleware is a middleware retrieving a participant from the request content.
// The participant id is the `as_team_id` parameter value if it is given or the user's `group_id` otherwise.
// If `as_team_id` is given, it should be an id of a team and the user should be a member of this team, otherwise
// the 'forbidden' error is returned.
func ParticipantMiddleware(srv GetStorer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			var participantID int64
			var failed bool
			AppHandler(func(_ http.ResponseWriter, httpRequest *http.Request) error {
				var err error
				defer func() {
					failed = err != nil
					if p := recover(); p != nil {
						failed = true
						panic(p)
					}
				}()
				user := auth.UserFromContext(httpRequest.Context())
				participantID, err = GetParticipantIDFromRequest(httpRequest, user, srv.GetStore(httpRequest))
				return err
			}).ServeHTTP(responseWriter, httpRequest)
			if failed {
				return
			}

			ctx := context.WithValue(httpRequest.Context(), ctxParticipant, participantID)
			next.ServeHTTP(responseWriter, httpRequest.WithContext(ctx))
		})
	}
}

// ParticipantIDFromContext retrieves a participant id  set by the middleware from a context.
func ParticipantIDFromContext(ctx context.Context) int64 {
	return ctx.Value(ctxParticipant).(int64) //nolint:forcetypeassert // panic if the value is of wrong type
}

// GetParticipantIDFromRequest returns `as_team_id` parameter value if it is given or the user's `group_id` otherwise.
// If `as_team_id` is given, it should be an id of a team and the user should be a member of this team, otherwise
// the 'forbidden' error is returned.
func GetParticipantIDFromRequest(httpReq *http.Request, user *database.User, store *database.DataStore) (int64, error) {
	groupID := user.GroupID
	var err error
	if len(httpReq.URL.Query()["as_team_id"]) != 0 {
		groupID, err = ResolveURLQueryGetInt64Field(httpReq, "as_team_id")
		if err != nil {
			return 0, ErrInvalidRequest(err)
		}

		var found bool
		found, err = store.Groups().TeamGroupForUser(groupID, user).HasRows()
		MustNotBeError(err)
		if !found {
			return 0, ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}
	return groupID, nil
}
