package contests

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /contests/{item_id}/group-by-name contests groups contestGetGroupByName
// ---
// summary: Get a group by name
// description: >
//                Return one group matching the name and satisfying:
//
//                  * the authenticated user has access to the contest (grayed, partial or full);
//                  * the authenticated user is an owner of the group;
//                  * the `groups.sName` (matching `sLogin` if a "UserSelf" group) is matching exactly the input `name` parameter.
//
//                If there are several groups or users matching, return the first one (by `ID`).
//
//
//                This service is only here for transition between the former interface and the new one.
//                This way of searching only by sName/sLogin and getting one result is not really convenient,
//                but matching the former UI. This service will have to be removed as soon as
//                the new interface is used.
//
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   required: true
// - name: name
//   in: query
//   type: string
//   required: true
// responses:
//   "200":
//     description: OK. Success response with the group_id
//     schema:
//       type: object
//       properties:
//         group_id:
//           type: integer
//           description: The group ID
//       required:
//         - group_id
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"

func (srv *Service) getGroupByName(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupName, err := service.ResolveURLQueryGetStringField(r, "name")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	hasAccessToItem, err := srv.Store.Items().VisibleByID(user, itemID).HasRows()
	service.MustNotBeError(err)
	if !hasAccessToItem {
		return service.InsufficientAccessRightsError
	}

	var groupID int64
	if err = srv.Store.Groups().OwnedBy(user).Where("BINARY sName = ?", groupName).
		Order("groups.ID").PluckFirst("groups.ID", &groupID).Error(); gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, map[string]interface{}{"group_id": strconv.FormatInt(groupID, 10)})
	return service.NoError
}
