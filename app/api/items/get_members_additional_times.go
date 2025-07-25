package items

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model itemAdditionalTimesInfo
type itemAdditionalTimesInfo struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	Name string `json:"name"`
	// required: true
	Type string `json:"type"`
	// required: true
	AdditionalTime int32 `json:"additional_time"`
	// required: true
	TotalAdditionalTime int32 `json:"total_additional_time"`
}

// swagger:operation GET /items/{item_id}/groups/{group_id}/members/additional-times items itemListMembersAdditionalTime
//
//	---
//	summary: List additional times on a time-limited item for a group
//	description: >
//							 For all descendant
//
//								 * teams if `items.entry_participant_type` = 'Team'
//								 * end-users groups otherwise
//
//							 linked to the item via `attempts.root_item_id`
//							 and able to view (at least 'can_view:info') or enter (`can_enter_from` < `can_enter_until`) the item,
//							 the service returns their `group_id`, `name`, `type` and `additional_time` & `total_additional_time`.
//
//
//							 * `additional_time` (in seconds) defaults to 0 if no such `group_item_additional_times`
//
//							 * `total_additional_time` (in seconds) is the sum of additional times of this group on the item through all its
//								 `groups_ancestors` (even from different branches, but each ancestor counted only once), defaulting to 0
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
//		- name: from.id
//			description: Start the page from the group next to the group with `id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [name,id]
//			type: array
//			items:
//				type: string
//				enum: [name,-name,id,-id]
//		- name: limit
//			description: Display the first N groups
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with item's info
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/itemAdditionalTimesInfo"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getMembersAdditionalTimes(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	store := srv.GetStore(httpRequest)
	participantType, err := getParticipantTypeForTimeLimitedItemManagedByUser(store, itemID, user)
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	ok, err := store.Groups().ManagedBy(user).Where("groups.id = ?", groupID).
		Having("MAX(can_grant_group_access) AND MAX(can_watch_members)").HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.ErrAPIInsufficientAccessRights
	}

	query := store.ActiveGroupAncestors().Where("groups_ancestors_active.ancestor_group_id = ?", groupID).
		Joins(`
			JOIN `+"`groups`"+` AS found_group
				ON found_group.id = groups_ancestors_active.child_group_id AND found_group.type = ?`, participantType).
		Joins(`
			JOIN attempts
				ON attempts.participant_id = found_group.id AND attempts.root_item_id = ?`, itemID).
		Joins(`
			JOIN groups_ancestors_active AS found_group_ancestors
				ON found_group_ancestors.child_group_id = found_group.id`).
		Joins(`
			LEFT JOIN permissions_generated ON permissions_generated.group_id = found_group_ancestors.ancestor_group_id AND
				permissions_generated.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN permissions_granted ON permissions_granted.group_id = found_group_ancestors.ancestor_group_id AND
				permissions_granted.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN group_item_additional_times ON group_item_additional_times.group_id = found_group_ancestors.ancestor_group_id AND
				group_item_additional_times.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN group_item_additional_times AS main_group_item_additional_time ON main_group_item_additional_time.group_id = found_group.id AND
				main_group_item_additional_time.item_id = ?`, itemID).
		Select(`
				found_group.id AS group_id,
				found_group.name,
				found_group.type,
				IFNULL(TIME_TO_SEC(MAX(main_group_item_additional_time.additional_time)), 0) AS additional_time,
				IFNULL(SUM(TIME_TO_SEC(group_item_additional_times.additional_time)), 0) AS total_additional_time`).
		Group("found_group.id").
		Having(`
			MAX(permissions_generated.can_view_generated_value) >= ? OR
			MAX(permissions_granted.can_enter_from < permissions_granted.can_enter_until)`,
			store.PermissionsGranted().ViewIndexByName("info"))

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"name": {ColumnName: "found_group.name"},
				"id":   {ColumnName: "found_group.id"},
			},
			DefaultRules: "name,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []itemAdditionalTimesInfo
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

func getParticipantTypeForTimeLimitedItemManagedByUser(
	store *database.DataStore, itemID int64, user *database.User,
) (string, error) {
	var participantType string
	err := store.Items().TimeLimitedByIDManagedByUser(itemID, user).
		PluckFirst("items.entry_participant_type", &participantType).Error()
	return participantType, err
}
