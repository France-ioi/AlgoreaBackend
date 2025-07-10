package items

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /items/{ids}/attempts items attemptCreate
//
//	---
//	summary: Create an attempt
//	description: >
//		Creates a new attempt for the given item with `creator_id` equal to `group_id` of the current user and make it
//		active for the user.
//		If `as_team_id` is given, the created attempt is linked to the `as_team_id` group instead of the user's self group.
//
//
//			Restrictions:
//
//		* if `as_team_id` is given, it should be a user's parent team group,
//		* the first item in `{ids}` should be a root activity/skill (groups.root_activity_id/root_skill_id)
//			of a group the participant is a descendant of or manages,
//		* `{ids}` should be an ordered list of parent-child items,
//		* the group creating the attempt should have at least 'content' access on each of the items in `{ids}`,
//		* the participant should have a started, allowing submission, not ended result for each item but the last,
//			with `{parent_attempt_id}` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//		* if `{ids}` consists of only one item, the `{parent_attempt_id}` should be zero,
//		* the final item in `{ids}` should be either 'Task', or 'Chapter',
//
//		otherwise the 'forbidden' error is returned.
//
//
//		If there is already an attempt for the (item, group) pair, `items.allows_multiple_attempts` should be true, otherwise
//		the "unprocessable entity" error is returned.
//	parameters:
//		- name: ids
//			in: path
//			type: string
//			description: slash-separated list of item IDs (no more than 10 IDs)
//			required: true
//		- name: parent_attempt_id
//			in: query
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"201":
//			"$ref": "#/responses/createdWithIDResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"422":
//			"$ref": "#/responses/unprocessableEntityResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createAttempt(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error

	ids, err := idsFromRequest(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	parentAttemptID, err := service.ResolveURLQueryGetInt64Field(httpRequest, "parent_attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	participantID := service.ParticipantIDFromContext(httpRequest.Context())

	var attemptID int64
	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		var ok bool
		ok, err = store.Items().IsValidParticipationHierarchyForParentAttempt(ids, participantID, parentAttemptID, true, true)
		service.MustNotBeError(err)
		if !ok {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		itemID := ids[len(ids)-1]
		err = checkIfAttemptCreationIsPossible(store, itemID, participantID)
		service.MustNotBeError(err)

		attemptID, err = store.Attempts().CreateNew(participantID, parentAttemptID, itemID, user.GroupID)
		service.MustNotBeError(err)

		store.ScheduleResultsPropagation()
		return nil
	})
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, service.CreationSuccess(map[string]interface{}{
		"id": strconv.FormatInt(attemptID, 10),
	}))
	return nil
}

func checkIfAttemptCreationIsPossible(store *database.DataStore, itemID, groupID int64) error {
	var allowsMultipleAttempts bool
	err := store.Items().ByID(itemID).
		Where("items.type IN('Task','Chapter')").
		PluckFirst("items.allows_multiple_attempts", &allowsMultipleAttempts).WithExclusiveWriteLock().Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	if !allowsMultipleAttempts {
		var found bool
		found, err = store.Results().
			Where("participant_id = ?", groupID).Where("item_id = ?", itemID).WithExclusiveWriteLock().HasRows()
		service.MustNotBeError(err)
		if found {
			return service.ErrUnprocessableEntity(errors.New("the item doesn't allow multiple attempts"))
		}
	}
	return nil
}
