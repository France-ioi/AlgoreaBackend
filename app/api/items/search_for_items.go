package items

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
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
	// enum: Chapter,Task,Skill
	Type string `json:"type"`

	// required: true
	Permissions structures.ItemPermissions `json:"permissions"`
}

type itemSearchResponseRowRaw struct {
	ID    int64
	Title *string
	Type  string

	*database.RawGeneratedPermissionFields
}

// swagger:operation GET /items/search items itemSearch
//
//	---
//	summary: Search for items
//	description: >
//		Searches for visible (`can_view` >= 'info') items, basing on a substring of their titles
//		in the current user's (if exists, otherwise default) language.
//
//
//		All the words of the search query must appear in the title for the item to be returned.
//	parameters:
//		- name: search
//			in: query
//			type: string
//			minLength: 3
//			required: true
//		- name: types_include
//			in: query
//			default: [Chapter,Task,Skill]
//			type: array
//			items:
//				type: string
//				enum: [Chapter,Task,Skill]
//		- name: types_exclude
//			in: query
//			default: []
//			type: array
//			items:
//				type: string
//				enum: [Chapter,Task,Skill]
//		- name: limit
//			description: Display the first N items
//			in: query
//			type: integer
//			maximum: 20
//			default: 20
//	responses:
//		"200":
//			description: OK. Success response with an array of items
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/itemSearchResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
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
	store := srv.GetStore(r)

	typesList, err := service.ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(r, "types",
		map[string]bool{"Chapter": true, "Task": true, "Skill": true})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	query := store.Items().GetSearchQuery(user, searchString, typesList)

	query = service.NewQueryLimiter().
		SetDefaultLimit(20).SetMaxAllowedLimit(20).Apply(r, query)

	var result []itemSearchResponseRowRaw
	service.MustNotBeError(query.Scan(&result).Error())

	convertedResult := make([]itemSearchResponseRow, 0, len(result))
	for i := range result {
		convertedResult = append(convertedResult, itemSearchResponseRow{
			ID:          result[i].ID,
			Title:       result[i].Title,
			Type:        result[i].Type,
			Permissions: *result[i].AsItemPermissions(store.PermissionsGranted()),
		})
	}
	render.Respond(w, r, convertedResult)
	return service.NoError
}
