package contests

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /contests/{item_id}/groups/{group_id}/members/additional-times contests contestListMembersAdditionalTime
// ---
// summary: List additional times on a contest
// description: >
//                For all
//
//                  * descendant teams linked to the item via `team_item_id` if `items.has_attempts`
//                  * end-users groups otherwise
//
//                having at least 'info' access to the item, the service returns their
//                `group_id`, `name`, `type` and `additional_time` & `total_additional_time`.
//
//
//                * `additional_time` defaults to 0 if no such `groups_contest_items`
//
//                * `total_additional_time` is the sum of additional times of this group on the item through all its
//                  `groups_ancestors` (even from different branches, but each ancestors counted only once), defaulting to 0
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * the authenticated user should have `solutions` or `full` access on the input item;
//                  * the authenticated user should be a manager of the `group_id`.
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
// - name: from.name
//   description: Start the page from the group next to the group with `name` = `from.name` and `id` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the group next to the group with `name` = `from.name` and `id`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [name,id]
//   type: array
//   items:
//     type: string
//     enum: [name,-name,id,-id]
// - name: limit
//   description: Display the first N groups
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with contests info
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/contestInfo"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getMembersAdditionalTimes(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	isTeamOnly, err := srv.isTeamOnlyContestManagedByUser(itemID, user)
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	ok, err := srv.Store.Groups().ManagedBy(user).Where("groups.id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.InsufficientAccessRightsError
	}

	query := srv.Store.ActiveGroupAncestors().Where("groups_ancestors_active.ancestor_group_id = ?", groupID)

	if isTeamOnly {
		query = query.
			Joins(`
				JOIN `+"`groups`"+` AS found_group
					ON found_group.id = groups_ancestors_active.child_group_id AND found_group.type = 'Team' AND
						(found_group.team_item_id IN (SELECT ancestor_item_id FROM items_ancestors WHERE child_item_id = ?) OR
						 found_group.team_item_id = ?)`, itemID, itemID)
	} else {
		query = query.
			Joins(`
				JOIN ` + "`groups`" + ` AS found_group
					ON found_group.id = groups_ancestors_active.child_group_id AND found_group.type = 'UserSelf'`)
	}

	query = query.
		Joins(`
			JOIN groups_ancestors_active AS found_group_ancestors
				ON found_group_ancestors.child_group_id = found_group.id`).
		Joins(`
			LEFT JOIN permissions_generated ON permissions_generated.group_id = found_group_ancestors.ancestor_group_id AND
				permissions_generated.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN groups_contest_items ON groups_contest_items.group_id = found_group_ancestors.ancestor_group_id AND
				groups_contest_items.item_id = ?`, itemID).
		Joins(`
			LEFT JOIN groups_contest_items AS main_group_contest_item ON main_group_contest_item.group_id = found_group.id AND
				main_group_contest_item.item_id = ?`, itemID).
		Select(`
				found_group.id AS group_id,
				found_group.name,
				found_group.type,
				IFNULL(TIME_TO_SEC(MAX(main_group_contest_item.additional_time)), 0) AS additional_time,
				IFNULL(SUM(TIME_TO_SEC(groups_contest_items.additional_time)), 0) AS total_additional_time`).
		Group("found_group.id").
		HavingMaxPermissionGreaterThan("view", "none")

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "found_group.name", FieldType: "string"},
			"id":   {ColumnName: "found_group.id", FieldType: "int64"}},
		"name,id", "id", false)
	if apiError != service.NoError {
		return apiError
	}

	var result []contestInfo
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
