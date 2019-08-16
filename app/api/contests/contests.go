// Package contests provides API services for contests managing
package contests

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `contests`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route contests
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))

	router.Get("/contests/administered", service.AppHandler(srv.getAdministeredList).ServeHTTP)
	router.Get("/contests/{item_id}/group-by-name", service.AppHandler(srv.getGroupByName).ServeHTTP)
}

func (srv *Service) checkThatUserCanManageTimedContest(itemID int64, user *database.User) service.APIError {
	ok, err := srv.Store.Items().ByID(itemID).Where("items.sDuration IS NOT NULL").
		Joins("JOIN groups_items ON groups_items.idItem = items.ID").
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupAncestor = groups_items.idGroup AND
				groups_ancestors.idGroupChild = ?`, user.SelfGroupID).
		Group("items.ID").
		Having("MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedAccessSolutionsDate) <= NOW()").
		HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}
