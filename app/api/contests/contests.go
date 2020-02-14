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

	router.Get("/contests/{item_id}/groups/by-name", service.AppHandler(srv.getGroupByName).ServeHTTP)

	router.Get("/contests/administered", service.AppHandler(srv.getAdministeredList).ServeHTTP)

	router.Put("/contests/{item_id}/groups/{group_id}/additional-times",
		service.AppHandler(srv.setAdditionalTime).ServeHTTP)
	router.Get("/contests/{item_id}/groups/{group_id}/members/additional-times",
		service.AppHandler(srv.getMembersAdditionalTimes).ServeHTTP)
	router.Get("/contests/{item_id}/qualification-state",
		service.AppHandler(srv.getQualificationState).ServeHTTP)
	router.Post("/contests/{item_id}/enter", service.AppHandler(srv.enter).ServeHTTP)
}

// swagger:model contestInfo
type contestInfo struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	Name string `json:"name"`
	// required: true
	Type string `json:"type"`
	// required: true
	AdditionalTime int32 `json:"additional_time"`
	// required: true
	TotalAdditionalTime int32 `json:"total_additional_time"`
}

func (srv *Service) isTeamOnlyContestManagedByUser(itemID int64, user *database.User) (bool, error) {
	var isTeamOnly bool
	err := srv.Store.Items().ContestManagedByUser(itemID, user).
		PluckFirst("IFNULL(items.entry_participant_type = 'Team', 0)", &isTeamOnly).Error()
	return isTeamOnly, err
}

type qualificationState string

const (
	alreadyStarted qualificationState = "already_started"
	notReady       qualificationState = "not_ready"
	ready          qualificationState = "ready"
)
