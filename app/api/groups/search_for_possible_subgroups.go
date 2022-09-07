package groups

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

const minSearchStringLength = 3

// swagger:operation GET /groups/possible-subgroups groups groupsPossibleSubgroupsSearch
// ---
// summary: Search for possible subgroups
// description: >
//   Searches for groups that can be added as subgroups, based on a substring of their name.
//   Returns groups for which the user is a manager with `can_manage` = 'memberships_and_group',
//   whose `name` has `{search}` as a substring.
// parameters:
// - name: search
//   in: query
//   type: string
//   minLength: 3
//   required: true
// - name: sort
//   in: query
//   default: [id]
//   type: array
//   items:
//     type: string
//     enum: [id,-id]
// - name: from.id
//   description: Start the page from the group next to one with `groups.id`=`{from.id}`
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N groups
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of found groups
//     schema:
//       type: array
//       items:
//         type: object
//         properties:
//           id:
//             type: string
//             format: int64
//           name:
//             type: string
//           type:
//             type: string
//             enum: [Class,Team,Club,Friends,Other,Session,Base]
//           description:
//             description: Nullable
//             type: string
//         required:
//           - id
//           - name
//           - type
//           - description
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) searchForPossibleSubgroups(w http.ResponseWriter, r *http.Request) service.APIError {
	searchString, err := service.ResolveURLQueryGetStringField(r, "search")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	searchString = strings.TrimSpace(searchString)

	if utf8.RuneCountInString(searchString) < minSearchStringLength {
		return service.ErrInvalidRequest(
			fmt.Errorf("the search string should be at least %d characters long", minSearchStringLength))
	}

	user := srv.GetUser(r)

	escapedSearchString := database.EscapeLikeString(searchString, '|')
	query := srv.GetStore(r).Groups().ManagedBy(user).
		Where("group_managers.can_manage = 'memberships_and_group'").
		Group("groups.id").
		Where("groups.type != 'User'").
		Select(`
			groups.id,
			groups.name,
			groups.type,
			groups.description`).
		Where("groups.name LIKE CONCAT('%', ?, '%') ESCAPE '|'", escapedSearchString)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields:       service.SortingAndPagingFields{"id": {ColumnName: "groups.id"}},
			DefaultRules: "id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
