package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/requests groups users groupRequestsView
// ---
// summary: List pending requests and invitations for a group
// description: >
//
//   Returns a list of group requests and invitations
//   (rows from the `groups_groups` table with `idGroupParent` = `group_id` and
//   `sType` = "invitationSent"/"requestSent"/"invitationRefused"/"requestRefused")
//   with basic info on joining (invited/requesting) users and inviting users.
//
//
//   When `old_rejections_weeks` is given, only those rejected invitations/requests
//   (`groups_groups.sType` is "invitationRefused" or "requestRefused") are shown
//   whose `sStatusDate` has changed in the last `old_rejections_weeks` weeks.
//   Otherwise all rejected invitations/requests are shown.
//
//
//   Invited user’s `sFirstName` and `sLastName` are nulls
//   if `groups_groups.sType` = "invitationSent" or "invitationRefused".
//
//
//   The authenticated user should be an owner of `group_id`, otherwise the 'forbidden' error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: old_rejections_weeks
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [-status_date,id]
//   type: array
//   items:
//     type: string
//     enum: [status_date,-status_date,joining_user.login,-joining_user.login,type,-type,id,-id]
// - name: from.status_date
//   description: Start the page from the request/invitation next to the request/invitation with
//                `groups_groups.sStatusDate` = `from.status_date`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.joining_user.login
//   description: Start the page from the request/invitation next to the request/invitation
//                whose joining user's login is `from.joining_user.login`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.type
//   description: Start the page from the request/invitation next to the request/invitation with
//                `groups_groups.sType` = `from.type`, sorted numerically.
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the request/invitation next to the request/invitation with `groups_groups.ID`=`from.id`
//                (depending on the `sort` parameter, some other `from.*` parameters may be required)
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
//     description: OK. The array of group requests/invitations
//     schema:
//       type: array
//       items:
//         type: object
//         required: [id, status_date, type, joining_user, inviting_user]
//         properties:
//           id:
//             description: "`groups_groups.ID`"
//             type: string
//             format: int64
//           status_date:
//             type: string
//             description: Nullable
//             format: date-time
//           type:
//             type: string
//             description: "`groups_groups.sType`"
//             enum: [invitationSent, requestSent, invitationRefused, requestRefused]
//           joining_user:
//             type: object
//             description: Nullable
//             required: [id, login, first_name, last_name, grade]
//             properties:
//               id:
//                 description: "`users.ID`"
//                 type: string
//                 format: int64
//               login:
//                 type: string
//               first_name:
//                 description: Nullable
//                 type: string
//               last_name:
//                 description: Nullable
//                 type: string
//               grade:
//                 description: Nullable
//                 type: integer
//           inviting_user:
//             type: object
//             description: Nullable
//             required: [id, login, first_name, last_name]
//             properties:
//               id:
//                 description: "`users.ID`"
//                 type: string
//                 format: int64
//               login:
//                 type: string
//               first_name:
//                 description: Nullable
//                 type: string
//               last_name:
//                 description: Nullable
//                 type: string
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.ID,
			groups_groups.sStatusDate,
			groups_groups.sType,
			joining_user.ID AS joiningUser__ID,
			joining_user.sLogin AS joiningUser__sLogin,
			IF(groups_groups.sType IN ('invitationSent', 'invitationRefused'), NULL, joining_user.sFirstName) AS joiningUser__sFirstName,
			IF(groups_groups.sType IN ('invitationSent', 'invitationRefused'), NULL, joining_user.sLastName) AS joiningUser__sLastName,
			joining_user.iGrade AS joiningUser__iGrade,
			inviting_user.ID AS invitingUser__ID,
			inviting_user.sLogin AS invitingUser__sLogin,
			inviting_user.sFirstName AS invitingUser__sFirstName,
			inviting_user.sLastName AS invitingUser__sLastName`).
		Joins("LEFT JOIN users AS inviting_user ON inviting_user.ID = groups_groups.idUserInviting").
		Joins("LEFT JOIN users AS joining_user ON joining_user.idGroupSelf = groups_groups.idGroupChild").
		Where("groups_groups.sType IN ('invitationSent', 'requestSent', 'invitationRefused', 'requestRefused')").
		Where("groups_groups.idGroupParent = ?", groupID)

	if len(r.URL.Query()["rejections_within_weeks"]) > 0 {
		oldRejectionsWeeks, err := service.ResolveURLQueryGetInt64Field(r, "rejections_within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where(`
			groups_groups.sType IN ('invitationSent', 'requestSent') OR
			NOW() - INTERVAL ? WEEK < groups_groups.sStatusDate`, oldRejectionsWeeks)
	}

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"type":               {ColumnName: "groups_groups.sType"},
			"joining_user.login": {ColumnName: "joining_user.sLogin"},
			"status_date":        {ColumnName: "groups_groups.sStatusDate", FieldType: "time"},
			"id":                 {ColumnName: "groups_groups.ID", FieldType: "int64"}},
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
