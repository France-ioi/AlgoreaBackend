package groups

import (
	"errors"
	"github.com/go-chi/render"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getRecentActivity(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryGetInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryGetInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var count int64
	if err = srv.Store.GroupAncestors().OwnedByUserID(user.UserID).
		Where("idGroupChild = ?", groupID).Count(&count).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	if count == 0 {
		return service.ErrForbidden(errors.New("insufficient access rights"))
	}

	query := srv.Store.UserAnswers().All().WithUsers().WithItems().
		Select(
			`users_answers.ID as ID, users_answers.sSubmissionDate, users_answers.bValidated, users_answers.iScore,
       items.ID AS Item__ID, items.sType AS Item__sType,
		   users.sLogin AS User__sLogin, users.sFirstName AS User__sFirstName, users.sLastName AS User__sLastName,
			 IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS Item__String__sTitle,
       COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS Item__String__idLanguage`).
		Where("users_answers.idItem IN (?)",
			srv.Store.ItemAncestors().All().DescendantsOf(itemID).Select("idItemChild").SubQuery()).
		Where("users_answers.sType='Submission'")
	query = srv.Store.Items().JoinStrings(user, query)
	query = srv.Store.Items().KeepItemsVisibleBy(user, query)
	query = srv.Store.GroupAncestors().KeepUsersThatAreDescendantsOf(groupID, query)
	query = query.Order("users_answers.sSubmissionDate DESC, users_answers.ID")
	query = srv.SetQueryLimit(r, query)
	query = srv.filterByValidated(r, query)

	if query, err = srv.filterByFromSubmissionDateAndFromID(r, query); err != nil {
		return service.ErrInvalidRequest(err)
	}

	var result []map[string]interface{}
	if err := query.ScanIntoSliceOfMaps(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}

func (srv *Service) filterByValidated(r *http.Request, query database.DB) database.DB {
	validated, err := service.ResolveURLQueryGetBoolField(r, "validated")
	if err == nil {
		query = query.Where("users_answers.validated = ?", validated)
	}
	return query
}

func (srv *Service) filterByFromSubmissionDateAndFromID(r *http.Request, query database.DB) (database.DB, error) {
	fromID, fromIDError := service.ResolveURLQueryGetInt64Field(r, "from.id")
	fromSubmissionDate, fromSubmissionDateError := service.ResolveURLQueryGetStringField(r, "from.submission_date")
	if (fromIDError != nil && fromSubmissionDateError == nil) || (fromIDError == nil && fromSubmissionDateError != nil) {
		return nil, errors.New("both from.id and from.submission_date or none of them must be present")
	}
	if fromIDError == nil {
		// include fromSubmissionDate, exclude fromID
		query = query.Where(
			"(users_answers.sSubmissionDate <= ? AND users_answers.ID > ?) OR users_answers.sSubmissionDate < ?",
			fromSubmissionDate, fromID, fromSubmissionDate)
	}
	return query, nil
}
