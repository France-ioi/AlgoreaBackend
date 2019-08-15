package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/group-invitations groups users invitationsView
// ---
// summary: List current invitations and requests to groups
// description:
//   Returns the list of invitations that the current user received and requests sent by him
//   (`groups_groups.sType` is “invitationSent” or “requestSent” or “requestRefused”)
//   with `groups_groups.sStatusDate` within `within_weeks` back from now (if `within_weeks` is present).
// parameters:
// - name: within_weeks
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,id,-id]
// - name: from.status_date
//   description: Start the page from the request/invitation next to one with `sStatusDate` = `from.status_date`
//                and `groups_groups.ID` = `from.id`
//                (`from.id` is required when `from.status_date` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the request/invitation next to one with `sStatusDate`=`from.status_date`
//                and `groups_groups.ID`=`from.id`
//                (`from.status_date` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display the first N requests/invitations
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with an array of invitations/requests
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/invitationsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupInvitations(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.ID,
			groups_groups.sStatusDate,
			groups_groups.sType,
			users.ID AS inviting_user__ID,
			users.sLogin AS inviting_user__sLogin,
			users.sFirstName AS inviting_user__sFirstName,
			users.sLastName AS inviting_user__sLastName,
			groups.ID AS group__ID,
			groups.sName AS group__sName,
			groups.sDescription AS group__sDescription,
			groups.sType AS group__sType`).
		Joins("LEFT JOIN users ON users.ID = groups_groups.idUserInviting").
		Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent").
		Where("groups_groups.sType IN ('invitationSent', 'requestSent', 'requestRefused')").
		Where("groups_groups.idGroupChild = ?", user.SelfGroupID)

	if len(r.URL.Query()["within_weeks"]) > 0 {
		withinWeeks, err := service.ResolveURLQueryGetInt64Field(r, "within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where("NOW() - INTERVAL ? WEEK < groups_groups.sStatusDate", withinWeeks)
	}

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
