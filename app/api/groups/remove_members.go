package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /groups/{group_id}/members groups users groupRemoveMembers
// ---
// summary: Remove members from a group
// description:
//   Lets an admin remove users from a group.
//   On success the service sets `groups_groups.type` to "removed" and `status_date` to current UTC time
//   for each self group of `user_ids`. It also refreshes the access rights.
//
//
//   The authenticated user should be an owner of the `group_id`, otherwise the 'forbidden' error is returned.
//
//
//   Each of the input `user_ids` should have the input `group_id` as a parent of their self group and the
//   `groups_groups.type` should be one of "invitationAccepted"/"requestAccepted"/"joinedByCode",
//   otherwise the `user_id` gets skipped with `unchanged` (if `type` = "removed") or `invalid` as the result.
//   If a user is not found or doesn't have a self group, it gets skipped with `not_found` as the result.
//
//
//   The response status code on success (200) doesn't depend on per-group results.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: user_ids
//   in: query
//   type: array
//   items:
//     type: integer
//     format: int64
//   required: true
// responses:
//   "200":
//     description: OK. Success response with the per-user deletion statuses
//     schema:
//       type: object
//       required: [message, success, data]
//       properties:
//         message:
//           type: string
//           description: success
//           enum: [success]
//         success:
//           type: string
//           description: "true"
//         data:
//           description: "`user_id` -> `result`"
//           type: object
//           additionalProperties:
//             type: string
//             enum: [invalid, success, unchanged, not_found]
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) removeMembers(w http.ResponseWriter, r *http.Request) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	userIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "user_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	if apiErr := checkThatUserOwnsTheGroup(srv.Store, user, parentGroupID); apiErr != service.NoError {
		return apiErr
	}

	results := make(database.GroupGroupTransitionResults, len(userIDs))
	for _, userID := range userIDs {
		results[userID] = notFound
	}

	var groupsToRemoveRows []struct {
		UserID      int64
		SelfGroupID int64
	}
	service.MustNotBeError(srv.Store.Users().Select("id AS user_id, self_group_id").
		Where("id IN (?)", userIDs).Where("self_group_id IS NOT NULL").
		Scan(&groupsToRemoveRows).Error())

	groupsToRemove := make([]int64, 0, len(groupsToRemoveRows))
	groupToUserMap := make(map[int64]int64, len(groupsToRemoveRows))
	for _, row := range groupsToRemoveRows {
		groupsToRemove = append(groupsToRemove, row.SelfGroupID)
		groupToUserMap[row.SelfGroupID] = row.UserID
	}

	var groupResults database.GroupGroupTransitionResults
	if len(groupsToRemove) > 0 {
		err = srv.Store.InTransaction(func(store *database.DataStore) error {
			groupResults, err = store.GroupGroups().Transition(database.AdminRemovesUser, parentGroupID, groupsToRemove, user.ID)
			return err
		})
	}

	service.MustNotBeError(err)
	for id, result := range groupResults {
		results[groupToUserMap[id]] = result
	}

	response := service.Response{
		Success: true,
		Message: "deleted",
		Data:    results,
	}
	render.Respond(w, r, &response)
	return service.NoError
}
