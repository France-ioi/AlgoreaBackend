package contests

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation PUT /contests/{item_id}/groups/{group_id}/additional-times contests groups contestSetAdditionalTime
// ---
// summary: Set additional time in the contest for the group
// description: >
//                For the input group and item, sets the `groups_items.additional_time` to the `time` value.
//                If there is no `groups_items` for the given `group_id`, `item_id` and the `seconds` != 0, creates it
//                (with default values in other columns).
//                If no `groups_items` and `seconds` == 0, succeed without doing any change.
//
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * the authenticated user should have `solutions` or `full` access on the input item;
//                  * the authenticated user should own the `group_id`;
//                  * if the contest is team-only (`items.has_attempts` = 1), then the group should not be a user group.
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a timed contest"
//   in: path
//   type: integer
//   required: true
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: seconds
//   description: additional time in seconds (can be negative)
//   in: query
//   type: integer
//   minimum: -3020399
//   maximum: 3020399
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) setAdditionalTime(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	seconds, err := service.ResolveURLQueryGetInt64Field(r, "seconds")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	const maxSeconds = 838*3600 + 59*60 + 59 // 838:59:59 is the maximum possible TIME value in MySQL
	if seconds < -maxSeconds || maxSeconds < seconds {
		return service.ErrInvalidRequest(fmt.Errorf("'seconds' should be between %d and %d", -maxSeconds, maxSeconds))
	}

	var groupType string
	err = srv.Store.Groups().OwnedBy(user).Where("groups.id = ?", groupID).
		PluckFirst("groups.type", &groupType).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	isTeamOnly, err := srv.getTeamModeForTimedContestManagedByUser(itemID, user)
	if gorm.IsRecordNotFoundError(err) || (isTeamOnly && groupType == "UserSelf") {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	srv.setAdditionalTimeForGroupInContest(groupID, itemID, seconds, user.ID)

	render.Respond(w, r, service.UpdateSuccess(nil))
	return service.NoError
}

func (srv *Service) setAdditionalTimeForGroupInContest(groupID, itemID, seconds, creatorUserID int64) {
	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		groupItemStore := store.GroupItems()
		scope := groupItemStore.Where("group_id = ?", groupID).Where("item_id = ?", itemID)
		found, err := scope.WithWriteLock().HasRows()
		service.MustNotBeError(err)
		if found {
			service.MustNotBeError(scope.UpdateColumn("additional_time", gorm.Expr("SEC_TO_TIME(?)", seconds)).Error())
		} else if seconds != 0 {
			service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
				id := retryStore.NewID()
				return retryStore.Exec(
					"INSERT INTO groups_items (id, group_id, item_id, additional_time, user_created_id) VALUES(?, ?, ?, SEC_TO_TIME(?), ?)",
					id, groupID, itemID, seconds, creatorUserID).Error()
			}))
		}
		return nil
	}))
}
