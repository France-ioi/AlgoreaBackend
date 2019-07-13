package currentuser

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation PUT /current-user/notification-read-date userNotificationReadDateUpdate
// ---
// summary: Update user's notification read date
// description: Set users.sNotificationReadDate to NOW() for the current user
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateNotificationReadDate(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	service.MustNotBeError(srv.Store.Users().ByID(user.ID).
		UpdateColumn("sNotificationReadDate", gorm.Expr("NOW()")).Error())

	response := service.Response{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}
