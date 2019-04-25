package groups

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `groups`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserIDMiddleware(&srv.Config.Auth))
	router.Get("/groups/", service.AppHandler(srv.getAll).ServeHTTP)
	router.Get("/groups/{group_id}/recent_activity", service.AppHandler(srv.getRecentActivity).ServeHTTP)
	router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
	router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
	router.Post("/groups/{group_id}/change_password", service.AppHandler(srv.changePassword).ServeHTTP)
	router.Get("/groups/{group_id}/children", service.AppHandler(srv.getChildren).ServeHTTP)

	router.Get("/groups/{group_id}/requests", service.AppHandler(srv.getRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/requests/accept", service.AppHandler(srv.acceptRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/requests/reject", service.AppHandler(srv.rejectRequests).ServeHTTP)

	router.Post("/groups/{parent_group_id}/invitations", service.AppHandler(srv.inviteUsers).ServeHTTP)

	router.Post("/groups/{parent_group_id}/relations/{child_group_id}", service.AppHandler(srv.addChild).ServeHTTP)
	router.Delete("/groups/{parent_group_id}/relations/{child_group_id}", service.AppHandler(srv.removeChild).ServeHTTP)
}

func checkThatUserOwnsTheGroup(store *database.DataStore, user *database.User, groupID int64) service.APIError {
	var count int64
	if err := store.GroupAncestors().OwnedByUser(user).
		Where("idGroupChild = ?", groupID).Count(&count).Error(); err != nil {
		if err == database.ErrUserNotFound {
			return service.InsufficientAccessRightsError
		}
		return service.ErrUnexpected(err)
	}
	if count == 0 {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func checkThatUserHasRightsForDirectRelation(
	store *database.DataStore, user *database.User, parentGroupID, childGroupID int64) service.APIError {
	groupStore := store.Groups()

	var groupData []struct {
		ID   int64  `gorm:"column:ID"`
		Type string `gorm:"column:sType"`
	}

	err := groupStore.OwnedBy(user).
		WithWriteLock().
		Select("groups.ID, sType").
		Where("groups.ID IN(?, ?)", parentGroupID, childGroupID).
		Scan(&groupData).Error()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if len(groupData) < 2 {
		return service.InsufficientAccessRightsError
	}

	for _, groupRow := range groupData {
		if (groupRow.ID == parentGroupID && groupRow.Type == "UserSelf") ||
			(groupRow.ID == childGroupID &&
				map[string]bool{"Root": true, "RootSelf": true, "RootAdmin": true, "UserAdmin": true}[groupRow.Type]) {
			return service.InsufficientAccessRightsError
		}
	}
	return service.NoError
}

type acceptOrRejectRequestsAction string

const (
	acceptRequestsAction acceptOrRejectRequestsAction = "accept"
	rejectRequestsAction acceptOrRejectRequestsAction = "reject"
)

func (srv *Service) acceptOrRejectRequests(w http.ResponseWriter, r *http.Request,
	action acceptOrRejectRequestsAction) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "group_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	if apiErr := checkThatUserOwnsTheGroup(srv.Store, user, parentGroupID); apiErr != service.NoError {
		return apiErr
	}

	var results database.GroupGroupTransitionResults
	if len(groupIDs) > 0 {
		err = srv.Store.InTransaction(func(store *database.DataStore) error {
			results, err = store.GroupGroups().Transition(
				map[acceptOrRejectRequestsAction]database.GroupGroupTransitionAction{
					acceptRequestsAction: database.AdminAcceptsRequest,
					rejectRequestsAction: database.AdminRefusesRequest,
				}[action], parentGroupID, groupIDs, user.UserID)
			return err
		})
	}

	service.MustNotBeError(err)

	renderGroupGroupTransitionResults(w, r, results)
	return service.NoError
}
