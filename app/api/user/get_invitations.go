package user

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getInvitations(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	err := user.Load()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	selfGroupID, _ := user.SelfGroupID()
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
		Where("groups_groups.idGroupChild = ?", selfGroupID)

	if len(r.URL.Query()["within_weeks"]) > 0 {
		withinWeeks, err := service.ResolveURLQueryGetInt64Field(r, "within_weeks")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		query = query.Where("NOW() - INTERVAL ? WEEK < groups_groups.sStatusDate", withinWeeks)
	}

	query = service.SetQueryLimit(r, query)
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
