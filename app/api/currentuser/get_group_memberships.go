package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-memberships groups users membershipsView
// ---
// summary: List groups that the current user has joined
// description:
//   Returns the list of groups memberships of the current user
//   (`groups_groups.sType` is “requestAccepted”, “invitationAccepted” or “direct”).
// parameters:
// - name: sort
//   in: query
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,id,-id]
// - name: from.status_date
//   description: Start the page from the membership next to one with `sStatusDate` = `from.status_date`
//                and `groups_groups.ID` = `from.id`
//                (`from.id` is required when `from.status_date` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the membership next to one with `sStatusDate`=`from.status_date`
//                and `groups_groups.ID`=`from.id`
//                (`from.status_date` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N memberships
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of groups memberships
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/membershipsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupMemberships(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.ID,
			groups_groups.sStatusDate,
			groups_groups.sType,
			groups.ID AS group__ID,
			groups.sName AS group__sName,
			groups.sDescription AS group__sDescription,
			groups.sType AS group__sType`).
		Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent").
		Where("groups_groups.sType IN ('invitationAccepted', 'requestAccepted', 'direct')").
		Where("groups_groups.idGroupChild = ?", user.SelfGroupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"status_date": {ColumnName: "groups_groups.sStatusDate", FieldType: "time"},
			"id":          {ColumnName: "groups_groups.ID", FieldType: "int64"}},
		"-status_date")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
