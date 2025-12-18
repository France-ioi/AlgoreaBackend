package groups

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model permissionExplanationViewResponseRow
type permissionExplanationViewResponseRow struct {
	SourceGroup *struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Name string `json:"name"`
	} `gorm:"embedded;embedded_prefix:source_group__" json:"source_group,omitempty"`

	Group *struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Name string `json:"name"`
	} `gorm:"embedded;embedded_prefix:group__" json:"group,omitempty"`

	Item *struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Title *string `json:"title"`
		// required: true
		LanguageTag string `json:"language_tag"`
		// required: true
		RequiresExplicitEntry bool `json:"requires_explicit_entry"`
		// required: true
		Type string `json:"type"`
	} `gorm:"embedded;embedded_prefix:item__" json:"item,omitempty"`

	// required: true
	Origin string `json:"origin"`

	// required: true
	GrantedPermissions structures.ItemPermissions `gorm:"embedded;embedded_prefix:granted_permissions__" json:"granted_permissions"`

	// the result of propagation of the original permission to `item_id`
	// required: true
	PropagatedPermissions struct {
		// required: true
		// enum: none,info,content,content_with_descendants,solution
		CanViewGenerated string `json:"can_view_generated"`
		// required: true
		// enum: none,enter,content,content_with_descendants,solution,solution_with_grant
		CanGrantViewGenerated string `json:"can_grant_view_generated"`
		// required: true
		// enum: none,result,answer,answer_with_grant
		CanWatchGenerated string `json:"can_watch_generated"`
		// required: true
		// enum: none,children,all,all_with_grant
		CanEditGenerated string `json:"can_edit_generated"`
		// required: true
		IsOwnerGenerated bool `json:"is_owner_generated"`
	} `gorm:"embedded;embedded_prefix:propagated_permissions__" json:"propagated_permissions"`

	// required: true
	UserCanUpdatePermission bool `json:"user_can_update_permission"`

	// required: true
	// for pagination
	From string `json:"from"`

	// not in response
	UserCanViewItem bool `json:"-"`
}

// swagger:operation GET /groups/{group_id}/permissions/{item_id}/explain groups permissionExplanationView
//
//	---
//	summary: Explain permissions
//	description: >
//
//		Displays explanation of where permissions of the given group on the given item come from, i.e.
//		for each group ancestor (including itself) of `group_id`, it returns entries of `permissions_granted`
//		impacting the permission and how they have propagated to `item_id`.
//		Non-visible item/group permissions are returned as well so that they can be listed.
//
//	  - The current user must be able to view permissions on `{item_id}` given to `{group_id}`, i.e.
//	    1) to have a permission to watch for `{item_id}` with `can_watch`>='result' and to watch for members of `{group_id}`
//	    (have `can_watch_members` on an ancestor of `{group_id}`), or
//	    2) to be able to grant permissions on `{item_id}` and grant permissions to `{group_id}`
//	    (have `can_grant_group_access` on an ancestor of `{group_id}`), or
//	    3) to be a descendant of `{group_id}` or be a team member of a team that is a descendant of `{group_id}`, or
//	    4) to be a manager of `{group_id}` with `can_manage`>='memberships',
//	    otherwise the 'forbidden' error is returned.
//
//	    The fields `group` and `source_group` of result rows are only returned for visible corresponding groups.
//	    A group is considered visible for the current user if
//	    1) it is an ancestor of the current user or it is an ancestor of at least one team the current user is a member of or
//	    2) it is an ancestor of a non-user group that the current user manages with `can_manage`>=”memberships” or
//	    `can_watch_members`=true or `can_grant_group_access`=true (explicitly or implicitly) or
//	    3) it is a user implicitly managed by the current user or it is a member of a team managed by the current user
//	    (explicitly or implicitly) with `can_manage`>=”membership” or `can_watch_members`=true or
//	    `can_grant_group_access`=true or
//	    4) it is public.
//
//	    The `item` field of result rows is only returned for items visible to the current user or to one of their teams
//	    (`can_view`>='info').
//
//	parameters:
//		- name: group_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: item_id
//			in: path
//			required: true
//			type: integer
//			format: int64
//		- name: from
//			in: query
//			type: string
//			description: Start the page from the permission next to the permission with `from`=`{from}`
//		- name: limit
//			description: Display the first N permissions
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Permission explanation for the group and the item.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/permissionExplanationViewResponseRow"
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
func (srv *Service) getPermissionExplanation(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var from string
	urlParams := httpRequest.URL.Query()
	if len(urlParams["from"]) > 0 {
		from = httpRequest.URL.Query().Get("from")
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	groupManagementPermissionsQuery := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = ?", groupID).
		Select(`
			MAX(can_watch_members) AS can_watch_members,
			MAX(can_grant_group_access) AS can_grant_group_access,
			MAX(can_manage_value) AS can_manage_value`)
	itemManagementPermissionsQuery := store.Permissions().
		Select(`
			MAX(can_grant_view_generated_value) AS can_grant_view_generated_value,
			MAX(can_watch_generated_value) AS can_watch_generated_value`).
		Joins("JOIN groups_ancestors_active AS ancestors ON ancestors.ancestor_group_id = permissions.group_id").
		Where("ancestors.child_group_id = ?", user.GroupID).
		Where("permissions.item_id = ?", itemID)
	found, err := store.Raw(`
		SELECT 1 FROM ? AS group_management_permissions
		JOIN ? AS item_permissions ON 1
		WHERE
			(item_permissions.can_watch_generated_value >= ? AND group_management_permissions.can_watch_members) OR
			(item_permissions.can_grant_view_generated_value >= ? AND group_management_permissions.can_grant_group_access) OR
			group_management_permissions.can_manage_value >= ? OR
			(SELECT 1 FROM groups_ancestors_active WHERE ancestor_group_id = ? AND child_group_id = ? LIMIT 1) OR
			(SELECT 1 FROM groups_groups_active WHERE parent_group_id = ? AND child_group_id = ? AND is_team_membership LIMIT 1)`,
		groupManagementPermissionsQuery.SubQuery(), itemManagementPermissionsQuery.SubQuery(),
		store.PermissionsGranted().WatchIndexByName("result"),
		store.PermissionsGranted().GrantViewIndexByName("enter"),
		store.GroupManagers().CanManageIndexByName("memberships"),
		groupID, user.GroupID, groupID, user.GroupID).HasRows()

	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	var result []permissionExplanationViewResponseRow

	// Use a fixed DB connection instead of a transaction to avoid blocking other requests
	err = store.WithFixedConnection(func(db *database.DB) error {
		permissionsGrantedStore := database.NewDataStore(db).PermissionsGranted()

		// group_id in these temporary tables is constructed as {group_id}_{item_id whose permissions we want to propagate}
		cleanupFunc, err := permissionsGrantedStore.CreateTemporaryTablesForPermissionsExplanation()
		defer cleanupFunc()
		service.MustNotBeError(err)

		insertGrantedPermissionsToBeExplained(db, itemID, groupID)

		service.MustNotBeError(db.Exec(`
			INSERT INTO permissions_propagate_exp (group_id, item_id, propagate_to)
			SELECT group_id, item_id, 'self'
			FROM permissions_granted_exp`).Error())

		service.MustNotBeError(permissionsGrantedStore.ComputePermissionsExplanation(&itemID))

		explanationQuery := db.Table("permissions_generated_exp").
			Select(`
				permissions_generated_exp.hashed_group_id AS `+"`from`"+`,
				source_group.id AS source_group__id, source_group.name AS source_group__name,
				`+"`group`"+`.id AS group__id, `+"`group`"+`.name AS group__name,
				items.id AS item__id,
				COALESCE(user_strings.title, default_strings.title) AS item__title,
				IF(user_strings.title IS NULL, default_strings.language_tag, user_strings.language_tag) AS item__language_tag,
				items.requires_explicit_entry AS item__requires_explicit_entry,
				items.type AS item__type,
				(
					permissions.can_view_generated_value != 1 /* none */ OR
					EXISTS(
						SELECT 1
						FROM groups_groups_active
						JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id
						JOIN permissions_generated ON
							permissions_generated.group_id = groups_ancestors_active.ancestor_group_id AND
							permissions_generated.item_id = items.id AND
							permissions_generated.can_view_generated != 'none'
						WHERE groups_groups_active.child_group_id = ? AND groups_groups_active.is_team_membership
					)
				) AS user_can_view_item,
				permissions_granted.origin AS origin,
				permissions_granted.can_view AS granted_permissions__can_view,
				permissions_granted.can_grant_view AS granted_permissions__can_grant_view,
				permissions_granted.can_watch AS granted_permissions__can_watch,
				permissions_granted.can_edit AS granted_permissions__can_edit,
				permissions_granted.is_owner AS granted_permissions__is_owner,
				permissions_generated_exp.can_view_generated AS propagated_permissions__can_view_generated,
				permissions_generated_exp.can_grant_view_generated AS propagated_permissions__can_grant_view_generated,
				permissions_generated_exp.can_watch_generated AS propagated_permissions__can_watch_generated,
				permissions_generated_exp.can_edit_generated AS propagated_permissions__can_edit_generated,
				permissions_generated_exp.is_owner_generated AS propagated_permissions__is_owner_generated,
				(permissions_granted.origin = 'group_membership' AND
					(
						permissions.can_grant_view_generated_value > 1 /* none */ OR
						permissions.can_watch_generated_value = ? OR
						permissions.can_edit_generated_value = ?
					) AND
					EXISTS(
						SELECT 1
						FROM groups_user_can_grant_group_access_to
						WHERE groups_user_can_grant_group_access_to.id = permissions_granted.source_group_id
					)
				) AS user_can_update_permission`,
				user.GroupID,
				permissionsGrantedStore.WatchIndexByName("answer_with_grant"),
				permissionsGrantedStore.EditIndexByName("all_with_grant")).
			Joins(`
				JOIN permissions_granted
				  ON permissions_granted.group_id = permissions_generated_exp.permissions_granted_group_id AND
				     permissions_granted.item_id = permissions_generated_exp.permissions_granted_item_id AND
				     permissions_granted.source_group_id = permissions_generated_exp.permissions_granted_source_group_id AND
				     permissions_granted.origin = permissions_generated_exp.permissions_granted_origin`).
			Joins("LEFT JOIN `groups` AS source_group ON source_group.id = permissions_granted.source_group_id AND "+
				groupVisibilityConditionForPermissionsExplanation(user, "source_group")).
			Joins("LEFT JOIN `groups` AS `group` ON group.id = permissions_generated_exp.group_id AND "+
				groupVisibilityConditionForPermissionsExplanation(user, "group")).
			Joins("LEFT JOIN `items` ON items.id = permissions_granted.item_id").
			JoinsPermissionsForGroupToItems(user.GroupID).
			Where("permissions_generated_exp.item_id = ?", itemID).
			// filter out not affecting permissions
			Where(`
				permissions_generated_exp.can_view_generated != 'none' OR
				permissions_generated_exp.can_grant_view_generated != 'none' OR
				permissions_generated_exp.can_watch_generated != 'none' OR
				permissions_generated_exp.can_edit_generated != 'none' OR
				permissions_generated_exp.is_owner_generated`).
			JoinsUserAndDefaultItemStrings(user).
			Order(`
				permissions_granted.group_id, permissions_granted.item_id,
				permissions_granted.source_group_id, permissions_granted.origin`).
			With("groups_user_can_grant_group_access_to",
				store.ActiveGroupAncestors().ManagedByUser(user).
					Select("groups_ancestors_active.child_group_id AS id").
					Having("MAX(can_grant_group_access)").
					Group("groups_ancestors_active.child_group_id"))

		explanationQuery = service.NewQueryLimiter().Apply(httpRequest, explanationQuery)

		startFromRowQuery := service.FromFirstRow()
		if from != "" {
			var fromID struct {
				GroupID       int64  `json:"group_id"`
				ItemID        int64  `json:"item_id"`
				SourceGroupID int64  `json:"source_group_id"`
				Origin        string `json:"origin"`
			}
			err = db.Table("permissions_generated_exp").
				Select(`
					permissions_granted_group_id AS group_id, permissions_granted_item_id AS item_id,
					permissions_granted_source_group_id AS source_group_id, permissions_granted_origin AS origin
				`).
				Where("hashed_group_id = ?", from).
				Where("item_id = ?", itemID).
				Take(&fromID).Error()
			if !gorm.IsRecordNotFoundError(err) {
				service.MustNotBeError(err)
				startFromRowQuery = db.Raw("SELECT ? AS group_id, ? AS item_id, ? AS source_group_id, ? AS origin",
					fromID.GroupID, fromID.ItemID, fromID.SourceGroupID, fromID.Origin)
			}
		}

		explanationQuery, err = service.ApplySortingAndPaging(httpRequest, explanationQuery, &service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"group_id":        {ColumnName: "permissions_generated_exp.permissions_granted_group_id"},
				"item_id":         {ColumnName: "permissions_generated_exp.permissions_granted_item_id"},
				"source_group_id": {ColumnName: "permissions_generated_exp.permissions_granted_source_group_id"},
				"origin":          {ColumnName: "permissions_generated_exp.permissions_granted_origin"},
			},
			TieBreakers: service.SortingAndPagingTieBreakers{
				"group_id":        service.FieldTypeInt64,
				"item_id":         service.FieldTypeInt64,
				"source_group_id": service.FieldTypeInt64,
				"origin":          service.FieldTypeString,
			},
			DefaultRules:        "group_id,item_id,source_group_id,origin",
			IgnoreSortParameter: true,
			StartFromRowQuery:   startFromRowQuery,
		})
		service.MustNotBeError(err)
		service.MustNotBeError(explanationQuery.Scan(&result).Error())

		return nil
	})

	service.MustNotBeError(err)

	postProcessPermissionExplanation(result)

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

func postProcessPermissionExplanation(result []permissionExplanationViewResponseRow) {
	for resultIndex := range result {
		if !result[resultIndex].UserCanViewItem {
			result[resultIndex].Item = nil
		}
		if result[resultIndex].SourceGroup != nil && result[resultIndex].SourceGroup.ID == 0 {
			result[resultIndex].SourceGroup = nil
		}
		if result[resultIndex].Group != nil && result[resultIndex].Group.ID == 0 {
			result[resultIndex].Group = nil
		}
	}
}

func insertGrantedPermissionsToBeExplained(db *database.DB, itemID, groupID int64) {
	service.MustNotBeError(db.Exec(`
		INSERT INTO permissions_granted_exp
			(group_id, item_id, source_group_id, origin, can_view, can_grant_view, can_watch, can_edit, is_owner) ?`,
		database.NewDataStore(db).PermissionsGranted().
			Select(`
				CONCAT(permissions_granted.group_id, '|',
				       permissions_granted.item_id, '|',
				       permissions_granted.source_group_id, '|',
				       permissions_granted.origin) AS group_id,
				permissions_granted.item_id,
				permissions_granted.source_group_id,
				permissions_granted.origin,
				permissions_granted.can_view,
				permissions_granted.can_grant_view,
				permissions_granted.can_watch,
				permissions_granted.can_edit,
				permissions_granted.is_owner`).
			Where(
				"permissions_granted.item_id IN (SELECT ? UNION SELECT ancestor_item_id FROM items_ancestors WHERE child_item_id = ?)",
				itemID, itemID).
			// filter out empty permissions_granted
			Where(`
				permissions_granted.can_view != 'none' OR
				permissions_granted.can_grant_view != 'none' OR
				permissions_granted.can_watch != 'none' OR
				permissions_granted.can_edit != 'none' OR
				permissions_granted.is_owner`).
			Joins(`
				JOIN groups_ancestors_active
					ON groups_ancestors_active.ancestor_group_id = permissions_granted.group_id AND
						 groups_ancestors_active.child_group_id = ?`, groupID).
			QueryExpr()).Error())
}

func groupVisibilityConditionForPermissionsExplanation(user *database.User, groupsTableAlias string) string {
	return `(
		` + groupsTableAlias + `.is_public OR
		EXISTS( /* the group is an ancestor of the user */
			SELECT 1
			FROM groups_ancestors_active
			WHERE
				groups_ancestors_active.child_group_id = ` + strconv.FormatInt(user.GroupID, 10) + ` AND
				groups_ancestors_active.ancestor_group_id = ` + groupsTableAlias + `.id
		) OR ( /* the group is an ancestor of a team the current user is a member of */
			` + groupsTableAlias + `.type != 'User' AND
			EXISTS(
				SELECT 1
				FROM groups_groups_active
				JOIN groups_ancestors_active ON
					groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id AND
					groups_ancestors_active.ancestor_group_id = ` + groupsTableAlias + `.id
				WHERE
					groups_groups_active.child_group_id = ` + strconv.FormatInt(user.GroupID, 10) + ` AND
					groups_groups_active.is_team_membership
			)
		) OR ( /* the group is an ancestor of a non-user group managed by the current user */
			` + groupsTableAlias + `.type != 'User' AND
			EXISTS(
				SELECT 1
				FROM groups_ancestors_active
				JOIN group_managers ON
					group_managers.group_id = groups_ancestors_active.ancestor_group_id AND
					(
						group_managers.can_manage != 'none' OR
						group_managers.can_grant_group_access OR
						group_managers.can_watch_members
					)
				JOIN groups_ancestors_active AS user_ancestors ON
					user_ancestors.ancestor_group_id = group_managers.manager_id AND
					user_ancestors.child_group_id = ` + strconv.FormatInt(user.GroupID, 10) + `
				JOIN groups_ancestors_active AS group_descendants ON
					group_descendants.child_group_id = groups_ancestors_active.child_group_id AND
					group_descendants.ancestor_group_id = ` + groupsTableAlias + `.id
				WHERE
					groups_ancestors_active.child_group_type != 'User'
			)
		) OR (
			` + groupsTableAlias + `.type = 'User' AND (
				EXISTS( /* the group is a user implicitly managed by the current user */
					SELECT 1
					FROM groups_ancestors_active
					JOIN group_managers ON
						group_managers.group_id = groups_ancestors_active.ancestor_group_id  AND
						(
							group_managers.can_manage != 'none' OR
							group_managers.can_grant_group_access OR
							group_managers.can_watch_members
						)
					JOIN groups_ancestors_active AS user_ancestors ON
						user_ancestors.ancestor_group_id = group_managers.manager_id AND
						user_ancestors.child_group_id = ` + strconv.FormatInt(user.GroupID, 10) + `
					WHERE
						groups_ancestors_active.child_group_id = ` + groupsTableAlias + `.id AND
						NOT groups_ancestors_active.is_self
				) OR
				EXISTS( /* the group is a user in a team managed by the current user */
					SELECT 1
					FROM groups_ancestors_active
					JOIN group_managers ON
						group_managers.group_id = groups_ancestors_active.ancestor_group_id  AND
						(
							group_managers.can_manage != 'none' OR
							group_managers.can_grant_group_access OR
							group_managers.can_watch_members
						)
					JOIN groups_ancestors_active AS user_ancestors ON
						user_ancestors.ancestor_group_id = group_managers.manager_id AND
						user_ancestors.child_group_id = ` + strconv.FormatInt(user.GroupID, 10) + `
					JOIN groups_groups_active ON
						groups_groups_active.parent_group_id = groups_ancestors_active.child_group_id AND
						groups_groups_active.child_group_id = ` + groupsTableAlias + `.id AND
						groups_groups_active.is_team_membership
					WHERE
						groups_ancestors_active.child_group_type = 'Team'
				)
			)
		)
	)`
}
