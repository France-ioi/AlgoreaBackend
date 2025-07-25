// Package currentuser provides the services related to the current user.
package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `currentuser`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/current-user", service.AppHandler(srv.getInfo).ServeHTTP)
	router.Put("/current-user", service.AppHandler(srv.update).ServeHTTP)
	router.Delete("/current-user", service.AppHandler(srv.delete).ServeHTTP)

	router.Get("/current-user/available-groups", service.AppHandler(srv.searchForAvailableGroups).ServeHTTP)
	router.Get("/current-user/check-login-id", service.AppHandler(srv.checkLoginID).ServeHTTP)

	router.Get("/current-user/group-invitations", service.AppHandler(srv.getGroupInvitations).ServeHTTP)
	router.Post("/current-user/group-invitations/{group_id}/accept", service.AppHandler(srv.acceptGroupInvitation).ServeHTTP)
	router.Post("/current-user/group-invitations/{group_id}/reject", service.AppHandler(srv.rejectGroupInvitation).ServeHTTP)

	router.Post("/current-user/group-requests/{group_id}", service.AppHandler(srv.createGroupJoinRequest).ServeHTTP)
	router.Post("/current-user/group-requests/{group_id}/withdraw", service.AppHandler(srv.withdrawGroupJoinRequest).ServeHTTP)
	router.Post("/current-user/group-leave-requests/{group_id}", service.AppHandler(srv.createGroupLeaveRequest).ServeHTTP)
	router.Post("/current-user/group-leave-requests/{group_id}/withdraw", service.AppHandler(srv.withdrawGroupLeaveRequest).ServeHTTP)

	router.Get("/current-user/managed-groups", service.AppHandler(srv.getManagedGroups).ServeHTTP)

	router.Get("/current-user/group-memberships", service.AppHandler(srv.getGroupMemberships).ServeHTTP)
	router.Post("/current-user/group-memberships/by-code", service.AppHandler(srv.joinGroupByCode).ServeHTTP)
	router.Delete("/current-user/group-memberships/{group_id}", service.AppHandler(srv.leaveGroup).ServeHTTP)
	router.Get("/current-user/group-memberships-history", service.AppHandler(srv.getGroupMembershipsHistory).ServeHTTP)

	routerWithParticipant := router.With(service.ParticipantMiddleware(srv.Base))
	routerWithParticipant.Get("/current-user/group-memberships/activities", service.AppHandler(srv.getRootActivities).ServeHTTP)
	routerWithParticipant.Get("/current-user/group-memberships/skills", service.AppHandler(srv.getRootSkills).ServeHTTP)

	router.Put("/current-user/notifications-read-at", service.AppHandler(srv.updateNotificationsReadAt).ServeHTTP)
	router.Put("/current-user/refresh", service.AppHandler(srv.refresh).ServeHTTP)

	router.Get("/current-user/full-dump", service.AppHandler(srv.getFullDump).ServeHTTP)
	router.Get("/current-user/dump", service.AppHandler(srv.getDump).ServeHTTP)
}

type userGroupRelationAction string

const (
	acceptInvitationAction               userGroupRelationAction = "acceptInvitation"
	rejectInvitationAction               userGroupRelationAction = "rejectInvitation"
	createGroupJoinRequestAction         userGroupRelationAction = "createJoinRequest"
	createAcceptedGroupJoinRequestAction userGroupRelationAction = "createAcceptedJoinRequest"
	createGroupLeaveRequestAction        userGroupRelationAction = "createLeaveRequest"
	withdrawGroupJoinRequestAction       userGroupRelationAction = "withdrawJoinRequest"
	withdrawGroupLeaveRequestAction      userGroupRelationAction = "withdrawLeaveRequest"
	leaveGroupAction                     userGroupRelationAction = "leaveGroup"
	joinGroupByCodeAction                userGroupRelationAction = "joinGroupByCode"
)

const team = "Team"

func (srv *Service) performGroupRelationAction(
	responseWriter http.ResponseWriter, httpRequest *http.Request, action userGroupRelationAction,
) error {
	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)

	var result database.GroupGroupTransitionResult
	var approvalsToRequest database.GroupApprovals
	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		if action == leaveGroupAction {
			var found bool
			found, err = store.Groups().ByID(groupID).
				Joins(`
				JOIN groups_groups_active
					ON groups_groups_active.parent_group_id = groups.id AND
						 groups_groups_active.child_group_id = ?`, user.GroupID).
				Where(`
					(groups_groups_active.lock_membership_approved AND NOW() < groups.require_lock_membership_approval_until) OR
					groups.frozen_membership OR groups.type = 'Base'`).HasRows()
			service.MustNotBeError(err)
			if found {
				return service.ErrForbidden(errors.New("user deletion is locked for this group")) // rollback
			}
		}

		if action == createGroupLeaveRequestAction {
			var found bool
			found, err = store.Groups().ByID(groupID).
				Joins(`
				JOIN groups_groups_active
					ON groups_groups_active.parent_group_id = groups.id AND
					   groups_groups_active.lock_membership_approved AND
					   groups_groups_active.child_group_id = ?`, user.GroupID).
				Where("NOW() < require_lock_membership_approval_until AND NOT groups.frozen_membership").HasRows()
			service.MustNotBeError(err)
			if !found {
				return service.ErrForbidden(
					errors.New(
						"user is not a member of the group or the group doesn't require approval for leaving or its membership is frozen")) // rollback
			}
		}

		var approvals database.GroupApprovals
		if map[userGroupRelationAction]bool{createGroupJoinRequestAction: true, acceptInvitationAction: true}[action] {
			if user.IsTempUser {
				return service.ErrAPIInsufficientAccessRights // rollback
			}

			approvals.FromString(httpRequest.URL.Query().Get("approvals"))
		}

		result, approvalsToRequest, err = performUserGroupRelationAction(action, store, user, groupID, approvals)
		return err
	})

	service.MustNotBeError(err)

	return RenderGroupGroupTransitionResult(responseWriter, httpRequest, result, approvalsToRequest, action)
}

func performUserGroupRelationAction(action userGroupRelationAction, store *database.DataStore, user *database.User,
	groupID int64, approvals database.GroupApprovals,
) (database.GroupGroupTransitionResult, database.GroupApprovals, error) {
	var err error

	if action == createGroupJoinRequestAction {
		var found bool
		found, err = store.Groups().ManagedBy(user).Where("can_manage != 'none'").Where("groups.id = ?", groupID).HasRows()
		service.MustNotBeError(err)
		if found {
			action = createAcceptedGroupJoinRequestAction
		}
	}
	if map[userGroupRelationAction]bool{
		createGroupJoinRequestAction: true, acceptInvitationAction: true, createAcceptedGroupJoinRequestAction: true,
	}[action] {
		err = checkPreconditionsForGroupRequests(store, user, groupID, action)
		if err != nil {
			return "", database.GroupApprovals{}, err
		}
	}
	if action == leaveGroupAction {
		var groupType string
		groupStore := store.Groups()
		service.MustNotBeError(groupStore.ByID(groupID).WithExclusiveWriteLock().PluckFirst("type", &groupType).Error())
		if groupType == team {
			var ok bool
			ok, err = groupStore.CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(groupID, user.GroupID, true, true)
			service.MustNotBeError(err)
			if !ok {
				return "", database.GroupApprovals{}, service.ErrUnprocessableEntity(errors.New("entry conditions would not be satisfied"))
			}
		}
	}
	var results database.GroupGroupTransitionResults
	var approvalsToRequest map[int64]database.GroupApprovals
	results, approvalsToRequest, err = store.GroupGroups().Transition(
		map[userGroupRelationAction]database.GroupGroupTransitionAction{
			acceptInvitationAction:               database.UserAcceptsInvitation,
			rejectInvitationAction:               database.UserRefusesInvitation,
			createGroupJoinRequestAction:         database.UserCreatesJoinRequest,
			createAcceptedGroupJoinRequestAction: database.UserCreatesAcceptedJoinRequest,
			withdrawGroupJoinRequestAction:       database.UserCancelsJoinRequest,
			withdrawGroupLeaveRequestAction:      database.UserCancelsLeaveRequest,
			leaveGroupAction:                     database.UserLeavesGroup,
			createGroupLeaveRequestAction:        database.UserCreatesLeaveRequest,
		}[action], groupID, []int64{user.GroupID}, map[int64]database.GroupApprovals{user.GroupID: approvals}, user.GroupID)
	service.MustNotBeError(err)
	return results[user.GroupID], approvalsToRequest[user.GroupID], nil
}

func checkPreconditionsForGroupRequests(store *database.DataStore, user *database.User,
	groupID int64, action userGroupRelationAction,
) error {
	// The group should exist (and optionally should have `is_public` = 1)
	query := store.Groups().ByID(groupID).
		Where("type != 'User'").Select("type, frozen_membership").WithExclusiveWriteLock()
	if action == createGroupJoinRequestAction {
		query = query.Where("is_public")
	}
	var groupInfo struct {
		Type             string
		FrozenMembership bool
	}
	err := query.Take(&groupInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	if groupInfo.FrozenMembership {
		return service.ErrUnprocessableEntity(errors.New("group membership is frozen"))
	}

	// If the group is a team, ensure that the current user is not a member of
	// another team having attempts for the same item requiring explicit entry.
	if groupInfo.Type == team {
		found, err := store.CheckIfTeamParticipationsConflictWithExistingUserMemberships(groupID, user.GroupID, true)
		service.MustNotBeError(err)
		if found {
			return service.ErrUnprocessableEntity(errors.New("team's participations are in conflict with the user's participations"))
		}
		var ok bool
		ok, err = store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(groupID, user.GroupID, true, true)
		service.MustNotBeError(err)
		if !ok {
			return service.ErrUnprocessableEntity(errors.New("entry conditions would not be satisfied"))
		}
	}

	return nil
}
