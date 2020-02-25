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
//   Lists the batches of users whose prefix can be used in the given group
//   (i.e., the `group_id` is a descendant of the prefix group).
//   Only those user batches are shown for which the authenticated user (or one of his group ancestors) is a manager of
//   the prefix group (or its ancestor) with at least 'can_manage:memberships'.
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

	managedByUser := srv.Store.ActiveGroupAncestors().ManagedByUser(user).
		Where("can_manage != 'none'").
		Select("groups_ancestors_active.child_group_id AS id")

	prefixAncestors := srv.Store.ActiveGroupAncestors().Where("child_group_id = ?", groupID).
		Select("ancestor_group_id AS id")

	query := srv.Store.UserBatches().
		Joins("JOIN user_batch_prefixes USING(group_prefix)").
		Where(`user_batch_prefixes.group_id IN(?)`, managedByUser.QueryExpr()).
		Where(`user_batch_prefixes.group_id IN(?)`, prefixAncestors.QueryExpr()).
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
