package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getMembers(w http.ResponseWriter, r *http.Request) service.APIError {
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
			users.ID AS user__ID,
			users.sLogin AS user__sLogin,
			users.sFirstName AS user__sFirstName,
			users.sLastName AS user__sLastName,
			users.iGrade AS user__iGrade`).
		Joins("LEFT JOIN users ON users.idGroupSelf = groups_groups.idGroupChild").
		Where("groups_groups.sType IN ('invitationAccepted', 'requestAccepted', 'direct')").
		Where("groups_groups.idGroupParent = ?", groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"user.login":  {ColumnName: "users.sLogin"},
			"user.grade":  {ColumnName: "users.iGrade"},
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
