package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /attempts/{attempt_id} items itemAttemptEnd
// ---
// summary: End an attempt
// description: >
//                Allows to end an attempt as a user or as a team (if `as_team_id` is given).
//
//
//                Restrictions:
//                  * `as_team_id` (if given) should be the current user's team;
//                  * an attempt with `participant_id` = `as_team_id` (or the current user) and `id` = `attempt_id`
//                    should exist and not be ended or expired;
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: attempt_id
//   description: "`id` of an attempt to end"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
//   format: int64
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) endAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	attemptID, err := service.ResolveURLQueryPathInt64Field(r, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	groupID := user.GroupID
	if len(r.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "as_team_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		var found bool
		found, err = srv.Store.Groups().TeamGroupForUser(groupID, user).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}

	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var found bool
		found, err = store.Attempts().
			Where("participant_id = ?", groupID).
			Where("id = ?", attemptID).
			Where("allows_submissions_until > NOW()").
			Where("ended_at IS NULL").
			WithWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.ErrForbidden(errors.New("active attempt not found"))
			return apiError.Error // rollback
		}

		// End this and descendant attempts, expire participations
		service.MustNotBeError(store.Exec(`
			WITH RECURSIVE attempts_to_update(id) AS (
				SELECT id FROM attempts where participant_id = ? and id = ?
				UNION
				SELECT attempts.id FROM attempts
					JOIN attempts_to_update ON attempts_to_update.id = attempts.parent_attempt_id
					WHERE attempts.participant_id = ?)
			UPDATE attempts
			LEFT JOIN items ON items.id = attempts.root_item_id
			LEFT JOIN groups_groups ON groups_groups.parent_group_id = items.participants_group_id AND
				groups_groups.child_group_id = attempts.participant_id
			SET
				attempts.ended_at = IFNULL(LEAST(NOW(), attempts.ended_at), NOW()),
				attempts.allows_submissions_until = LEAST(NOW(), attempts.allows_submissions_until),
				groups_groups.expires_at = LEAST(NOW(), groups_groups.expires_at)
			WHERE attempts.participant_id = ? AND attempts.id IN(SELECT id FROM attempts_to_update)`,
			groupID, attemptID, groupID, groupID).
			Error())

		return store.GroupGroups().After()
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
