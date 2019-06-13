package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getGroupMembershipsHistory(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	err := user.Load()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	selfGroupID, _ := user.SelfGroupID()
	notificationReadDate, _ := user.NotificationReadDate()
	query := srv.Store.GroupGroups().
		Select(`
			groups_groups.ID,
			groups_groups.sStatusDate,
			groups_groups.sType,
			groups.sName AS group__sName,
			groups.sType AS group__sType`).
		Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent").
		Where("groups_groups.sType != 'direct'").
		Where("groups_groups.idGroupChild = ?", selfGroupID)
	if notificationReadDate != nil {
		query = query.Where("groups_groups.sStatusDate >= ?", notificationReadDate)
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
