package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /user-batches/{group_prefix}/{custom_prefix} groups userBatchRemove
// ---
// summary: Remove a user batch
// description:
//   Lets an admin remove user batches and all users having "{group_prefix}_{custom_prefix}_" as login prefix.
//
//   On success the service
//
//     * requests the login module to delete the users with {group_prefix}_{custom_prefix}_" as prefix
//       (/platform_api/accounts_manager/delete with prefix parameter);
//
//     * deletes all users with "{group_prefix}_{custom_prefix}_" as prefix
//       (ignoring the membership locks on groups the authenticated user manage (but not others!));
//
//     * deletes the user batch entry.
//
//
//   The authenticated user should be a manager of the `group_prefix`'s group (or its ancestor)
//   with `can_manage` >= 'memberships', otherwise the 'forbidden' error is returned.
//
//
//   If there are users with locked membership in groups the current user cannot manage,
//   the 'unprocessable entity' error is returned.
// parameters:
// - name: group_prefix
//   in: path
//   type: string
//   required: true
// - name: custom_prefix
//   in: path
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/deletedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) removeUserBatch(w http.ResponseWriter, r *http.Request) service.APIError {
	groupPrefix := chi.URLParam(r, "group_prefix")
	customPrefix := chi.URLParam(r, "custom_prefix")

	user := srv.GetUser(r)
	managedByUser := srv.Store.ActiveGroupAncestors().ManagedByUser(user).
		Where("can_manage != 'none'").
		Select("groups_ancestors_active.child_group_id AS id")

	// The user batch should exist and the current user should be a manager of the group
	// linked to the group_prefix
	found, err := srv.Store.UserBatches().
		Joins("JOIN user_batch_prefixes USING(group_prefix)").
		Where("group_prefix = ?", groupPrefix).
		Where("custom_prefix = ?", customPrefix).
		Where("user_batch_prefixes.group_id IN (?)", managedByUser.QueryExpr()).
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	// There should not be users with locked membership in the groups the current user cannot manage
	found, err = srv.Store.Users().
		Joins(`
			JOIN groups_groups_active
				ON groups_groups_active.child_group_id = users.group_id`).
		Joins("JOIN `groups` AS parent_group ON parent_group_id = groups_groups_active.parent_group_id").
		Where("NOW() < parent_group.require_lock_membership_approval_until AND groups_groups_active.lock_membership_approved").
		Where("login LIKE CONCAT(?, '\\_', ?, '\\_%')", groupPrefix, customPrefix).
		Where("parent_group.id NOT IN(?)", managedByUser.QueryExpr()).
		HasRows()
	service.MustNotBeError(err)
	if found {
		return service.ErrUnprocessableEntity(errors.New("there are users with locked membership"))
	}

	service.MustNotBeError(loginmodule.NewClient(srv.Config.Auth.LoginModuleURL).
		DeleteUsers(r.Context(), srv.Config.Auth.ClientID, srv.Config.Auth.ClientSecret, groupPrefix+"_"+customPrefix+"_"))

	service.MustNotBeError(srv.Store.Users().DeleteWithTrapsByScope(func(store *database.DataStore) *database.DB {
		return store.Users().Where("login LIKE CONCAT(?, '\\_', ?, '\\_%')", groupPrefix, customPrefix)
	}))
	service.MustNotBeError(
		srv.Store.UserBatches().
			Where("group_prefix = ?", groupPrefix).
			Where("custom_prefix = ?", customPrefix).Delete().Error())

	service.MustNotBeError(render.Render(w, r, service.DeletionSuccess(nil)))
	return service.NoError
}
