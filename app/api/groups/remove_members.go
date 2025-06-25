package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /groups/{group_id}/members group-memberships groupRemoveMembers
//
//	---
//	summary: Remove members from a group
//	description: >
//
//		Lets an admin remove users from a group.
//		On success the service removes relations from `groups_groups` and creates `group_membership_changes` rows
//		with `action` = 'removed and `at` = current UTC time
//		for each of `user_ids`. It also refreshes the access rights.
//
//
//		The authenticated user should be a manager of the `group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//
//
//		Each of the input `user_ids` should have the input `group_id` as a parent in `groups_groups`,
//		otherwise the `user_id` gets skipped with `invalid` as the result.
//
//
//		The response status code on success (200) doesn't depend on per-group results.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: user_ids
//			in: query
//			type: array
//			items:
//				type: integer
//				format: int64
//			required: true
//	responses:
//		"200":
//			description: OK. Success response with the per-user deletion statuses
//			schema:
//					type: object
//					required: [message, success, data]
//					properties:
//						message:
//							type: string
//							description: success
//							enum: [success]
//						success:
//							type: string
//							description: "true"
//						data:
//							description: "`user_id` -> `result`"
//							type: object
//							additionalProperties:
//								type: string
//								enum: [invalid, success, unchanged, not_found]
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) removeMembers(w http.ResponseWriter, r *http.Request) error {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	userIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "user_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)
	service.MustNotBeError(checkThatUserCanManageTheGroupMemberships(store, user, parentGroupID))

	results := make(database.GroupGroupTransitionResults, len(userIDs))
	for _, userID := range userIDs {
		results[userID] = notFound
	}

	var groupsToRemove []int64
	service.MustNotBeError(store.Users().Select("group_id").
		Where("group_id IN (?)", userIDs).Pluck("group_id", &groupsToRemove).Error())

	var groupResults database.GroupGroupTransitionResults
	if len(groupsToRemove) > 0 {
		err = store.InTransaction(func(store *database.DataStore) error {
			groupResults, _, err = store.GroupGroups().
				Transition(database.AdminRemovesUser, parentGroupID, groupsToRemove, nil, user.GroupID)
			return err
		})
	}

	service.MustNotBeError(err)
	for id, result := range groupResults {
		results[id] = result
	}

	response := service.Response[database.GroupGroupTransitionResults]{
		Success: true,
		Message: "deleted",
		Data:    results,
	}
	render.Respond(w, r, &response)
	return nil
}
