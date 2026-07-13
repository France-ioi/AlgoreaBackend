package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model itemOwnersResponseRow
type itemOwnersResponseRow struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Name string `json:"name"`
	// required: true
	Type string `json:"type"`
}

// swagger:operation GET /items/{item_id}/owners items itemOwnersList
//
//	---
//	summary: List owner groups for an item
//	description: >
//	 Lists the groups having `is_owner_generated` = true in `permissions_generated` for the given item.
//
//
//	 Restrictions:
//		 * the current user should have at least `can_edit` = 'all' permission on the item,
//
//	 otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: from.id
//			description: Start the page from the group next to the group with `id` = `{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [name,id]
//			type: array
//			items:
//				type: string
//				enum: [name,-name,id,-id]
//		- name: limit
//			description: Display first N owner groups
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with an array of owner groups
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/itemOwnersResponseRow"
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
func (srv *Service) listOwners(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)
	found, err := store.Permissions().MatchingUserAncestors(user).
		Where("item_id = ?", itemID).
		WherePermissionIsAtLeast("edit", "all").
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	query := store.Groups().
		Joins("JOIN permissions_generated ON permissions_generated.group_id = groups.id").
		Where("permissions_generated.item_id = ?", itemID).
		Where("permissions_generated.is_owner_generated = 1").
		Select("groups.id, groups.name, groups.type")

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err = service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"name": {ColumnName: "groups.name"},
				"id":   {ColumnName: "groups.id"},
			},
			DefaultRules: "name,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	result := make([]itemOwnersResponseRow, 0)
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(responseWriter, httpRequest, result)
	return nil
}
