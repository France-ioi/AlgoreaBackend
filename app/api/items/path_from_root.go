package items

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// ItemPath represents a path to an item.
// swagger:model ItemPath
type ItemPath struct {
	// required:true
	Path []string `json:"path"`
	// required:true
	IsStarted bool `json:"is_started"`
}

type rawItemPath struct {
	Path      string
	IsStarted bool
	Sorting   string
}

// swagger:operation GET /items/{item_id}/path-from-root items itemPathFromRootFind
//
//	---
//	summary: Find an item path
//	description: >
//		Finds a path from any of root items to a given item.
//
//		The path consists only of the items visible to the participant
//		(`can_view`>='content' for all the items except for the final one and `can_view`>='info' for the final one).
//		The chain of attempts in the path cannot have missing results for non-final items that require explicit entry.
//		It also cannot have not-started results within or below ended or non-submission-allowing attempts for non-final items.
//
//
//		Note that the path may contain items without results for its final item or non-final items not requiring explicit entry.
//		Also, the path may contain not-started results for its final item even within or below ended or non-submission-allowing attempts.
//		It is even possible that the final item has no linked attempt at all while the final item requires explicit entry.
//
//
//		Of all possible paths, the service chooses the one having:
//			* an attempt linked to the final item if such a path exists,
//			* missing/not-started results located closer to the end of the path,
//			* preferring paths having less missing/not-started results,
//			* and having higher values of `attempt_id`.
//
//		If `as_team_id` is given, the attempts/results of the path are linked to the `as_team_id` group instead of
//		the current user's self group, the participant becomes the given team group.
//
//		Restrictions:
//
//			* if `as_team_id` is given, it should be a user's parent team group,
//			* at least one path should exist,
//
//			otherwise, the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Success response with the found item path
//			schema:
//				type: object
//				properties:
//					path:
//						type: array
//						items:
//							type: string
//							format: int64
//						example: ["1", "2", "3"]
//				required:
//					- path
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
func (srv *Service) getPathFromRoot(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID := service.ParticipantIDFromContext(httpRequest.Context())

	itemPaths := findItemPaths(srv.GetStore(httpRequest), participantID, itemID, 1)
	if itemPaths == nil {
		return service.ErrAPIInsufficientAccessRights
	}
	render.Respond(responseWriter, httpRequest, map[string]interface{}{"path": itemPaths[0].Path})
	return nil
}

// findItemPaths gets the paths from root items to the given item for the given participant.
//
// When {limit}=0, return all the paths.
func findItemPaths(store *database.DataStore, participantID, itemID int64, limit int) []ItemPath {
	var limitStatement string
	if limit > 0 {
		limitStatement = " LIMIT " + strconv.Itoa(limit)
	}

	participantAncestors := store.ActiveGroupAncestors().Where("child_group_id = ?", participantID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.ancestor_group_id").
		Select("groups.id, root_activity_id, root_skill_id")

	groupsManagedByParticipant := store.ActiveGroupAncestors().ManagedByGroup(participantID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id").
		Select("groups.id, root_activity_id, root_skill_id")

	groupsWithRootItems := participantAncestors.Union(groupsManagedByParticipant)

	visibleItems := store.Permissions().MatchingGroupAncestors(participantID).
		WherePermissionIsAtLeast("view", "info").
		Joins("JOIN items ON items.id = permissions.item_id").
		Select("items.id, requires_explicit_entry, MAX(can_view_generated_value) AS can_view_generated_value").
		Group("items.id")

	canViewContentIndex := store.PermissionsGranted().ViewIndexByName("content")

	var rawItemPaths []rawItemPath
	service.MustNotBeError(store.Raw(`
		WITH RECURSIVE paths (path, final_item_id, final_attempt_id, score, attempts, is_started, is_active) AS (
			WITH
				groups_with_root_items AS ?,
				visible_items AS ?,
				root_items AS (
					SELECT visible_items.id AS id FROM groups_with_root_items JOIN visible_items ON visible_items.id = root_activity_id
					UNION
					SELECT visible_items.id FROM groups_with_root_items JOIN visible_items ON visible_items.id = root_skill_id
				),
				item_ancestors AS (
					SELECT visible_items.id, requires_explicit_entry, can_view_generated_value
					FROM items_ancestors
					JOIN visible_items ON visible_items.id = items_ancestors.ancestor_item_id
					WHERE child_item_id = ?
					UNION
					SELECT id, requires_explicit_entry, can_view_generated_value FROM visible_items WHERE id = ?
				),
				root_ancestors AS (
					SELECT item_ancestors.id, requires_explicit_entry, can_view_generated_value
					FROM item_ancestors
					JOIN root_items ON root_items.id = item_ancestors.id
				)
			(SELECT CAST(root_ancestors.id AS CHAR(1024)), root_ancestors.id, attempts.id, results.started_at IS NULL,
			        CAST(LPAD(IFNULL(attempts.id, '!'), 20, 0) AS CHAR(1024)), results.started_at IS NOT NULL,
			        attempts.ended_at IS NULL AND NOW() < attempts.allows_submissions_until
			FROM root_ancestors
			LEFT JOIN attempts ON attempts.participant_id = ? AND
				(NOT root_ancestors.requires_explicit_entry OR attempts.root_item_id = root_ancestors.id)
			LEFT JOIN results ON results.participant_id = attempts.participant_id AND
				attempts.id = results.attempt_id AND results.item_id = root_ancestors.id
			WHERE root_ancestors.id = ? OR (
				root_ancestors.can_view_generated_value >= ? AND
				(NOT root_ancestors.requires_explicit_entry OR results.attempt_id IS NOT NULL) AND
				(results.started_at IS NOT NULL OR attempts.ended_at IS NULL AND NOW() < attempts.allows_submissions_until) AND
				(results.attempt_id IS NOT NULL OR attempts.id = 0)
			))
			UNION
			(SELECT CONCAT(paths.path, '/', item_ancestors.id), item_ancestors.id, attempts.id, (paths.score << 1) + (results.started_at IS NULL),
			        CONCAT(paths.attempts, '/', LPAD(IFNULL(attempts.id, '!'), 20, 0)),
			        paths.is_started AND results.started_at IS NOT NULL,
			        paths.is_active AND attempts.ended_at IS NULL AND NOW() < attempts.allows_submissions_until
			FROM paths
			JOIN items_items ON items_items.parent_item_id = paths.final_item_id
			JOIN item_ancestors ON item_ancestors.id = items_items.child_item_id
			LEFT JOIN attempts ON attempts.participant_id = ? AND
				(NOT item_ancestors.requires_explicit_entry OR attempts.root_item_id = item_ancestors.id) AND
				IF(attempts.root_item_id = item_ancestors.id, attempts.parent_attempt_id, attempts.id) = paths.final_attempt_id
			LEFT JOIN results ON results.participant_id = attempts.participant_id AND
				attempts.id = results.attempt_id AND results.item_id = item_ancestors.id
			WHERE paths.final_item_id <> ? AND (
				item_ancestors.id = ? OR (
					item_ancestors.can_view_generated_value >= ? AND
					(NOT item_ancestors.requires_explicit_entry OR results.attempt_id IS NOT NULL) AND
					(results.started_at IS NOT NULL OR (attempts.ended_at IS NULL AND NOW() < attempts.allows_submissions_until AND paths.is_active))
				)
			)
		))
		SELECT
			path, MAX(is_started) AS is_started,
			MAX(CONCAT(final_attempt_id IS NOT NULL, ':', LPAD(1000000000000000000000000000000000000000-score, 40, 0), ':', attempts)) AS sorting
		FROM paths
		WHERE paths.final_item_id = ?
		GROUP BY path
		ORDER BY sorting DESC
		`+limitStatement,
		groupsWithRootItems.SubQuery(), visibleItems.SubQuery(), itemID, itemID, participantID, itemID, canViewContentIndex,
		participantID, itemID, itemID, canViewContentIndex, itemID).
		Scan(&rawItemPaths).Error())

	if len(rawItemPaths) == 0 {
		return nil
	}

	itemPaths := make([]ItemPath, 0, len(rawItemPaths))
	for _, itemPathRow := range rawItemPaths {
		itemPaths = append(itemPaths, ItemPath{
			Path:      strings.Split(itemPathRow.Path, "/"),
			IsStarted: itemPathRow.IsStarted,
		})
	}

	return itemPaths
}
