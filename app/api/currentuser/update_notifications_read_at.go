package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation PUT /current-user/notifications-read-at users userNotificationReadDateUpdate
//
//	---
//	summary: Update user's notification read date
//	description: Set users.notifications_read_at to NOW() for the current user
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateNotificationsReadAt(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	// the user middleware has already checked that the user exists so we just ignore the case where nothing is updated
	service.MustNotBeError(srv.GetStore(r).Users().ByID(user.GroupID).
		UpdateColumn("notifications_read_at", database.Now()).Error())

	response := service.Response[*struct{}]{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}
