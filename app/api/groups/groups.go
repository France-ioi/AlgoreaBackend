// Package groups provides API services to manage groups.
package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// Service is the mount point for services related to `groups`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))
	router.Post("/groups", service.AppHandler(srv.createGroup).ServeHTTP)
	router.Get("/groups/possible-subgroups", service.AppHandler(srv.searchForPossibleSubgroups).ServeHTTP)
	router.Get("/groups/roots", service.AppHandler(srv.getRoots).ServeHTTP)
	router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
	router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
	router.Delete("/groups/{group_id}", service.AppHandler(srv.deleteGroup).ServeHTTP)
	router.Get("/groups/{source_group_id}/permissions/{group_id}/{item_id}",
		service.AppHandler(srv.getPermissions).ServeHTTP)
	router.Get("/groups/{group_id}/granted_permissions",
		service.AppHandler(srv.getGrantedPermissions).ServeHTTP)
	router.Put("/groups/{source_group_id}/permissions/{group_id}/{item_id}",
		service.AppHandler(srv.updatePermissions).ServeHTTP)

	router.Post("/groups/{group_id}/code", service.AppHandler(srv.createCode).ServeHTTP)
	router.Delete("/groups/{group_id}/code", service.AppHandler(srv.removeCode).ServeHTTP)
	router.Get("/groups/is-code-valid", service.AppHandler(srv.checkCode).ServeHTTP)

	router.Get("/groups/{group_id}/navigation", service.AppHandler(srv.getNavigation).ServeHTTP)
	router.Get("/groups/{group_id}/path-from-root", service.AppHandler(srv.getPathFromRoot).ServeHTTP)
	router.Get("/groups/{ids:(\\d+/)+}breadcrumbs", service.AppHandler(srv.getBreadcrumbs).ServeHTTP)
	router.Get("/groups/{group_id}/children", service.AppHandler(srv.getChildren).ServeHTTP)
	router.Get("/groups/{group_id}/team-descendants", service.AppHandler(srv.getTeamDescendants).ServeHTTP)
	router.Get("/groups/{group_id}/user-descendants", service.AppHandler(srv.getUserDescendants).ServeHTTP)
	router.Get("/groups/{group_id}/members", service.AppHandler(srv.getMembers).ServeHTTP)
	router.Delete("/groups/{group_id}/members", service.AppHandler(srv.removeMembers).ServeHTTP)

	router.Get("/groups/{group_id}/managers", service.AppHandler(srv.getManagers).ServeHTTP)
	router.Post("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.createGroupManager).ServeHTTP)
	router.Put("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.updateGroupManager).ServeHTTP)
	router.Delete("/groups/{group_id}/managers/{manager_id}", service.AppHandler(srv.removeGroupManager).ServeHTTP)

	router.Get("/groups/{group_id}/parents", service.AppHandler(srv.getParents).ServeHTTP)

	router.Get("/groups/{group_id}/requests", service.AppHandler(srv.getRequests).ServeHTTP)
	router.Get("/groups/user-requests", service.AppHandler(srv.getUserRequests).ServeHTTP)
	router.Get("/groups/{group_id}/group-progress", service.AppHandler(srv.getGroupProgress).ServeHTTP)
	router.Get("/groups/{group_id}/group-progress-csv", service.AppHandler(srv.getGroupProgressCSV).ServeHTTP)
	router.Get("/groups/{group_id}/team-progress", service.AppHandler(srv.getTeamProgress).ServeHTTP)
	router.Get("/groups/{group_id}/team-progress-csv", service.AppHandler(srv.getTeamProgressCSV).ServeHTTP)
	router.Get("/groups/{group_id}/user-progress", service.AppHandler(srv.getUserProgress).ServeHTTP)
	router.Get("/groups/{group_id}/user-progress-csv", service.AppHandler(srv.getUserProgressCSV).ServeHTTP)
	router.With(service.ParticipantMiddleware(srv.Base)).
		Get("/items/{item_id}/participant-progress", service.AppHandler(srv.getParticipantProgress).ServeHTTP)
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

func checkThatUserCanManageTheGroup(store *database.DataStore, user *database.User, groupID int64) error {
	found, err := store.GroupAncestors().ManagedByUser(user).
		Where("groups_ancestors.child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}
	return nil
}

func checkThatUserCanManageTheGroupMemberships(store *database.DataStore, user *database.User, groupID int64) error {
	found, err := store.GroupAncestors().ManagedByUser(user).
		Joins("JOIN `groups` ON groups.id = groups_ancestors.child_group_id").
		Where("groups_ancestors.child_group_id = ?", groupID).
		Where("group_managers.can_manage != 'none'").
		Where("groups.type != 'User'").HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}
	return nil
}

type createOrDeleteRelation bool

const (
	createRelation createOrDeleteRelation = true
	deleteRelation createOrDeleteRelation = false
)

const (
	groupTypeTeam = "Team"
	groupTypeUser = "User"
)

func checkThatUserHasRightsForDirectRelation(
	store *database.DataStore, user *database.User,
	parentGroupID, childGroupID int64, createOrDelete createOrDeleteRelation,
) error {
	groupStore := store.Groups()

	var groupData []struct {
		ID   int64
		Type string
	}

	query := groupStore.ManagedBy(user).
		WithCustomWriteLocks(golang.NewSet("groups"), golang.NewSet[string]()).
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

	//nolint:mnd // one row for the parent group, one for the child group
	if len(groupData) < 2 {
		return service.ErrAPIInsufficientAccessRights
	}

	for _, groupRow := range groupData {
		if (groupRow.ID == parentGroupID && map[string]bool{"User": true, "Team": true}[groupRow.Type]) ||
			(groupRow.ID == childGroupID &&
				map[string]bool{"Base": true, "User": true}[groupRow.Type]) {
			return service.ErrAPIInsufficientAccessRights
		}
	}
	return nil
}

type bulkMembershipAction string

const (
	acceptJoinRequestsAction  bulkMembershipAction = "acceptJoinRequests"
	rejectJoinRequestsAction  bulkMembershipAction = "rejectJoinRequests"
	acceptLeaveRequestsAction bulkMembershipAction = "acceptLeaveRequests"
	rejectLeaveRequestsAction bulkMembershipAction = "rejectLeaveRequests"
	withdrawInvitationsAction bulkMembershipAction = "withdrawInvitations"
)

const (
	inAnotherTeam = "in_another_team"
	notFound      = "not_found"
	team          = "Team"
)

func (srv *Service) performBulkMembershipAction(
	responseWriter http.ResponseWriter, httpRequest *http.Request, action bulkMembershipAction,
) error {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupIDs, err := service.ResolveURLQueryGetInt64SliceField(httpRequest, "group_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	var results database.GroupGroupTransitionResults
	if len(groupIDs) > 0 {
		err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
			groupType, err1 := checkPreconditionsForBulkMembershipAction(action, user, store, parentGroupID, groupIDs)
			if err1 != nil {
				return err1 // rollback
			}

			results = performBulkMembershipActionTransition(store, action, parentGroupID, groupIDs, groupType, user)
			return nil
		})
	}

	service.MustNotBeError(err)

	renderGroupGroupTransitionResults(responseWriter, httpRequest, results)
	return nil
}

func performBulkMembershipActionTransition(store *database.DataStore, action bulkMembershipAction, parentGroupID int64,
	groupIDs []int64, groupType string, user *database.User,
) database.GroupGroupTransitionResults {
	groupID := groupIDs[0]
	if groupType == team {
		if action == acceptJoinRequestsAction && isOtherTeamMember(store, parentGroupID, groupID) {
			return database.GroupGroupTransitionResults{groupID: inAnotherTeam}
		}

		if map[bulkMembershipAction]bool{acceptJoinRequestsAction: true, acceptLeaveRequestsAction: true}[action] {
			ok, err := store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(
				parentGroupID, groupID, action == acceptJoinRequestsAction, true)
			service.MustNotBeError(err)
			if !ok {
				return database.GroupGroupTransitionResults{groupID: "entry_condition_failed"}
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
	service.MustNotBeError(err)

	return results
}

func checkPreconditionsForBulkMembershipAction(action bulkMembershipAction, user *database.User, store *database.DataStore,
	parentGroupID int64, groupIDs []int64,
) (groupType string, err error) {
	if err := checkThatUserCanManageTheGroupMemberships(store, user, parentGroupID); err != nil {
		return "", err
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
	return groupInfo.Type, nil
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
