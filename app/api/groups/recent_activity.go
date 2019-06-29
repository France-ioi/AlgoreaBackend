package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getRecentActivity(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryGetInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemDescendants := srv.Store.ItemAncestors().DescendantsOf(itemID).Select("idItemChild")
	service.MustNotBeError(itemDescendants.Error())
	query := srv.Store.UserAnswers().WithUsers().WithItems().
		Select(
			`users_answers.ID as ID, users_answers.sSubmissionDate, users_answers.bValidated, users_answers.iScore,
       items.ID AS Item__ID, items.sType AS Item__sType,
		   users.sLogin AS User__sLogin, users.sFirstName AS User__sFirstName, users.sLastName AS User__sLastName,
			 IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS Item__String__sTitle`).
		JoinsUserAndDefaultItemStrings(user).
		Where("users_answers.idItem IN ?", itemDescendants.SubQuery()).
		Where("users_answers.sType='Submission'").
		WhereItemsAreVisible(user).
		WhereUsersAreDescendantsOfGroup(groupID)

	query = service.NewQueryLimiter().Apply(r, query)
	query = srv.filterByValidated(r, query)

	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"submission_date": {ColumnName: "users_answers.sSubmissionDate", FieldType: "time"},
			"id":              {ColumnName: "users_answers.ID", FieldType: "int64"}},
		"-submission_date")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	if err := query.ScanIntoSliceOfMaps(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}

func (srv *Service) filterByValidated(r *http.Request, query *database.DB) *database.DB {
	validated, err := service.ResolveURLQueryGetBoolField(r, "validated")
	if err == nil {
		query = query.Where("users_answers.bValidated = ?", validated)
	}
	return query
}
