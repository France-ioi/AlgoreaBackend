package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := srv.checkThatUserOwnsTheGroup(user, groupID); apiError != service.NoError {
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

	if len(r.URL.Query()["old_rejections_weeks"]) > 0 {
		oldRejectionsWeeks, err := service.ResolveURLQueryGetInt64Field(r, "old_rejections_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where(`
			groups_groups.sType IN ('invitationSent', 'requestSent') OR
			groups_groups.sStatusDate IS NULL OR
			DATE_SUB(NOW(), INTERVAL ? WEEK) < groups_groups.sStatusDate`, oldRejectionsWeeks)
	}

	query = service.SetQueryLimit(r, query)
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
