package currentuser

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

const minSearchStringLength = 3

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
	selfGroupID, err := user.SelfGroupID()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	skipGroups := srv.Store.GroupGroups().
		Select("groups_groups.idGroupParent").
		Where("groups_groups.idGroupChild = ?", selfGroupID).
		Where("groups_groups.sType IN ('requestSent', 'invitationSent', 'requestAccepted', 'invitationAccepted', 'direct')").
		SubQuery()

	escapedSearchString := escapeLikeString(searchString, '|')
	query := srv.Store.Groups().
		Select(`
			groups.ID,
			groups.sName,
			groups.sType,
			groups.sDescription`).
		Where("groups.bFreeAccess").
		Where("groups.ID NOT IN ?", skipGroups).
		Where("groups.sName LIKE CONCAT('%', ?, '%') ESCAPE '|'", escapedSearchString)

	query = service.SetQueryLimit(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"id": {ColumnName: "groups.ID", FieldType: "int64"}},
		"id")
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
