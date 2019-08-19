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
// swagger:ignore
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route contests
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))

	router.Get("/contests/{item_id}/group-by-name", service.AppHandler(srv.getGroupByName).ServeHTTP)

	router.Put("/contests/{item_id}/groups/{group_id}/additional-times",
		service.AppHandler(srv.setAdditionalTime).ServeHTTP)
	router.Get("/contests/{item_id}/groups/{group_id}/members/additional-times",
		service.AppHandler(srv.getMembersAdditionalTimes).ServeHTTP)
}

// swagger:model contestInfo
type contestInfo struct {
	// required: true
	GroupID int64 `gorm:"column:idGroup" json:"group_id,string"`
	// required: true
	Name string `gorm:"column:sName" json:"name"`
	// required: true
	Type string `gorm:"column:sType" json:"type"`
	// required: true
	AdditionalTime int32 `gorm:"column:iAdditionalTime" json:"additional_time"`
	// required: true
	TotalAdditionalTime int32 `gorm:"column:iTotalAdditionalTime" json:"total_additional_time"`
}

func (srv *Service) getTeamModeForTimedContestManagedByUser(itemID int64, user *database.User) (bool, error) {
	var isTeamOnly bool
	err := srv.Store.Items().ByID(itemID).Where("items.sDuration IS NOT NULL").
		Joins("JOIN groups_items ON groups_items.idItem = items.ID").
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupAncestor = groups_items.idGroup AND
				groups_ancestors.idGroupChild = ?`, user.SelfGroupID).
		Group("items.ID").
		Having("MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedAccessSolutionsDate) <= NOW()").
		PluckFirst("items.bHasAttempts", &isTeamOnly).Error()
	return isTeamOnly, err
}
