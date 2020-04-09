package items

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /items/{item_id}/attempts items attemptCreate
// ---
// summary: Create an attempt
// description: >
//   Creates a new attempt for the given item with `creator_id` equal to `group_id` of the current user and make it
//   active for the user.
//   If `as_team_id` is given, the created attempt is linked to the `as_team_id` group instead of the user's self group.
//
//
//   Restrictions:
//
//     * if `as_team_id` is given, it should be a user's parent team group,
//     * the group creating the attempt should have at least 'content' access to the item,
//     * the item should be either 'Task', 'Course', or 'Chapter',
//
//   otherwise the 'forbidden' error is returned.
//
//
//   If there is already an attempt for the (item, group) pair, `items.allows_multiple_attempts` should be true, otherwise
//   the "unprocessable entity" error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
// responses:
//   "201":
//     "$ref": "#/responses/createdWithIDResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
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

	var allowsMultipleAttempts bool
	err = srv.Store.Items().ByID(itemID).WhereGroupHasViewPermissionOnItems(groupID, "content").
		Where("items.type IN('Task','Course','Chapter')").
		PluckFirst("items.allows_multiple_attempts", &allowsMultipleAttempts).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var attemptID int64
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		if !allowsMultipleAttempts {
			var found bool
			found, err = store.Results().
				Where("participant_id = ?", groupID).Where("item_id = ?", itemID).WithWriteLock().HasRows()
			service.MustNotBeError(err)
			if found {
				apiError = service.ErrUnprocessableEntity(errors.New("the item doesn't allow multiple attempts"))
				return apiError.Error // rollback
			}
		}

		attemptID, err = store.Attempts().CreateNew(groupID, itemID, user.GroupID)
		service.MustNotBeError(err)

		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, map[string]interface{}{
		"id": strconv.FormatInt(attemptID, 10),
	})
	return service.NoError
}
