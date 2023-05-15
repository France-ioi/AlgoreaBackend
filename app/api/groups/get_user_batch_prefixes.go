package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model userBatchPrefix
type userBatchPrefix struct {
	// required: true
	GroupPrefix string `json:"group_prefix"`
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	MaxUsers int `json:"max_users"`
	// total number of users in the batches already using this group_prefix
	// required: true
	TotalSize int `json:"total_size"`
}

// swagger:operation GET /groups/{group_id}/user-batch-prefixes groups userBatchPrefixesView
//
//	---
//	summary: List user-batch prefixes
//	description: >
//
//		Lists the user-batch prefixes  with `allow_new` = 1 matching the input group's ancestors
//		that are managed by the current user with 'can_manage:membership' permission
//		(i.e., the `group_id` is a descendant of `user_batch_prefixes.group_id`).
//
//
//		The authenticated user should be a manager of `group_id` with 'can_manage:membership' permission at least,
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//		- name: sort
//			in: query
//			default: [group_prefix]
//			type: array
//			items:
//				type: string
//				enum: [group_prefix,-group_prefix]
//		- name: from.group_prefix
//			description: Start the page from the prefix next to the prefix with `user_batch_prefixes.group_prefix` = `{from.group_prefix}`
//			in: query
//			type: string
//		- name: limit
//			description: Display the first N user-batch prefixes
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array of user-batch prefixes
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/userBatchPrefix"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserBatchPrefixes(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanManageTheGroupMemberships(store, user, groupID); apiError != service.NoError {
		return apiError
	}

	managedByUser := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("can_manage != 'none'").
		Select("groups_ancestors_active.child_group_id AS id")

	query := store.UserBatchPrefixes().
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.ancestor_group_id = user_batch_prefixes.group_id AND
				   groups_ancestors_active.child_group_id = ?`, groupID).
		Where("allow_new").
		Where("user_batch_prefixes.group_id IN (?)", managedByUser.QueryExpr()).
		Select(`
			group_prefix, group_id, max_users,
			(SELECT COUNT(*) FROM user_batches
			 WHERE user_batches.group_prefix = user_batch_prefixes.group_prefix) AS total_size`)

	query, apiErr := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"group_prefix": {ColumnName: "group_prefix"},
			},
			DefaultRules: "group_prefix",
			TieBreakers:  service.SortingAndPagingTieBreakers{"group_prefix": service.FieldTypeString},
		})
	if apiErr != service.NoError {
		return apiErr
	}

	var result []userBatchPrefix
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
