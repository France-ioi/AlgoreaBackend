// Package contests provides API services for contests managing.
package contests

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `contests`
// swagger:ignore
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route contests.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/contests/{item_id}/groups/by-name", service.AppHandler(srv.getGroupByName).ServeHTTP)

	router.Get("/contests/administered", service.AppHandler(srv.getAdministeredList).ServeHTTP)

	router.Get("/contests/{item_id}/groups/{group_id}/additional-times",
		service.AppHandler(srv.getGroupAdditionalTimes).ServeHTTP)
	router.Get("/items/{item_id}/groups/{group_id}/members/additional-times",
		service.AppHandler(srv.getMembersAdditionalTimes).ServeHTTP)
}

// swagger:model itemAdditionalTimesInfo
type itemAdditionalTimesInfo struct {
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

const team = "Team"

func getParticipantTypeForItemWithDurationManagedByUser(
	store *database.DataStore, itemID int64, user *database.User,
) (string, error) {
	var participantType string
	err := store.Items().WithDurationByIDAndManagedByUser(itemID, user).
		PluckFirst("items.entry_participant_type", &participantType).Error()
	return participantType, err
}
