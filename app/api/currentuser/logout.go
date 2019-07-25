package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user users auth userLogout
// ---
// summary: Sign the current user out
// description: Removes the current user’s session (all access and refresh tokens)
// responses:
//   "200":
//     "$ref": "#/responses/deletedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) logout(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.Sessions().Delete("idUser = ?", user.ID).Error())
		service.MustNotBeError(store.RefreshTokens().Delete("idUser = ?", user.ID).Error())
		return nil
	}))

	render.Respond(w, r, service.DeletionSuccess(nil))

	return service.NoError
}
