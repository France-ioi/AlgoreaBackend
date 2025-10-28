package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /items/{dependent_item_id}/prerequisites/{prerequisite_item_id}/apply items itemDependencyApply
//
//	---
//	summary: (re)Apply a specific existing item-dependency rule on existing results
//	description: Applies the rule, i.e. grants the content access, for all existing participants which meet
//						 the condition defined by this dependency.
//						 The action doesn't affect access rights of those who doesn't meet the condition anymore.
//
//
//						 * The item dependency between `{dependent_item_id}` and `{prerequisite_item_id}`
//							 must exist with `grant_content_view` = 1, otherwise the 'not found' error is returned.
//
//						 * The current-user must have `can_edit` = 'all' and `can_grant_view` >= 'content' on the `{dependent_item_id}`,
//							 otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: dependent_item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: prerequisite_item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) applyDependency(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	dependentItemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "dependent_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	prerequisiteItemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "prerequisite_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)

	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		var found bool
		found, err = store.ItemDependencies().
			Where("dependent_item_id = ?", dependentItemID).
			Where("item_id = ?", prerequisiteItemID).
			Where("grant_content_view").WithExclusiveWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrNotFound(errors.New("no such dependency")) // rollback
		}
		found, err = store.Permissions().AggregatedPermissionsForItemsOnWhichGroupHasPermission(user.GroupID, "edit", "all").
			HavingMaxPermissionAtLeast("grant_view", "content").
			Where("item_id = ?", dependentItemID).WithExclusiveWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights // rollback
		}
		canViewContentIndex := store.PermissionsGranted().ViewIndexByName("content")
		result := store.Exec(`
			INSERT INTO permissions_granted
				(group_id, item_id, source_group_id, origin, can_view, can_enter_from, latest_update_at)
				SELECT
					results.participant_id,
					item_dependencies.dependent_item_id AS item_id,
					results.participant_id,
					'item_unlocking',
					IF(items.requires_explicit_entry, 'none', 'content'),
					IF(items.requires_explicit_entry, NOW(), '9999-12-31 23:59:59'),
					NOW()
				FROM item_dependencies
				JOIN results ON results.item_id = item_dependencies.item_id AND item_dependencies.score <= results.score_computed
				JOIN items ON items.id = item_dependencies.dependent_item_id
				WHERE item_dependencies.item_id = ? AND item_dependencies.dependent_item_id = ? AND
				      item_dependencies.grant_content_view
			ON DUPLICATE KEY UPDATE
				latest_update_at = IF(
					VALUES(can_view) = 'content' AND can_view_value < ? OR
					VALUES(can_enter_from) <> '9999-12-31 23:59:59' AND can_enter_from > VALUES(can_enter_from) OR
					VALUES(can_enter_from) <> '9999-12-31 23:59:59' AND can_enter_until <> '9999-12-31 23:59:59',
					NOW(), latest_update_at),
				can_view = IF(VALUES(can_view) = 'content' AND can_view_value < ?, 'content', can_view),
				can_enter_from = IF(
					VALUES(can_enter_from) <> '9999-12-31 23:59:59' AND can_enter_from > VALUES(can_enter_from),
					VALUES(can_enter_from), can_enter_from)`,
			prerequisiteItemID, dependentItemID, canViewContentIndex, canViewContentIndex)

		service.MustNotBeError(result.Error())
		groupsUnlocked := result.RowsAffected()
		// If items have been unlocked, need to recompute access
		if groupsUnlocked > 0 {
			// generate permissions_generated from permissions_granted
			store.SchedulePermissionsPropagation()
			// we should compute results again as new permissions were set and
			// triggers on permissions_generated likely marked some results as 'to_be_propagated'
			store.ScheduleResultsPropagation()
		}
		return err
	})

	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess[*struct{}](nil)))
	return nil
}
