package currentuser

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `items`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))

	router.Get("/current-user", service.AppHandler(srv.getInfo).ServeHTTP)

	router.Get("/current-user/available-groups", service.AppHandler(srv.searchForAvailableGroups).ServeHTTP)

	router.Get("/current-user/group-invitations", service.AppHandler(srv.getGroupInvitations).ServeHTTP)
	router.Post("/current-user/group-invitations/{group_id}/accept", service.AppHandler(srv.acceptGroupInvitation).ServeHTTP)
	router.Post("/current-user/group-invitations/{group_id}/reject", service.AppHandler(srv.rejectGroupInvitation).ServeHTTP)

	router.Post("/current-user/group-requests/{group_id}", service.AppHandler(srv.sendGroupRequest).ServeHTTP)

	router.Get("/current-user/group-memberships", service.AppHandler(srv.getGroupMemberships).ServeHTTP)
	router.Delete("/current-user/group-memberships/{group_id}", service.AppHandler(srv.leaveGroup).ServeHTTP)
	router.Get("/current-user/group-memberships-history", service.AppHandler(srv.getGroupMembershipsHistory).ServeHTTP)

	router.Put("/current-user/notification-read-date", service.AppHandler(srv.updateNotificationReadDate).ServeHTTP)
	router.Put("/current-user/refresh", service.AppHandler(srv.refresh).ServeHTTP)

	router.Get("/current-user/dump", service.AppHandler(srv.getDump).ServeHTTP)
}

type userGroupRelationAction string

const (
	acceptInvitationAction   userGroupRelationAction = "acceptInvitation"
	rejectInvitationAction   userGroupRelationAction = "rejectInvitation"
	createGroupRequestAction userGroupRelationAction = "createRequest"
	leaveGroupAction         userGroupRelationAction = "leaveGroup"
)

func (srv *Service) performGroupRelationAction(w http.ResponseWriter, r *http.Request, action userGroupRelationAction) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	if user.SelfGroupID == nil {
		return service.InsufficientAccessRightsError
	}

	if action == createGroupRequestAction {
		var found bool
		found, err = srv.Store.Groups().ByID(groupID).Where("bFreeAccess").HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.InsufficientAccessRightsError
		}
	}

	var results database.GroupGroupTransitionResults
	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		results, err = store.GroupGroups().Transition(
			map[userGroupRelationAction]database.GroupGroupTransitionAction{
				acceptInvitationAction:   database.UserAcceptsInvitation,
				rejectInvitationAction:   database.UserRefusesInvitation,
				createGroupRequestAction: database.UserCreatesRequest,
				leaveGroupAction:         database.UserLeavesGroup,
			}[action], groupID, []int64{*user.SelfGroupID}, user.ID)
		return err
	}))

	return service.RenderGroupGroupTransitionResult(w, r, results[*user.SelfGroupID],
		action == createGroupRequestAction, action == leaveGroupAction)
}
