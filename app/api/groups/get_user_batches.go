package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model userBatch
type userBatch struct {
	// required: true
	GroupPrefix string `json:"group_prefix"`
	// required: true
	CustomPrefix string `json:"custom_prefix"`
	// required: true
	Size int `json:"size"`
	// Nullable
	// required: true
	CreatorID *int64 `json:"creator_id"`
}

// swagger:operation GET /user-batches/by-group/{group_id} groups userBatchesView
// ---
// summary: List user batches
// description: >
//
//   Lists existing batches of users.
//
//
//   Restrictions:
//
//   * The authenticated user (or one of his group ancestors) should be a manager of the `group_id`
//     (directly, or of one of its ancestors)
//     with at least 'can_manage:memberships', otherwise the 'forbidden' response is returned.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: integer
// - name: sort
//   in: query
//   default: [group_prefix,custom_prefix]
//   type: array
//   items:
//     type: string
//     enum: [group_prefix,-group_prefix,custom_prefix,-custom_prefix,size,-size]
// - name: from.group_prefix
//   description: Start the page from the batch next to the batch with `user_batches.group_prefix` = `from.group_prefix`
//                (`from.custom_prefix` is required when `from.group_prefix` is given)
//   in: query
//   type: string
// - name: from.custom_prefix
//   description: Start the page from the batch next to the batch with `user_batches.custom_prefix` = `from.custom_prefix`
//                (`from.group_prefix` is required when `from.custom_prefix` is given)
//   in: query
//   type: string
// - name: from.size
//   description: Start the page from the batch next to the batch with `user_batches.size` = `from.size`
//                (`from.group_prefix` & `from.custom_prefix` are required when `from.size` is given)
//   in: query
//   type: string
// - name: limit
//   description: Display the first N members
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. The array of user batches
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/userBatch"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserBatches(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	found, err := srv.Store.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = ?", groupID).
		Where("can_manage != 'none'").HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	query := srv.Store.UserBatches().
		Joins("JOIN user_batch_prefixes USING(group_prefix)").
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.ancestor_group_id = ? AND
				   groups_ancestors_active.child_group_id = user_batch_prefixes.group_id`, groupID).
		Select("group_prefix, custom_prefix, size, creator_id")

	query, apiErr := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"group_prefix":  {ColumnName: "group_prefix", FieldType: "string"},
		"custom_prefix": {ColumnName: "custom_prefix", FieldType: "string"},
		"size":          {ColumnName: "size", FieldType: "int64"},
	}, "group_prefix,custom_prefix", []string{"group_prefix", "custom_prefix"}, false)
	if apiErr != service.NoError {
		return apiErr
	}

	var result []userBatch
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
