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
	router.Use(auth.UserIDMiddleware(&srv.Config.Auth))

	router.Get("/current-user/invitations", service.AppHandler(srv.getInvitations).ServeHTTP)
	router.Post("/current-user/invitations/{group_id}/accept", service.AppHandler(srv.acceptInvitation).ServeHTTP)
	router.Post("/current-user/invitations/{group_id}/reject", service.AppHandler(srv.rejectInvitation).ServeHTTP)

	router.Post("/current-user/requests/{group_id}", service.AppHandler(srv.sendRequest).ServeHTTP)

	router.Get("/current-user/memberships", service.AppHandler(srv.getMemberships).ServeHTTP)
	router.Delete("/current-user/memberships/{group_id}", service.AppHandler(srv.leaveGroup).ServeHTTP)
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
	selfGroupID, err := user.SelfGroupID()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var results database.GroupGroupTransitionResults
	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		results, err = store.GroupGroups().Transition(
			map[userGroupRelationAction]database.GroupGroupTransitionAction{
				acceptInvitationAction:   database.UserAcceptsInvitation,
				rejectInvitationAction:   database.UserRefusesInvitation,
				createGroupRequestAction: database.UserCreatesRequest,
				leaveGroupAction:         database.UserLeavesGroup,
			}[action], groupID, []int64{selfGroupID}, user.UserID)
		return err
	}))

	return service.RenderGroupGroupTransitionResult(w, r, results[selfGroupID],
		action == createGroupRequestAction, action == leaveGroupAction)
}
