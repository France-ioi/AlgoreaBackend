package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `groups`
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))
	router.Post("/groups", service.AppHandler(srv.createGroup).ServeHTTP)
	router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
	router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
	router.Put("/groups/{source_group_id}/permissions/{group_id}/{item_id}",
		service.AppHandler(srv.updatePermissions).ServeHTTP)

	router.Post("/groups/{group_id}/code", service.AppHandler(srv.createCode).ServeHTTP)
	router.Delete("/groups/{group_id}/code", service.AppHandler(srv.removeCode).ServeHTTP)
	router.Get("/groups/is-code-valid", service.AppHandler(srv.checkCode).ServeHTTP)

	router.Get("/groups/{group_id}/children", service.AppHandler(srv.getChildren).ServeHTTP)
	router.Get("/groups/{group_id}/team-descendants", service.AppHandler(srv.getTeamDescendants).ServeHTTP)
	router.Get("/groups/{group_id}/user-descendants", service.AppHandler(srv.getUserDescendants).ServeHTTP)
	router.Get("/groups/{group_id}/members", service.AppHandler(srv.getMembers).ServeHTTP)
	router.Delete("/groups/{group_id}/members", service.AppHandler(srv.removeMembers).ServeHTTP)

	router.Get("/groups/{group_id}/managers", service.AppHandler(srv.getManagers).ServeHTTP)
	router.Post("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.createGroupManager).ServeHTTP)
	router.Put("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.updateGroupManager).ServeHTTP)
	router.Delete("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.removeGroupManager).ServeHTTP)

	router.Get("/groups/{group_id}/requests", service.AppHandler(srv.getRequests).ServeHTTP)
	router.Get("/groups/user-requests", service.AppHandler(srv.getUserRequests).ServeHTTP)
	router.Get("/groups/{group_id}/group-progress", service.AppHandler(srv.getGroupProgress).ServeHTTP)
	router.Get("/groups/{group_id}/team-progress", service.AppHandler(srv.getTeamProgress).ServeHTTP)
	router.Get("/groups/{group_id}/user-progress", service.AppHandler(srv.getUserProgress).ServeHTTP)
	router.Post("/groups/{parent_group_id}/join-requests/accept", service.AppHandler(srv.acceptJoinRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/join-requests/reject", service.AppHandler(srv.rejectJoinRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/leave-requests/accept", service.AppHandler(srv.acceptLeaveRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/leave-requests/reject", service.AppHandler(srv.rejectLeaveRequests).ServeHTTP)

	router.Post("/groups/{parent_group_id}/invitations", service.AppHandler(srv.createGroupInvitations).ServeHTTP)
	router.Post("/groups/{parent_group_id}/invitations/withdraw", service.AppHandler(srv.withdrawInvitations).ServeHTTP)

	router.Post("/groups/{parent_group_id}/relations/{child_group_id}", service.AppHandler(srv.addChild).ServeHTTP)
	router.Delete("/groups/{parent_group_id}/relations/{child_group_id}", service.AppHandler(srv.removeChild).ServeHTTP)

	router.Get("/current-user/teams/by-item/{item_id}", service.AppHandler(srv.getCurrentUserTeamByItem).ServeHTTP)
	router.Post("/user-batches", service.AppHandler(srv.createUserBatch).ServeHTTP)
	router.Get("/user-batches/by-group/{group_id}", service.AppHandler(srv.getUserBatches).ServeHTTP)
	router.Delete("/user-batches/{group_prefix}/{custom_prefix}", service.AppHandler(srv.removeUserBatch).ServeHTTP)
	router.Get("/groups/{group_id}/user-batch-prefixes", service.AppHandler(srv.getUserBatchPrefixes).ServeHTTP)
}

func checkThatUserCanManageTheGroup(store *database.DataStore, user *database.User, groupID int64) service.APIError {
	found, err := store.GroupAncestors().ManagedByUser(user).
		Where("groups_ancestors.child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func checkThatUserCanManageTheGroupMemberships(store *database.DataStore, user *database.User, groupID int64) service.APIError {
	found, err := store.GroupAncestors().ManagedByUser(user).
		Joins("JOIN `groups` ON groups.id = groups_ancestors.child_group_id").
		Where("groups_ancestors.child_group_id = ?", groupID).
		Where("group_managers.can_manage != 'none'").
		Where("groups.type != 'User'").HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

type createOrDeleteRelation bool

const (
	createRelation createOrDeleteRelation = true
	deleteRelation createOrDeleteRelation = false
)

func checkThatUserHasRightsForDirectRelation(
	store *database.DataStore, user *database.User,
	parentGroupID, childGroupID int64, createOrDelete createOrDeleteRelation) service.APIError {

	groupStore := store.Groups()

	var groupData []struct {
		ID   int64
		Type string
	}

	query := groupStore.ManagedBy(user).
		WithWriteLock().
		Select("groups.id, type").
		Where("groups.id IN(?, ?)", parentGroupID, childGroupID).
		Where("IF(groups.id = ?, group_managers.can_manage != 'none', 1)", parentGroupID)

	if createOrDelete == createRelation {
		query = query.Where("IF(groups.id = ?, group_managers.can_manage = 'memberships_and_group', 1)", childGroupID)
	}

	err := query.
		Group("groups.id").
		Scan(&groupData).Error()
	service.MustNotBeError(err)

	if len(groupData) < 2 {
		return service.InsufficientAccessRightsError
	}

	for _, groupRow := range groupData {
		if (groupRow.ID == parentGroupID && map[string]bool{"User": true, "Team": true}[groupRow.Type]) ||
			(groupRow.ID == childGroupID &&
				map[string]bool{"Base": true, "User": true}[groupRow.Type]) {
			return service.InsufficientAccessRightsError
		}
	}
	return service.NoError
}

type bulkMembershipAction string

const (
	acceptJoinRequestsAction  bulkMembershipAction = "acceptJoinRequests"
	rejectJoinRequestsAction  bulkMembershipAction = "rejectJoinRequests"
	acceptLeaveRequestsAction bulkMembershipAction = "acceptLeaveRequests"
	rejectLeaveRequestsAction bulkMembershipAction = "rejectLeaveRequests"
	withdrawInvitationsAction bulkMembershipAction = "withdrawInvitations"
)

const inAnotherTeam = "in_another_team"
const notFound = "not_found"
const team = "Team"

func (srv *Service) performBulkMembershipAction(w http.ResponseWriter, r *http.Request,
	action bulkMembershipAction) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "group_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	apiError := service.NoError
	var results database.GroupGroupTransitionResults
	if len(groupIDs) > 0 {
		err = srv.Store.InTransaction(func(store *database.DataStore) error {
			var groupType string
			groupType, apiError = checkPreconditionsForBulkMembershipAction(action, user, store, parentGroupID, groupIDs)
			if apiError != service.NoError {
				return apiError.Error // rollback
			}

			results, err = performBulkMembershipActionTransition(store, action, parentGroupID, groupIDs, groupType, user)
			return err
		})
	}

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	renderGroupGroupTransitionResults(w, r, results)
	return service.NoError
}

func performBulkMembershipActionTransition(store *database.DataStore, action bulkMembershipAction, parentGroupID int64,
	groupIDs []int64, groupType string, user *database.User) (database.GroupGroupTransitionResults, error) {
	groupID := groupIDs[0]
	if groupType == team {
		if action == acceptJoinRequestsAction && isOtherTeamMember(store, parentGroupID, groupID) {
			return database.GroupGroupTransitionResults{groupID: inAnotherTeam}, nil
		}

		if map[bulkMembershipAction]bool{acceptJoinRequestsAction: true, acceptLeaveRequestsAction: true}[action] {
			ok, err := store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(
				parentGroupID, groupID, action == acceptJoinRequestsAction, true)
			service.MustNotBeError(err)
			if !ok {
				return database.GroupGroupTransitionResults{groupID: "entry_condition_failed"}, nil
			}
		}
	}

	results, _, err := store.GroupGroups().Transition(
		map[bulkMembershipAction]database.GroupGroupTransitionAction{
			acceptJoinRequestsAction:  database.AdminAcceptsJoinRequest,
			rejectJoinRequestsAction:  database.AdminRefusesJoinRequest,
			withdrawInvitationsAction: database.AdminWithdrawsInvitation,
			acceptLeaveRequestsAction: database.AdminAcceptsLeaveRequest,
			rejectLeaveRequestsAction: database.AdminRefusesLeaveRequest,
		}[action], parentGroupID, groupIDs, nil, user.GroupID)
	return results, err
}

func checkPreconditionsForBulkMembershipAction(action bulkMembershipAction, user *database.User, store *database.DataStore,
	parentGroupID int64, groupIDs []int64) (groupType string, apiError service.APIError) {
	if apiError = checkThatUserCanManageTheGroupMemberships(store, user, parentGroupID); apiError != service.NoError {
		return "", apiError
	}

	var groupInfo struct {
		Type             string
		FrozenMembership bool
	}
	if action == acceptJoinRequestsAction || action == acceptLeaveRequestsAction {
		service.MustNotBeError(
			store.Groups().ByID(parentGroupID).Select("frozen_membership, type").Scan(&groupInfo).Error())
		if groupInfo.FrozenMembership {
			return groupInfo.Type, service.ErrForbidden(errors.New("group membership is frozen"))
		}
		if groupInfo.Type == team && len(groupIDs) > 1 {
			return groupInfo.Type, service.ErrInvalidRequest(
				errors.New("there should be no more than one id in group_ids when the parent group is a team"))
		}
	}
	return groupInfo.Type, service.NoError
}

type descendantParent struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`

	LinkedGroupID int64 `json:"-"`
}

func isOtherTeamMember(store *database.DataStore, parentGroupID, userID int64) bool {
	return len(getOtherTeamsMembers(store, parentGroupID, []int64{userID})) > 0
}
