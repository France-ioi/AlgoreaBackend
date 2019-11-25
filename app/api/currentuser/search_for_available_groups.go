package currentuser

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

const minSearchStringLength = 3

// swagger:operation GET /current-user/available-groups groups users groupsAvailableSearch
// ---
// summary: Search for available groups
// description: >
//   Searches for groups that can be joined freely, based on a substring of their name.
//   Returns groups with `free_access`=1, whose `name` has `search` as a substring, and for that the current user
//   is not already a member and don’t have pending requests/invitations.
//
//
//   Note: The current implementation may be very slow because it uses `LIKE` with a percentage wildcard
//   at the beginning. This causes MySQL to explore every row having `free_access`=1. Moreover, actually
//   it has to examine every row of the `groups` table since there is no index for the `free_access` column.
//   But since there are not too many groups and the result rows count is limited, the search works almost well.
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
//   description: Start the page from the group next to one with `groups.id`=`from.id`
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
//             enum: [Class,Team,Club,Friends,Other,UserSelf,UserAdmin,Base]
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
func (srv *Service) searchForAvailableGroups(w http.ResponseWriter, r *http.Request) service.APIError {
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

	skipGroups := srv.Store.GroupGroups().
		Select("groups_groups.parent_group_id").
		Where("groups_groups.child_group_id = ?", user.GroupID).
		Where("groups_groups.type IN ('requestSent', 'invitationSent', 'requestAccepted', 'invitationAccepted', 'direct', 'joinedByCode')").
		SubQuery()

	escapedSearchString := escapeLikeString(searchString, '|')
	query := srv.Store.Groups().
		Select(`
			groups.id,
			groups.name,
			groups.type,
			groups.description`).
		Where("groups.free_access").
		Where("groups.id NOT IN ?", skipGroups).
		Where("groups.name LIKE CONCAT('%', ?, '%') ESCAPE '|'", escapedSearchString)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"id": {ColumnName: "groups.id", FieldType: "int64"}},
		"id", false)
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}

// escapeLikeStringBackslash escapes string with backslashes the given escape character.
// This escapes the contents of a string (provided as string)
// by adding the escape character before percent signs (%), and underscore signs (_).
func escapeLikeString(v string, escapeCharacter byte) string {
	pos := 0
	buf := make([]byte, len(v)*3)

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case escapeCharacter:
			buf[pos] = escapeCharacter
			buf[pos+1] = escapeCharacter
			pos += 2
		case '%':
			buf[pos] = escapeCharacter
			buf[pos+1] = '%'
			pos += 2
		case '_':
			buf[pos] = escapeCharacter
			buf[pos+1] = '_'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	result := buf[:pos]
	return *(*string)(unsafe.Pointer(&result)) // nolint:gosec
}
