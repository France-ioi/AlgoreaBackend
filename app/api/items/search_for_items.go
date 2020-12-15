package items

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

// swagger:model itemSearchResponseRow
type itemSearchResponseRow struct {
	// required:true
	ID int64 `json:"id,string"`
	// Title (in current user's language); Nullable
	// required:true
	Title *string `json:"title"`
	// required:true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type"`
}

// swagger:operation GET /items/search items itemSearch
// ---
// summary: Search for items
// description: >
//   Searches for visible (`can_view` >= 'info') items, basing on a substring of their titles
//   in the current user's (if exists, otherwise default) language.
// parameters:
// - name: search
//   in: query
//   type: string
//   minLength: 3
//   required: true
// - name: types_include
//   in: query
//   default: [Chapter,Task,Course,Skill]
//   type: array
//   items:
//     type: string
//     enum: [Chapter,Task,Course,Skill]
// - name: types_exclude
//   in: query
//   default: []
//   type: array
//   items:
//     type: string
//     enum: [Chapter,Task,Course,Skill]
// - name: limit
//   description: Display the first N items
//   in: query
//   type: integer
//   maximum: 20
//   default: 20
// responses:
//   "200":
//     description: OK. Success response with an array of items
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/itemSearchResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) searchForItems(w http.ResponseWriter, r *http.Request) service.APIError {
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

	typesList, err := service.ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(r, "types",
		map[string]bool{"Chapter": true, "Task": true, "Course": true, "Skill": true})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	escapedSearchString := database.EscapeLikeString(searchString, '|')
	query := srv.Store.Items().Visible(user.GroupID).JoinsUserAndDefaultItemStrings(user).
		Select(`
			items.id,
			COALESCE(user_strings.title, default_strings.title) AS title,
			items.type`).
		Where("items.type IN (?)", typesList).
		Where("COALESCE(user_strings.title, default_strings.title) LIKE CONCAT('%', ?, '%') ESCAPE '|'", escapedSearchString).
		Order("items.id")

	query = service.NewQueryLimiter().
		SetDefaultLimit(20).SetMaxAllowedLimit(20).Apply(r, query)

	var result []itemSearchResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
