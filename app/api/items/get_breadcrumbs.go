package items

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /items/{ids}/breadcrumbs items itemsBreadCrumbsData
// ---
// summary: Get breadcrumbs
// description: >
//
//   Returns titles for items listed in `ids` in the user's preferred language (if exist) or the items'
//   default language.
//
//
//   Restrictions:
//     * the list of item IDs should be a valid path from a root item (`type`=’Root’), otherwise the 'bad request'
//       error is returned)
//     * the user should have partial or full access for each listed item except the last one through that path,
//       and at least gray access for the last item, otherwise the 'forbidden' error is returned.
// parameters:
// - name: ids
//   in: path
//   type: string
//   description: slash-separated list of IDs
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
	valid, err := srv.Store.Items().ValidateUserAccess(user, ids)
	service.MustNotBeError(err)
	if !valid {
		return service.ErrForbidden(errors.New("insufficient access rights on given item ids"))
	}

	// Validate the hierarchy
	valid, err = srv.Store.Items().IsValidHierarchy(ids)
	service.MustNotBeError(err)
	if !valid {
		return service.ErrInvalidRequest(errors.New("the IDs chain is corrupt"))
	}

	idsInterface := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsInterface = append(idsInterface, id)
	}
	var result []map[string]interface{}
	service.MustNotBeError(srv.Store.Items().Select(`
			items.id AS item_id,
			COALESCE(user_strings.title, default_strings.title) AS title,
			COALESCE(user_strings.language_id, default_strings.language_id) AS language_id`).
		JoinsUserAndDefaultItemStrings(user).
		Where("items.id IN (?)", ids).
		Order(gorm.Expr("FIELD(items.id"+strings.Repeat(", ?", len(idsInterface))+")", idsInterface...)).
		ScanIntoSliceOfMaps(&result).Error())

	render.Respond(w, r, service.ConvertSliceOfMapsFromDBToJSON(result))
	return service.NoError
}

func idsFromRequest(r *http.Request) ([]int64, error) {
	ids, err := service.ResolveURLQueryPathInt64SliceField(r, "ids")
	if err != nil {
		return nil, err
	}
	if len(ids) > 10 {
		return nil, errors.New("no more than 10 ids expected")
	}
	return ids, nil
}
