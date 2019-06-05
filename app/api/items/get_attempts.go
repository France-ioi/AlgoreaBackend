package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getAttempts(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	user := srv.GetUser(r)
	query := srv.Store.GroupAttempts().ByUserAndItemID(user, itemID).
		Joins("LEFT JOIN users AS creators ON creators.ID = groups_attempts.idUserCreator").
		Select(`
			groups_attempts.ID, groups_attempts.iOrder, groups_attempts.iScore, groups_attempts.bValidated,
			groups_attempts.sStartDate, creators.sLogin AS userCreator__sLogin,
			creators.sFirstName AS userCreator__sFirstName, creators.sLastName AS userCreator__sLastName`)
	query = service.SetQueryLimit(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"order": {ColumnName: "groups_attempts.iOrder", FieldType: "int64"},
	}, "order")
	if apiError != service.NoError {
		return apiError
	}
	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}
