package items

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /items items itemsBreadCrumbsData
// ---
// summary: Get breadcrumbs
// description: >
//
//   Returns titles for items listed in `ids` in the user's preferred language (if exist) or the items'
//   default language.
//
//
//   Restrictions:
//     * the list of item IDs should be a valid path from a root item (`sType`=’Root’), otherwise the 'bad request'
//       error is returned)
//     * the user should have partial or full access for each listed item except the last one through that path,
//       and at least gray access for the last item, otherwise the 'forbidden' error is returned.
// parameters:
// - name: ids
//   in: query
//   type: array
//   items:
//     type: integer
//     format: int64
//   required: true
// responses:
//   "200":
//     description: OK. Breadcrumbs data
//     schema:
//       type: array
//       items:
//         type: object
//         properties:
//           item_id:
//             type: string
//             format: int64
//           title:
//             type: string
//           language_id:
//             type: string
//             format: int64
//         required: [item_id, title, language_id]
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getBreadcrumbs(w http.ResponseWriter, r *http.Request) service.APIError {
	// Get IDs from request and validate it.
	ids, err := idsFromRequest(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	// Validate that the user can see the item IDs.
	user := srv.GetUser(r)
	if valid, err := srv.Store.Items().ValidateUserAccess(user, ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrForbidden(errors.New("insufficient access rights on given item ids"))
	}

	// Validate the hierarchy
	if valid, err := srv.Store.Items().IsValidHierarchy(ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrInvalidRequest(errors.New("the IDs chain is corrupt"))
	}

	idsInterface := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsInterface = append(idsInterface, id)
	}
	var result []map[string]interface{}
	service.MustNotBeError(srv.Store.Items().Select(`
			items.ID AS idItem,
			COALESCE(user_strings.sTitle, default_strings.sTitle) AS sTitle,
			COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS idLanguage`).
		JoinsUserAndDefaultItemStrings(user).
		Where("items.ID IN (?)", ids).
		Order(gorm.Expr("FIELD(items.ID"+strings.Repeat(", ?", len(idsInterface))+")", idsInterface...)).
		ScanIntoSliceOfMaps(&result).Error())

	render.Respond(w, r, service.ConvertSliceOfMapsFromDBToJSON(result))
	return service.NoError
}

func idsFromRequest(r *http.Request) ([]int64, error) {
	ids, err := service.ResolveURLQueryGetInt64SliceField(r, "ids")
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("no ids given")
	}
	if len(ids) > 10 {
		return nil, errors.New("no more than 10 ids expected")
	}
	return ids, nil
}
