package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

const (
	// OriginGroupMembership is the origin for permissions granted to the group via membership in another group.
	OriginGroupMembership = "group_membership"
	// OriginItemUnlocking is the origin for permissions granted to the group via unlocking an item.
	OriginItemUnlocking = "item_unlocking"
	// OriginSelf is the origin for permissions granted to the creator of the item.
	OriginSelf = "self"
	// OriginOther is the origin for permissions granted to the group via other means.
	OriginOther = "other"
	// OriginComputed is a fake origin for computed permissions used by algorithms related to adding CanRequestHelpTo into output.
	OriginComputed = "computed"
	// OriginGranted is a fake origin for computed permissions used by algorithms related to adding CanRequestHelpTo into output.
	OriginGranted = "granted"
)

type permissionsStruct struct {
	structures.ItemPermissions
	// required: true
	CanMakeSessionOfficial bool `json:"can_make_session_official"`
}

// The group which can be asked for help.
type canRequestHelpTo struct {
	// required: true
	ID int64 `json:"id,string"`
	// The name is present only if the group is visible to the current user.
	// required: false
	Name *string `json:"name,omitempty"`
	// Whether the group is the "all-users" group.
	// required: true
	IsAllUsersGroup bool `json:"is_all_users_group"`
}

// Permissions granted directly to the group via `origin` = 'group_membership' and `source_group_id` = `{source_group_id}`.
type grantedPermissionsStruct struct {
	permissionsStruct
	// required: true
	CanEnterFrom string `json:"can_enter_from"`
	// required: true
	CanEnterUntil string `json:"can_enter_until"`
	// Nullable
	// required: true
	CanRequestHelpTo *canRequestHelpTo `json:"can_request_help_to"`
}

type aggregatedPermissionsWithCanEnterFromStruct struct {
	permissionsStruct
	// The next time the group can enter the item (>= NOW())
	// required: true
	CanEnterFrom string `json:"can_enter_from"`
	// required: true
	CanRequestHelpTo []canRequestHelpTo `json:"can_request_help_to"`
}

// Computed permissions for the group
// (respecting permissions of its ancestors).
// It combines the aggregation of permissions from the given group and its ancestors,
// with the propagation of permissions from all ancestor items computed in `permissions_generated`.
type computedPermissions struct {
	aggregatedPermissionsWithCanEnterFromStruct
}

// Permissions granted to the group or its ancestors
// via `origin` = 'group_membership' excluding the row from `granted`.
type permissionsGrantedViaGroupMembership struct {
	aggregatedPermissionsWithCanEnterFromStruct
}

// Permissions granted to the group or its ancestors
// via `origin` = 'item_unlocking'.
type permissionsGrantedViaItemUnlocking struct {
	aggregatedPermissionsWithCanEnterFromStruct
}

// Permissions granted to the group or its ancestors
// via `origin` = 'self'.
type permissionsGrantedViaSelf struct {
	aggregatedPermissionsWithCanEnterFromStruct
}

// Permissions granted to the group or its ancestors
// via `origin` = 'other'.
type permissionsGrantedViaOther struct {
	aggregatedPermissionsWithCanEnterFromStruct
}

// swagger:model permissionsViewResponse
type permissionsViewResponse struct {
	// required: true
	Granted grantedPermissionsStruct `json:"granted"`
	// required: true
	Computed computedPermissions `json:"computed"`
	// required: true
	GrantedViaGroupMembership permissionsGrantedViaGroupMembership `json:"granted_via_group_membership"`
	// required: true
	GrantedViaItemUnlocking permissionsGrantedViaItemUnlocking `json:"granted_via_item_unlocking"`
	// required: true
	GrantedViaSelf permissionsGrantedViaSelf `json:"granted_via_self"`
	// required: true
	GrantedViaOther permissionsGrantedViaOther `json:"granted_via_other"`
}

type canRequestHelpToPermissionsRaw struct {
	Origin            string
	SourceGroupID     int64
	PermissionGroupID int64
	PermissionItemID  int64
	GroupID           int64
	GroupName         string
}

// swagger:operation GET /groups/{source_group_id}/permissions/{group_id}/{item_id} groups permissionsView
//
//	---
//	summary: View permissions
//	description: Lets a manager of a group view permissions on an item for the group.
//
//		Used to see the aggregated permissions a group has on an item,
//		by `origin`,
//		besides the permissions given directly by the group `source_group_id`
//		(which are shown in "granted").
//
//		See documentation about
//		[aggregation](https://france-ioi.github.io/algorea-devdoc/design/access-rights/items/#aggregation-of-permissions-from-multiple-sources)
//		as well as a UI image on how this service is used to see the permissions.
//
//		* The current user must be a manager (with `can_grant_group_access` permission)
//			of `{source_group_id}` which should be an ancestor of the `{group_id}`.
//
//		* The current user must have `can_grant_view` > 'none' or
//			`can_watch` = 'answer_with_grant' or `can_edit` = 'all_with_grant' on `{item_id}` on the item.
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//		- name: source_group_id
//			in: path
//			required: true
//			type: integer
//		- name: item_id
//			in: path
//			required: true
//			type: integer
//	responses:
//		"200":
//			description: OK. Permissions for the group.
//			schema:
//				"$ref": "#/definitions/permissionsViewResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getPermissions(w http.ResponseWriter, r *http.Request) service.APIError {
	sourceGroupID, err := service.ResolveURLQueryPathInt64Field(r, "source_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	apiErr := checkIfUserIsManagerAllowedToGrantPermissionsToGroupID(store, user, sourceGroupID, groupID)
	if apiErr != service.NoError {
		return apiErr
	}

	found, err := store.Permissions().MatchingUserAncestors(user).
		Where("? OR can_watch_generated = 'answer_with_grant' OR can_edit_generated = 'all_with_grant'",
			store.PermissionsGranted().PermissionIsAtLeastSQLExpr("grant_view", enter)).
		Where("item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	var permissions []map[string]interface{}

	const canMakeSessionOfficialColumn = "IFNULL(MAX(can_make_session_official), 0) AS can_make_session_official"

	// This sub-query retrieves a single row from permissions_granted because it does a "where" on the four columns
	// composing the primary key (group_id, item_id, source_group_id, origin).
	//
	// Tricky part:
	// We want this query to return a row with the default values when there is no matching row in permissions_granted.
	//
	// Reason:
	// If this query returns nothing (no rows),
	// then the bigger query this query is a part of would return nothing as well, and we don't want that,
	// because we need the values from the other parts of the bigger query.
	//
	// We use "IFNULL(MAX(permission_value), default_value)" to return default values when there is no matching row.
	grantedPermissions := store.PermissionsGranted().
		Where("group_id = ?", groupID).
		Where("item_id = ?", itemID).
		Where("source_group_id = ?", sourceGroupID).
		Where("origin = ?", OriginGroupMembership).
		Select(`
			IFNULL(MAX(can_view_value), 1) AS can_view_value,
			IFNULL(MAX(can_grant_view_value), 1) AS can_grant_view_value,
			IFNULL(MAX(can_watch_value), 1) AS can_watch_value,
			IFNULL(MAX(can_edit_value), 1) AS can_edit_value,
			IFNULL(MAX(can_enter_from), '9999-12-31 23:59:59') AS can_enter_from,
			IFNULL(MAX(can_enter_until), '9999-12-31 23:59:59') AS can_enter_until,
			IFNULL(MAX(is_owner), 0) AS is_owner, ` + canMakeSessionOfficialColumn)

	generatedPermissions := store.Permissions().
		Joins("JOIN groups_ancestors_active AS ancestors ON ancestors.ancestor_group_id = permissions.group_id").
		Where("ancestors.child_group_id = ?", groupID).
		Where("item_id = ?", itemID).
		Select(`
			IFNULL(MAX(can_view_generated_value), 1) AS can_view_generated_value,
			IFNULL(MAX(can_grant_view_generated_value), 1) AS can_grant_view_generated_value,
			IFNULL(MAX(can_watch_generated_value), 1) AS can_watch_generated_value,
			IFNULL(MAX(can_edit_generated_value), 1) AS can_edit_generated_value,
			IFNULL(MAX(is_owner_generated), 0) AS is_owner_generated`)

	ancestorPermissions := store.PermissionsGranted().
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.ancestor_group_id = permissions_granted.group_id").
		Where("groups_ancestors_active.child_group_id = ?", groupID)
	ancestorPermissionsOnItem := ancestorPermissions.Where("item_id = ?", itemID)

	const canEnterFromColumn = `
		IFNULL(MIN(IF(
			NOW() BETWEEN can_enter_from AND can_enter_until,
			NOW(),
			IF(can_enter_from BETWEEN NOW() AND can_enter_until, can_enter_from, '9999-12-31 23:59:59')
		)), '9999-12-31 23:59:59') AS can_enter_from`

	grantedPermissionsWithAncestors := ancestorPermissionsOnItem.
		Select(`
			IFNULL(MAX(can_view_value), 1) AS can_view_value,
			IFNULL(MAX(can_grant_view_value), 1) AS can_grant_view_value,
			IFNULL(MAX(can_watch_value), 1) AS can_watch_value,
			IFNULL(MAX(can_edit_value), 1) AS can_edit_value,
			IFNULL(MAX(is_owner), 0) AS is_owner, ` + canMakeSessionOfficialColumn + ", " + canEnterFromColumn)

	aggregatedPermissions := ancestorPermissionsOnItem.Select(canEnterFromColumn + ", " + canMakeSessionOfficialColumn)

	grantedPermissionsGroupMembership := grantedPermissionsWithAncestors.
		Where("origin = ?", OriginGroupMembership).
		Where("NOT (group_id = ? AND source_group_id = ?)", groupID, sourceGroupID)
	grantedPermissionsItemUnlocking := grantedPermissionsWithAncestors.Where("origin = ?", OriginItemUnlocking)
	grantedPermissionsSelf := grantedPermissionsWithAncestors.Where("origin = ?", OriginSelf)
	grantedPermissionsOther := grantedPermissionsWithAncestors.Where("origin = ?", OriginOther)

	err = store.
		Raw(`
			SELECT
				grp.can_view_value AS granted_directly_can_view_value, grp.can_grant_view_value AS granted_directly_can_grant_view_value,
				grp.can_watch_value AS granted_directly_can_watch_value, grp.can_edit_value AS granted_directly_can_edit_value,
				grp.can_make_session_official AS granted_directly_can_make_session_official, grp.can_enter_from AS granted_directly_can_enter_from,
				grp.can_enter_until AS granted_directly_can_enter_until, grp.is_owner AS granted_directly_is_owner,

				gep.can_view_generated_value AS generated_can_view_value, gep.can_grant_view_generated_value AS generated_can_grant_view_value,
				gep.can_watch_generated_value AS generated_can_watch_value, gep.can_edit_generated_value AS generated_can_edit_value,
				gep.is_owner_generated AS generated_is_owner,
				grp_aggregated.can_enter_from AS generated_can_enter_from,
				grp_aggregated.can_make_session_official AS generated_can_make_session_official,

				grp_membership.can_view_value AS granted_anc_membership_can_view_value,
				grp_membership.can_grant_view_value AS granted_anc_membership_can_grant_view_value,
				grp_membership.can_watch_value AS granted_anc_membership_can_watch_value,
				grp_membership.can_edit_value AS granted_anc_membership_can_edit_value,
				grp_membership.can_make_session_official AS granted_anc_membership_can_make_session_official,
				grp_membership.can_enter_from AS granted_anc_membership_can_enter_from,
				grp_membership.is_owner AS granted_anc_membership_is_owner,

				grp_unlocking.can_view_value AS granted_anc_unlocking_can_view_value,
				grp_unlocking.can_grant_view_value AS granted_anc_unlocking_can_grant_view_value,
				grp_unlocking.can_watch_value AS granted_anc_unlocking_can_watch_value,
				grp_unlocking.can_edit_value AS granted_anc_unlocking_can_edit_value,
				grp_unlocking.can_make_session_official AS granted_anc_unlocking_can_make_session_official,
				grp_unlocking.can_enter_from AS granted_anc_unlocking_can_enter_from,
				grp_unlocking.is_owner AS granted_anc_unlocking_is_owner,

				grp_self.can_view_value AS granted_anc_self_can_view_value, grp_self.can_grant_view_value AS granted_anc_self_can_grant_view_value,
				grp_self.can_watch_value AS granted_anc_self_can_watch_value, grp_self.can_edit_value AS granted_anc_self_can_edit_value,
				grp_self.can_make_session_official AS granted_anc_self_can_make_session_official,
				grp_self.can_enter_from AS granted_anc_self_can_enter_from,
				grp_self.is_owner AS granted_anc_self_is_owner,

				grp_other.can_view_value AS granted_anc_other_can_view_value, grp_other.can_grant_view_value AS granted_anc_other_can_grant_view_value,
				grp_other.can_watch_value AS granted_anc_other_can_watch_value, grp_other.can_edit_value AS granted_anc_other_can_edit_value,
				grp_other.can_make_session_official AS granted_anc_other_can_make_session_official,
				grp_other.can_enter_from AS granted_anc_other_can_enter_from,
				grp_other.is_owner AS granted_anc_other_is_owner
			FROM ? AS grp,
			     ? AS gep, ? AS grp_membership, ? AS grp_unlocking, ? AS grp_self, ? AS grp_other, ? AS grp_aggregated`,
			grantedPermissions.SubQuery(), generatedPermissions.SubQuery(), grantedPermissionsGroupMembership.SubQuery(),
			grantedPermissionsItemUnlocking.SubQuery(), grantedPermissionsSelf.SubQuery(), grantedPermissionsOther.SubQuery(),
			aggregatedPermissions.SubQuery()).
		ScanIntoSliceOfMaps(&permissions).Error()
	service.MustNotBeError(err)

	permissionsRow := permissions[0]
	permissionsGrantedStore := store.PermissionsGranted()

	allUsersGroupID := domain.ConfigFromContext(r.Context()).AllUsersGroupID
	canRequestHelpToByOrigin := getCanRequestHelpToByOrigin(ancestorPermissions, store, groupID, itemID, sourceGroupID, allUsersGroupID, user)

	// Filter on "granted" can have a maximum of one match because it is filtered on the primary key.
	// (item_id, group_id, origin, source_group_id).
	// If there is none, we want to return nil.
	var canRequestHelpToPermission *canRequestHelpTo
	if len(canRequestHelpToByOrigin[OriginGranted]) > 0 {
		canRequestHelpToPermission = &canRequestHelpToByOrigin[OriginGranted][0]
	}

	response := permissionsViewResponse{
		Granted: grantedPermissionsStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["granted_directly_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["granted_directly_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["granted_directly_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["granted_directly_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["granted_directly_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["granted_directly_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["granted_directly_can_enter_from"]),
			CanEnterUntil:    service.ConvertDBTimeToJSONTime(permissionsRow["granted_directly_can_enter_until"]),
			CanRequestHelpTo: canRequestHelpToPermission,
		},
		Computed: computedPermissions{aggregatedPermissionsWithCanEnterFromStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["generated_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["generated_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["generated_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["generated_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["generated_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["generated_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["generated_can_enter_from"]),
			CanRequestHelpTo: canRequestHelpToByOrigin[OriginComputed],
		}},
		GrantedViaGroupMembership: permissionsGrantedViaGroupMembership{aggregatedPermissionsWithCanEnterFromStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["granted_anc_membership_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["granted_anc_membership_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["granted_anc_membership_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["granted_anc_membership_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["granted_anc_membership_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["granted_anc_membership_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["granted_anc_membership_can_enter_from"]),
			CanRequestHelpTo: canRequestHelpToByOrigin[OriginGroupMembership],
		}},
		GrantedViaItemUnlocking: permissionsGrantedViaItemUnlocking{aggregatedPermissionsWithCanEnterFromStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["granted_anc_unlocking_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["granted_anc_unlocking_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["granted_anc_unlocking_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["granted_anc_unlocking_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["granted_anc_unlocking_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["granted_anc_unlocking_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["granted_anc_unlocking_can_enter_from"]),
			CanRequestHelpTo: canRequestHelpToByOrigin[OriginItemUnlocking],
		}},
		GrantedViaSelf: permissionsGrantedViaSelf{aggregatedPermissionsWithCanEnterFromStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["granted_anc_self_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["granted_anc_self_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["granted_anc_self_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["granted_anc_self_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["granted_anc_self_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["granted_anc_self_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["granted_anc_self_can_enter_from"]),
			CanRequestHelpTo: canRequestHelpToByOrigin[OriginSelf],
		}},
		GrantedViaOther: permissionsGrantedViaOther{aggregatedPermissionsWithCanEnterFromStruct{
			permissionsStruct: permissionsStruct{
				ItemPermissions: structures.ItemPermissions{
					CanView:      permissionsGrantedStore.ViewNameByIndex(int(permissionsRow["granted_anc_other_can_view_value"].(int64))),
					CanGrantView: permissionsGrantedStore.GrantViewNameByIndex(int(permissionsRow["granted_anc_other_can_grant_view_value"].(int64))),
					CanWatch:     permissionsGrantedStore.WatchNameByIndex(int(permissionsRow["granted_anc_other_can_watch_value"].(int64))),
					CanEdit:      permissionsGrantedStore.EditNameByIndex(int(permissionsRow["granted_anc_other_can_edit_value"].(int64))),
					IsOwner:      permissionsRow["granted_anc_other_is_owner"].(int64) == 1,
				},
				CanMakeSessionOfficial: permissionsRow["granted_anc_other_can_make_session_official"].(int64) == 1,
			},
			CanEnterFrom:     service.ConvertDBTimeToJSONTime(permissionsRow["granted_anc_other_can_enter_from"]),
			CanRequestHelpTo: canRequestHelpToByOrigin[OriginOther],
		}},
	}

	render.Respond(w, r, &response)

	return service.NoError
}

// getCanRequestHelpToByOrigin returns a map of canRequestHelpTo permissions by origin.
// We first get all the can_request_help_to groups, and then we filter them by origin.
func getCanRequestHelpToByOrigin(
	ancestorPermissions *database.DB,
	store *database.DataStore,
	groupID int64,
	itemID int64,
	sourceGroupID int64,
	allUsersGroupID int64,
	user *database.User,
) map[string][]canRequestHelpTo {
	itemAncestorsRequestHelpPropagationQuery := store.Items().GetAncestorsRequestHelpPropagatedQuery(itemID)

	var canRequestHelpToPermissions []canRequestHelpToPermissionsRaw
	err := ancestorPermissions.
		Joins("JOIN `groups` AS can_request_help_to_group ON can_request_help_to_group.id = permissions_granted.can_request_help_to").
		Select(`
			permissions_granted.origin AS origin,
			permissions_granted.source_group_id AS source_group_id,
			permissions_granted.group_id AS permission_group_id,
			permissions_granted.item_id AS permission_item_id,
			can_request_help_to_group.id AS group_id,
			can_request_help_to_group.name AS group_name
		`).
		Where("item_id IN (?)", itemAncestorsRequestHelpPropagationQuery.SubQuery()).
		Scan(&canRequestHelpToPermissions).
		Error()
	service.MustNotBeError(err)

	canRequestHelpToByOrigin := make(map[string][]canRequestHelpTo)
	for _, origin := range []string{OriginGroupMembership, OriginItemUnlocking, OriginSelf, OriginOther, OriginComputed, OriginGranted} {
		canRequestHelpToByOrigin[origin] = filterCanRequestHelpTo(
			store,
			canRequestHelpToPermissions,
			origin,
			groupID,
			itemID,
			sourceGroupID,
			user.GroupID,
			allUsersGroupID,
		)
	}

	return canRequestHelpToByOrigin
}

// filterCanRequestHelpTo filters the canRequestHelpTo permissions to only keep the ones matching the wanted origin.
func filterCanRequestHelpTo(
	store *database.DataStore,
	permissions []canRequestHelpToPermissionsRaw,
	origin string,
	groupID int64,
	itemID int64,
	sourceGroupID int64,
	visibleGroupID int64,
	allUsersGroupID int64,
) []canRequestHelpTo {
	results := make([]canRequestHelpTo, 0)

	for _, canRequestHelpToPermission := range permissions {
		if canRequestHelpToShouldBeAdded(canRequestHelpToPermission, origin, groupID, itemID, sourceGroupID) {
			results = append(results, canRequestHelpToForUser(canRequestHelpToPermission, store, visibleGroupID, allUsersGroupID))
		}
	}

	return uniqueCanRequestHelpTo(results)
}

// canRequestHelpToShouldBeAdded checks whether a canRequestHelpToPermission should be added to the results of a given origin.
func canRequestHelpToShouldBeAdded(
	canRequestHelpToPermission canRequestHelpToPermissionsRaw,
	origin string,
	groupID int64,
	itemID int64,
	sourceGroupID int64,
) bool {
	// Permissions granted on ancestor items are only present in "computed".
	if origin != OriginComputed && canRequestHelpToPermission.PermissionItemID != itemID {
		return false
	}

	// The canRequestHelpToPermission matching "group_membership" origin as well as GroupID and SourceGroupID
	// is a special case that goes into "granted" and "computed", and not into "group_membership".
	if canRequestHelpToPermission.Origin == OriginGroupMembership &&
		canRequestHelpToPermission.PermissionGroupID == groupID &&
		canRequestHelpToPermission.SourceGroupID == sourceGroupID {
		if origin == OriginGranted || origin == OriginComputed {
			return true
		}

		return false
	}

	if origin == OriginComputed || canRequestHelpToPermission.Origin == origin {
		// Otherwise, we want everything in "computed", or everything matching the origin.
		return true
	}

	return false
}

// canRequestHelpToForUser converts a canRequestHelpToPermissionsRaw to a canRequestHelpTo returned to the user.
func canRequestHelpToForUser(
	permission canRequestHelpToPermissionsRaw,
	store *database.DataStore,
	visibleGroupID int64,
	allUsersGroupID int64,
) canRequestHelpTo {
	curCanRequestHelpTo := canRequestHelpTo{
		ID: permission.GroupID,
	}

	if allUsersGroupID == permission.GroupID {
		curCanRequestHelpTo.IsAllUsersGroup = true
		curCanRequestHelpTo.Name = &permission.GroupName
	} else if store.Groups().IsVisibleForGroup(permission.GroupID, visibleGroupID) {
		curCanRequestHelpTo.Name = &permission.GroupName
	}

	return curCanRequestHelpTo
}

// uniqueCanRequestHelpTo removes duplicates from the canRequestHelpTo slice.
func uniqueCanRequestHelpTo(canRequestHelpTos []canRequestHelpTo) []canRequestHelpTo {
	hasID := make(map[int64]bool)
	result := make([]canRequestHelpTo, 0)

	for _, entry := range canRequestHelpTos {
		if _, value := hasID[entry.ID]; !value {
			hasID[entry.ID] = true
			result = append(result, entry)
		}
	}

	return result
}
