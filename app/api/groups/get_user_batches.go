package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model userBatch
type userBatch struct {
	// required: true
	GroupPrefix string `json:"group_prefix"`
	// required: true
	CustomPrefix string `json:"custom_prefix"`
	// required: true
	Size int `json:"size"`
	// required: true
	CreatorID *int64 `json:"creator_id,string"`
}

// swagger:operation GET /user-batches/by-group/{group_id} groups userBatchesView
//
//	---
//	summary: List user batches
//	description: >
//
//		Lists the batches of users whose prefix can be used in the given group
//		(i.e., the `group_id` is a descendant of the prefix group).
//		Only those user batches are shown for which the authenticated user (or one of his group ancestors) is a manager of
//		the prefix group (or its ancestor) with at least 'can_manage:memberships'.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [group_prefix,custom_prefix]
//			type: array
//			items:
//				type: string
//				enum: [group_prefix,-group_prefix,custom_prefix,-custom_prefix,size,-size]
//		- name: from.group_prefix
//			description: Start the page from the batch next to the batch with `user_batches.group_prefix` = `{from.group_prefix}`
//							 (`{from.custom_prefix}` is required when `{from.group_prefix}` is given)
//			in: query
//			type: string
//		- name: from.custom_prefix
//			description: Start the page from the batch next to the batch with `user_batches.custom_prefix` = `{from.custom_prefix}`
//							 (`{from.group_prefix}` is required when `{from.custom_prefix}` is given)
//			in: query
//			type: string
//		- name: limit
//			description: Display the first N user batches
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array of user batches
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/userBatch"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserBatches(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	managedByUser := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("can_manage != 'none'").
		Select("groups_ancestors_active.child_group_id AS id")

	prefixAncestors := store.ActiveGroupAncestors().Where("child_group_id = ?", groupID).
		Select("ancestor_group_id AS id")

	query := store.UserBatches().
		Joins("JOIN user_batch_prefixes USING(group_prefix)").
		Where(`user_batch_prefixes.group_id IN(?)`, managedByUser.QueryExpr()).
		Where(`user_batch_prefixes.group_id IN(?)`, prefixAncestors.QueryExpr()).
		Select("user_batches_v2.group_prefix, user_batches_v2.custom_prefix, user_batches_v2.size, user_batches_v2.creator_id")

	query, apiErr := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"group_prefix":  {ColumnName: "user_batches_v2.group_prefix"},
				"custom_prefix": {ColumnName: "user_batches_v2.custom_prefix"},
				"size":          {ColumnName: "user_batches_v2.size"},
			},
			DefaultRules: "group_prefix,custom_prefix",
			TieBreakers: service.SortingAndPagingTieBreakers{
				"group_prefix":  service.FieldTypeString,
				"custom_prefix": service.FieldTypeString,
			},
		})
	if apiErr != service.NoError {
		return apiErr
	}

	var result []userBatch
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
