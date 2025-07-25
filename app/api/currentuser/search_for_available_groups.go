package currentuser

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const minSearchStringLength = 3

// swagger:operation GET /current-user/available-groups groups groupsJoinableSearch
//
//	---
//	summary: Search for groups to join
//	description: >
//		Searches for groups that can be joined freely, based on a substring of their name.
//		Returns groups with `is_public` = 1 and `type` != 'User'/'ContestParticipants', whose `name` has `{search}` as a substring,
//		and for that the current user is not already a member and don’t have pending requests/invitations.
//
//
//		All the words of the search query must appear in the name for the group to be returned.
//
//
//		Note: MySQL Full-Text Search IN BOOLEAN MODE is used for the search, "amazing team" is transformed to "+amazing* +team*",
//		so the words must all appear, as a prefix of a word.
//	parameters:
//		- name: search
//			in: query
//			type: string
//			minLength: 3
//			required: true
//		- name: sort
//			in: query
//			default: [id]
//			type: array
//			items:
//				type: string
//				enum: [id,-id]
//		- name: from.id
//			description: Start the page from the group next to one with `groups.id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display the first N groups
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//				description: OK. Success response with an array of found groups
//				schema:
//					type: array
//					items:
//						type: object
//						properties:
//							id:
//								type: string
//								format: int64
//							name:
//								type: string
//							type:
//								type: string
//								enum: [Class,Team,Club,Friends,Other,Session,Base]
//							description:
//								description: Nullable
//								type: string
//						required:
//							- id
//							- name
//							- type
//							- description
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) searchForAvailableGroups(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	searchString, err := service.ResolveURLQueryGetStringField(httpRequest, "search")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	searchString = strings.TrimSpace(searchString)

	if utf8.RuneCountInString(searchString) < minSearchStringLength {
		return service.ErrInvalidRequest(
			fmt.Errorf("the search string should be at least %d characters long", minSearchStringLength))
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	query := store.Groups().AvailableGroupsBySearchString(user, searchString)

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields:       service.SortingAndPagingFields{"id": {ColumnName: "groups.id"}},
			DefaultRules: "id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(responseWriter, httpRequest, convertedResult)
	return nil
}
