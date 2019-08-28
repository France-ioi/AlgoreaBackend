package currentuser

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/teams/by-item/{item_id} users groups teamGetByItemID
// ---
// summary: Get current user's team for item
// description: >
//   The team identified by `item_id` i.e. a group which:
//
//     * has `idTeamItem` equal to the input `item_id`,
//
//     * is a direct parent (i.e. via `groups_groups` with `sType` = "invitationAccepted"/"requestAccepted"/"joinedByCode")
//       of the authenticated user’s `selfGroup`,
//
//     * is of type "Team".
//
//
//   If there are several matching teams, returns the first one (in the `group_id` order) is returned.
// parameters:
// - name: item_id
//   type: integer
//   required: true
//   in: path
// responses:
//   "200":
//     description: OK. Success response with the team's ID
//     schema:
//       type: object
//       properties:
//         group_id:
//           type: integer
//       required: [group_id]
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getTeamByItem(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var teamID int64
	user := srv.GetUser(r)
	err = srv.Store.Groups().TeamGroupByTeamItemAndUser(itemID, user).PluckFirst("groups.ID", &teamID).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrNotFound(errors.New("no team for this item"))
	}
	service.MustNotBeError(err)

	render.Respond(w, r, &map[string]string{"group_id": strconv.FormatInt(teamID, 10)})
	return service.NoError
}
