package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const maxNumberOfIDsInGroupPath = 10

// swagger:model groupBreadcrumbsViewResponseRow
type groupBreadcrumbsViewResponseRow struct {
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
	// required:true
	// enum: Class,Team,Club,Friends,Other,User,Session,Base
	Type string `json:"type"`
}

// swagger:operation GET /groups/{ids}/breadcrumbs group-memberships groupBreadcrumbsView
//
//	---
//	summary: Get group breadcrumbs
//	description: >
//
//		Returns brief information for groups listed in `ids`.
//
//
//		Each group must be visible to the current user, so it should be either
//
//			1. an ancestor of a group the current user joined, or
//			2. an ancestor of a non-user group he manages, or
//			3. a descendant of a group he manages, or
//			4. a group with is_public=1,
//
//		otherwise the 'forbidden' error is returned. Also, there must be no duplicates in the list.
//
//	parameters:
//		- name: ids
//			in: path
//			type: string
//			description: slash-separated list of IDs (no more than 10 IDs allowed)
//			required: true
//	responses:
//		"200":
//			description: OK. Success response with an array of group information
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/groupBreadcrumbsViewResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getBreadcrumbs(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	ids, err := service.ResolveURLQueryPathInt64SliceFieldWithLimit(httpRequest, "ids", maxNumberOfIDsInGroupPath)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	idsInterface := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idsInterface = append(idsInterface, id)
	}
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	var result []groupBreadcrumbsViewResponseRow
	err = store.Groups().PickVisibleGroups(store.Groups().Where("id IN(?)", ids), user).
		Select("id, name, type").
		Order(gorm.Expr("FIELD(id"+strings.Repeat(", ?", len(idsInterface))+")", idsInterface...)).
		Scan(&result).Error()

	if gorm.IsRecordNotFoundError(err) || len(result) != len(ids) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, result)
	return nil
}
