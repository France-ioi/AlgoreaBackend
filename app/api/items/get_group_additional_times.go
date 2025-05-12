package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model additionalTimes
type additionalTimes struct {
	// required: true
	AdditionalTime int32 `json:"additional_time"`
	// required: true
	TotalAdditionalTime int32 `json:"total_additional_time"`
}

// swagger:operation GET /items/{item_id}/groups/{group_id}/additional-times items itemGetAdditionalTime
//
//	---
//	summary: Get additional time for a time-limited item and a group
//	description: >
//							 For the given group and the given time-limited item, the service returns `additional_time` & `total_additional_time`:
//
//
//							 * `additional_time` (in seconds) defaults to 0 if no such `group_item_additional_times`
//
//							 * `total_additional_time` (in seconds) is the sum of additional times of this group on the item through all its
//								 `groups_ancestors` (even from different branches, but each ancestor counted only once), defaulting to 0.
//
//							 Restrictions:
//								 * `item_id` should be a time-limited item (with duration <> NULL);
//								 * the authenticated user should have `can_view` >= 'content', `can_grant_view` >= 'enter',
//									 and `can_watch` >= 'result' on the input item;
//								 * the authenticated user should be a manager of the `group_id`
//									 with `can_grant_group_access` and `can_watch_members` permissions.
//	parameters:
//		- name: item_id
//			description: "`id` of a time-limited item"
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			description: OK. Success response with additional times
//			schema:
//				"$ref": "#/definitions/additionalTimes"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupAdditionalTimes(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	store := srv.GetStore(r)
	found, err := store.Items().TimeLimitedByIDManagedByUser(itemID, user).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	ok, err := store.Groups().ManagedBy(user).Where("groups.id = ?", groupID).
		Having("MAX(can_grant_group_access) AND MAX(can_watch_members)").HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.InsufficientAccessRightsError
	}

	query := store.Groups().Where("groups.id = ?", groupID).
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.child_group_id = groups.id`).
		Joins(`
			LEFT JOIN group_item_additional_times ON group_item_additional_times.group_id = groups_ancestors_active.ancestor_group_id AND
				group_item_additional_times.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN group_item_additional_times AS main_group_item_additional_time ON main_group_item_additional_time.group_id = groups.id AND
				main_group_item_additional_time.item_id = ?`, itemID).
		Select(`
				IFNULL(TIME_TO_SEC(MAX(main_group_item_additional_time.additional_time)), 0) AS additional_time,
				IFNULL(SUM(TIME_TO_SEC(group_item_additional_times.additional_time)), 0) AS total_additional_time`).
		Group("groups.id")

	var result additionalTimes
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
